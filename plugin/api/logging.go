//

package api

import (
	"context"

	"github.com/lazygpt/lazygpt/plugin/log"
)

// InitLogging initializes a logger in the context.
func InitLogging(ctx context.Context, name string) context.Context {
	return log.NewContext(ctx, log.NewLogger(name))
}
