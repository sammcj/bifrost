package scenarios

import (
	"context"
	"testing"

	"github.com/maximhq/bifrost/tests/core-providers/config"

	bifrost "github.com/maximhq/bifrost/core"
	"github.com/maximhq/bifrost/core/schemas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RunTextCompletionTest tests text completion functionality
func RunTextCompletionTest(t *testing.T, client *bifrost.Bifrost, ctx context.Context, testConfig config.ComprehensiveTestConfig) {
	if !testConfig.Scenarios.TextCompletion || testConfig.TextModel == "" {
		t.Logf("⏭️ Text completion not supported for provider %s", testConfig.Provider)
		return
	}

	t.Run("TextCompletion", func(t *testing.T) {
		prompt := "The future of artificial intelligence is"
		request := &schemas.BifrostRequest{
			Provider: testConfig.Provider,
			Model:    testConfig.TextModel,
			Input: schemas.RequestInput{
				TextCompletionInput: &prompt,
			},
			Params: MergeModelParameters(&schemas.ModelParameters{
				MaxTokens: bifrost.Ptr(100),
			}, testConfig.CustomParams),
			Fallbacks: testConfig.Fallbacks,
		}

		response, err := client.TextCompletionRequest(ctx, request)
		require.Nilf(t, err, "Text completion failed: %v", err)
		require.NotNil(t, response)
		require.NotEmpty(t, response.Choices)

		content := GetResultContent(response)
		assert.NotEmpty(t, content, "Response content should not be empty")
		t.Logf("✅ Text completion result: %s", content)
	})
}
