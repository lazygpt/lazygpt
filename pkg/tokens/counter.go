//

package tokens

import (
	"fmt"

	"github.com/tiktoken-go/tokenizer"

	"github.com/lazygpt/lazygpt/plugin/api"
)

// PrimedTokens is the number of tokens that are primed for every message.
const PrimedTokens = 3

// Counter is a token counter to sum for messages.
type Counter struct {
	// Model the counter is used for.
	Model string

	// Token count.
	Tokens int

	// encoding used for the model.
	encoding tokenizer.Codec

	// perMessage number of tokens per message.
	perMessage int

	// perName number of tokens per name.
	perName int
}

// NewCounter creates a new counter.
func NewCounter(model string) (*Counter, error) {
	var perMessage, perName int

	switch model {
	case "gpt-4-32k-0314", "gpt-4-32k", "gpt-4-0314", "gpt-4":
		perMessage = 3
		perName = 1
	case "gpt-3.5-turbo-0301", "gpt-3.5-turbo":
		perMessage = 4
		perName = -1
	default:
		// NOTE(jkoelker) default to gpt4
		perMessage = 3
		perName = 1
	}

	enc, err := tokenizer.Get(tokenizer.Cl100kBase)
	if err != nil {
		return nil, fmt.Errorf("failed to get encoding: %w", err)
	}

	return &Counter{
		Model:  model,
		Tokens: PrimedTokens,

		encoding:   enc,
		perMessage: perMessage,
		perName:    perName,
	}, nil
}

// Add adds a message to the counter taking into account the model to
// add necessary tokens.
func (c *Counter) Add(messages ...api.Message) error {
	for _, msg := range messages {
		c.Tokens += c.perMessage

		_, tokens, err := c.encoding.Encode(msg.Content)
		if err != nil {
			return fmt.Errorf("failed to encode message: %w", err)
		}

		c.Tokens += len(tokens)

		_, tokens, err = c.encoding.Encode(msg.Name)
		if err != nil {
			return fmt.Errorf("failed to encode name: %w", err)
		}

		c.Tokens += len(tokens)

		_, tokens, err = c.encoding.Encode(msg.Role)
		if err != nil {
			return fmt.Errorf("failed to encode role: %w", err)
		}

		c.Tokens += len(tokens)

		if msg.Name != "" {
			c.Tokens += c.perName
		}
	}

	return nil
}
