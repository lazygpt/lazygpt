//

package memory

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/dgraph-io/badger/v4"

	// NOTE(jkoelker) This is a hack to get embedings in the plugin.
	//                We need to formailize how plugins would be able to
	//                call other plugins.
	"github.com/lazygpt/lazygpt/pkg/plugin"
	"github.com/lazygpt/lazygpt/plugin/api"
	"github.com/lazygpt/lazygpt/plugin/log"
)

const (
	GarbageCollectionDiscardRatio = 0.5
	GarbageCollectionInterval     = 15 * time.Minute
)

// Local is a local memory system that uses a badger database to store
// the data.
type Local struct {
	DB      *badger.DB
	DataDir string

	closing   chan struct{}
	embedding api.Embedding
	gcStopped chan struct{}
	manager   *plugin.Manager
	logger    *log.Logger
}

var _ api.Memory = (*Local)(nil)

// NewLocal creates a new Local memory instance backed by a badger
// database.
func NewLocal(datadir string) *Local {
	return &Local{
		DataDir: datadir,
	}
}

func (local *Local) SetupLogger(ctx context.Context) {
	if local.logger == nil {
		local.logger = log.FromContext(ctx).WithName("memory")
	}
}

// Open opens the database and sets the logger and starts the garbage
// collector.
func (local *Local) Open(ctx context.Context) error {
	local.SetupLogger(ctx)

	local.logger.Info("Starting local plugin")

	options := badger.DefaultOptions(filepath.Join(local.DataDir, "memorydb"))
	options = options.WithLogger(NewLogger(local.logger.WithName("badger")))

	database, err := badger.Open(options)
	if err != nil {
		return fmt.Errorf("failed to open local database: %w", err)
	}

	local.DB = database
	local.closing = make(chan struct{})
	local.gcStopped = make(chan struct{})
	local.manager = plugin.NewManager()

	// NOTE(jkoelker) Load the embedding plugin. This is a hack and is
	//                hardcoded to openai. We need to formalize how plugins
	//                can call other plugins.
	client, err := local.manager.Client(ctx, "openai")
	if err != nil {
		return fmt.Errorf("failed to load embedding plugin: %w", err)
	}

	protocol, err := client.Client()
	if err != nil {
		return fmt.Errorf("failed to load embedding protocol: %w", err)
	}

	raw, err := protocol.Dispense("embedding")
	if err != nil {
		return fmt.Errorf("failed to dispense embedding: %w", err)
	}

	embedding, ok := raw.(api.Embedding)
	if !ok {
		return fmt.Errorf(
			"failed to cast embedding to api.Embedding: %w",
			plugin.ErrUnexpectedInterface,
		)
	}

	local.embedding = embedding

	// NOTE(jkoelker) Start the garabage collector on the database.
	go func() {
		defer close(local.gcStopped)

		for {
			select {
			case <-time.After(GarbageCollectionInterval):
				if err := local.CollectGarbage(); err != nil {
					local.logger.Error("Failed to collect garbage", "error", err)
				}

			case <-local.closing:
				return
			}
		}
	}()

	return nil
}

// Close closes the database.
func (local *Local) Close(ctx context.Context) error {
	local.SetupLogger(ctx)

	local.logger.Info("Closing local plugin")

	close(local.closing)

	local.logger.Info("Waiting for garbage collector to stop")
	<-local.gcStopped

	if err := local.DB.Close(); err != nil {
		return fmt.Errorf("failed to close local database: %w", err)
	}

	local.manager.Close()

	return nil
}

// CollectGarbage runs a garbage collectin on the database.
func (local *Local) CollectGarbage() error {
	if local.logger != nil {
		local.logger.Debug("Collecting garbage")
	}

	if err := local.DB.RunValueLogGC(GarbageCollectionDiscardRatio); err != nil {
		// NOTE(jkoelker) If there is nothing to collect, badger will return
		//                `badger.ErrNoRewrite`. We can safely ignore this.
		if errors.Is(err, badger.ErrNoRewrite) {
			return nil
		}

		return fmt.Errorf("failed to run garbage collection: %w", err)
	}

	return nil
}

// Memorize implements the `api.Memory` interface by storing the data in
// the database, the data is embeddeed using the embedding plugin. The
// resulting vector is binary encoded as the key and the data is stored
// as the value.
func (local *Local) Memorize(ctx context.Context, data []string) error {
	local.SetupLogger(ctx)

	if local.DB == nil {
		if err := local.Open(ctx); err != nil {
			return fmt.Errorf("failed to open local database: %w", err)
		}
	}

	if err := local.DB.Update(func(txn *badger.Txn) error {
		for _, entry := range data {
			embedding, err := local.embedding.Embedding(ctx, entry)
			if err != nil {
				return fmt.Errorf("failed to embed data: %w", err)
			}

			var buf bytes.Buffer
			if err := binary.Write(&buf, binary.LittleEndian, embedding); err != nil {
				return fmt.Errorf("failed to encode embedding: %w", err)
			}

			if err := txn.Set(buf.Bytes(), []byte(entry)); err != nil {
				return fmt.Errorf("failed to memorize data: %w", err)
			}
		}

		return nil
	}); err != nil {
		return fmt.Errorf("failed to memorize data: %w", err)
	}

	return nil
}

// Recall implements the `api.Memory` interface by iterating the database and
// returning the nearest count entries, if count is not provided, it will
// return the nearest 1 entries.
func (local *Local) Recall(ctx context.Context, data string, count ...int) ([]string, error) {
	local.SetupLogger(ctx)

	if local.DB == nil {
		if err := local.Open(ctx); err != nil {
			return nil, fmt.Errorf("failed to open local database: %w", err)
		}
	}

	nearest := 1
	if len(count) > 0 {
		nearest = count[0]
	}

	embedding, err := local.embedding.Embedding(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("failed to embed data: %w", err)
	}

	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.LittleEndian, embedding); err != nil {
		return nil, fmt.Errorf("failed to encode embedding: %w", err)
	}

	closest := NewClosest(embedding, nearest)

	if err := local.DB.View(func(txn *badger.Txn) error {
		iter := txn.NewIterator(
			badger.IteratorOptions{
				PrefetchValues: false,
			},
		)
		defer iter.Close()

		// Iterate all the keys and keep trac of the nearest values. sorted by distance.
		for iter.Rewind(); iter.Valid(); iter.Next() {
			item := iter.Item()
			key := item.Key()

			var stored []float32
			if err := binary.Read(bytes.NewReader(key), binary.LittleEndian, &stored); err != nil {
				return fmt.Errorf("failed to decode embedding: %w", err)
			}

			value, err := item.ValueCopy(nil)
			if err != nil {
				return fmt.Errorf("failed to copy value: %w", err)
			}

			closest.Add(stored, value)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to recall data: %w", err)
	}

	return closest.Strings(), nil
}
