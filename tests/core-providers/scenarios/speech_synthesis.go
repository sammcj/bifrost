package scenarios

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/maximhq/bifrost/tests/core-providers/config"

	bifrost "github.com/maximhq/bifrost/core"
	"github.com/maximhq/bifrost/core/schemas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RunSpeechSynthesisTest executes the speech synthesis test scenario
func RunSpeechSynthesisTest(t *testing.T, client *bifrost.Bifrost, ctx context.Context, testConfig config.ComprehensiveTestConfig) {
	if !testConfig.Scenarios.SpeechSynthesis {
		t.Logf("Speech synthesis not supported for provider %s", testConfig.Provider)
		return
	}

	t.Run("SpeechSynthesis", func(t *testing.T) {
		// Test with shared text constants for round-trip validation with transcription
		testCases := []struct {
			name           string
			text           string
			voiceType      string
			format         string
			expectMinBytes int
			saveForSST     bool // Whether to save this audio for SST round-trip testing
		}{
			{
				name:           "BasicText_Primary_MP3",
				text:           TTSTestTextBasic,
				voiceType:      "primary",
				format:         "mp3",
				expectMinBytes: 1000,
				saveForSST:     true,
			},
			{
				name:           "MediumText_Secondary_MP3",
				text:           TTSTestTextMedium,
				voiceType:      "secondary",
				format:         "mp3",
				expectMinBytes: 2000,
				saveForSST:     true,
			},
			{
				name:           "TechnicalText_Tertiary_MP3",
				text:           TTSTestTextTechnical,
				voiceType:      "tertiary",
				format:         "mp3",
				expectMinBytes: 500,
				saveForSST:     true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				voice := GetProviderVoice(testConfig.Provider, tc.voiceType)
				request := &schemas.BifrostRequest{
					Provider: testConfig.Provider,
					Model:    testConfig.SpeechSynthesisModel, // Use configured model
					Input: schemas.RequestInput{
						SpeechInput: &schemas.SpeechInput{
							Input: tc.text,
							VoiceConfig: schemas.SpeechVoiceInput{
								Voice: &voice,
							},
							ResponseFormat: tc.format,
						},
					},
					Params:    MergeModelParameters(&schemas.ModelParameters{}, testConfig.CustomParams),
					Fallbacks: testConfig.Fallbacks,
				}

				response, err := client.SpeechRequest(ctx, request)
				require.Nilf(t, err, "Speech synthesis failed: %v", err)
				require.NotNil(t, response)
				require.NotNil(t, response.Speech)
				require.NotNil(t, response.Speech.Audio)

				// Validate audio data
				assert.Greater(t, len(response.Speech.Audio), tc.expectMinBytes, "Audio data should have minimum expected size")
				assert.Equal(t, "audio.speech", response.Object)
				assert.Equal(t, testConfig.SpeechSynthesisModel, response.Model)

				// Save audio file for SST round-trip testing if requested
				if tc.saveForSST {
					tempDir := os.TempDir()
					audioFileName := filepath.Join(tempDir, "tts_"+tc.name+"."+tc.format)

					err := os.WriteFile(audioFileName, response.Speech.Audio, 0644)
					require.NoError(t, err, "Failed to save audio file for SST testing")

					// Register cleanup to remove temp file
					t.Cleanup(func() {
						os.Remove(audioFileName)
					})

					t.Logf("💾 Audio saved for SST testing: %s (text: '%s')", audioFileName, tc.text)
				}

				t.Logf("✅ Speech synthesis successful: %d bytes of %s audio generated for voice '%s'",
					len(response.Speech.Audio), tc.format, voice)
			})
		}
	})
}

// RunSpeechSynthesisAdvancedTest executes advanced speech synthesis test scenarios
func RunSpeechSynthesisAdvancedTest(t *testing.T, client *bifrost.Bifrost, ctx context.Context, testConfig config.ComprehensiveTestConfig) {
	if !testConfig.Scenarios.SpeechSynthesis {
		t.Logf("Speech synthesis not supported for provider %s", testConfig.Provider)
		return
	}

	t.Run("SpeechSynthesisAdvanced", func(t *testing.T) {
		t.Run("LongText_HDModel", func(t *testing.T) {
			// Test with longer text and HD model
			longText := `
			This is a comprehensive test of the text-to-speech functionality using a longer piece of text.
			The system should be able to handle multiple sentences, proper punctuation, and maintain 
			consistent voice quality throughout the entire speech generation process. This test ensures
			that the speech synthesis can handle realistic use cases with substantial content.
			`

			voice := "shimmer"
			request := &schemas.BifrostRequest{
				Provider: testConfig.Provider,
				Model:    "tts-1-hd", // Test with HD model
				Input: schemas.RequestInput{
					SpeechInput: &schemas.SpeechInput{
						Input: longText,
						VoiceConfig: schemas.SpeechVoiceInput{
							Voice: &voice,
						},
						ResponseFormat: "mp3",
						Instructions:   "Speak slowly and clearly with natural intonation.",
					},
				},
				Params:    MergeModelParameters(&schemas.ModelParameters{}, testConfig.CustomParams),
				Fallbacks: testConfig.Fallbacks,
			}

			response, err := client.SpeechRequest(ctx, request)
			require.Nilf(t, err, "HD speech synthesis failed: %v", err)
			require.NotNil(t, response)
			require.NotNil(t, response.Speech)
			require.NotNil(t, response.Speech.Audio)

			// Validate longer audio
			assert.Greater(t, len(response.Speech.Audio), 5000, "HD audio should be substantial")
			assert.Equal(t, "tts-1-hd", response.Model)

			t.Logf("✅ HD speech synthesis successful: %d bytes generated", len(response.Speech.Audio))
		})

		t.Run("AllVoiceOptions", func(t *testing.T) {
			// Test provider-specific voice options
			voiceTypes := []string{"primary", "secondary", "tertiary"}
			testText := TTSTestTextBasic // Use shared constant

			for _, voiceType := range voiceTypes {
				t.Run("VoiceType_"+voiceType, func(t *testing.T) {
					voice := GetProviderVoice(testConfig.Provider, voiceType)
					request := &schemas.BifrostRequest{
						Provider: testConfig.Provider,
						Model:    testConfig.SpeechSynthesisModel,
						Input: schemas.RequestInput{
							SpeechInput: &schemas.SpeechInput{
								Input: testText,
								VoiceConfig: schemas.SpeechVoiceInput{
									Voice: &voice,
								},
								ResponseFormat: "mp3",
							},
						},
						Params:    MergeModelParameters(&schemas.ModelParameters{}, testConfig.CustomParams),
						Fallbacks: testConfig.Fallbacks,
					}

					response, err := client.SpeechRequest(ctx, request)
					require.Nilf(t, err, "Speech synthesis failed for voice %s (%s): %v", voice, voiceType, err)
					require.NotNil(t, response)
					require.NotNil(t, response.Speech)
					require.NotNil(t, response.Speech.Audio)

					assert.Greater(t, len(response.Speech.Audio), 500, "Audio should be generated for voice %s", voice)
					t.Logf("✅ Voice %s (%s): %d bytes generated", voice, voiceType, len(response.Speech.Audio))
				})
			}
		})
	})
}
