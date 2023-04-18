//

package tokens_test

import (
	"testing"

	"gotest.tools/v3/assert"

	. "github.com/lazygpt/lazygpt/pkg/tokens"
	"github.com/lazygpt/lazygpt/plugin/api"
)

func exampleMessages() []api.Message {
	return []api.Message{
		{
			Role:    "system",
			Content: "You are a helpful, pattern-following assistant that translates corporate jargon into plain English.",
		},
		{
			Role:    "system",
			Name:    "example_user",
			Content: "New synergies will help drive top-line growth.",
		},
		{
			Role:    "system",
			Name:    "example_assistant",
			Content: "Things working well together will increase revenue.",
		},
		{
			Role:    "system",
			Name:    "example_user",
			Content: "Let's circle back when we have more bandwidth to touch base on opportunities for increased leverage.",
		},
		{
			Role:    "system",
			Name:    "example_assistant",
			Content: "Let's talk later when we're less busy about how to do better.",
		},
		{
			Role:    "user",
			Content: "This late pivot means we don't have time to boil the ocean for the client deliverable.",
		},
	}
}

func TestCounterGPT35Turbo0301(t *testing.T) {
	t.Parallel()

	counter, err := NewCounter("gpt-3.5-turbo-0301")
	assert.NilError(t, err)

	messages := exampleMessages()
	for _, message := range messages {
		counter.Add(message)
	}

	assert.Equal(t, counter.Tokens, 127)
}

func TestCounterGPT40314(t *testing.T) {
	t.Parallel()

	counter, err := NewCounter("gpt-4-0314")
	assert.NilError(t, err)

	messages := exampleMessages()
	for _, message := range messages {
		counter.Add(message)
	}

	assert.Equal(t, counter.Tokens, 129)
}
