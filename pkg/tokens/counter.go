//

package tokens

import (
	"fmt"

	"github.com/pkoukk/tiktoken-go"

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
	encoding *tiktoken.Tiktoken

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

	tkm, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return nil, fmt.Errorf("failed to get encoding: %w", err)
	}

	return &Counter{
		Model:  model,
		Tokens: PrimedTokens,

		encoding:   tkm,
		perMessage: perMessage,
		perName:    perName,
	}, nil
}

// Add adds a message to the counter taking into account the model to
// add necessary tokens.
func (c *Counter) Add(msg api.Message) {
	c.Tokens += c.perMessage
	c.Tokens += len(c.encoding.Encode(msg.Content, nil, nil))
	c.Tokens += len(c.encoding.Encode(msg.Name, nil, nil))
	c.Tokens += len(c.encoding.Encode(msg.Role, nil, nil))

	if msg.Name != "" {
		c.Tokens += c.perName
	}
}
