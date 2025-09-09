package scenarios

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/maximhq/bifrost/tests/core-providers/config"

	bifrost "github.com/maximhq/bifrost/core"
	"github.com/maximhq/bifrost/core/schemas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RunTranscriptionStreamTest executes the streaming transcription test scenario
func RunTranscriptionStreamTest(t *testing.T, client *bifrost.Bifrost, ctx context.Context, testConfig config.ComprehensiveTestConfig) {
	if !testConfig.Scenarios.TranscriptionStream {
		t.Logf("Transcription streaming not supported for provider %s", testConfig.Provider)
		return
	}

	t.Run("TranscriptionStream", func(t *testing.T) {
		// Generate TTS audio for streaming round-trip validation
		streamRoundTripCases := []struct {
			name           string
			text           string
			voiceType      string
			format         string
			responseFormat *string
			expectChunks   int
		}{
			{
				name:           "StreamRoundTrip_Basic_MP3",
				text:           TTSTestTextBasic,
				voiceType:      "primary",
				format:         "mp3",
				responseFormat: nil, // Default JSON streaming
				expectChunks:   1,
			},
			{
				name:           "StreamRoundTrip_Medium_MP3",
				text:           TTSTestTextMedium,
				voiceType:      "secondary",
				format:         "mp3",
				responseFormat: bifrost.Ptr("json"),
				expectChunks:   1,
			},
			{
				name:           "StreamRoundTrip_Technical_MP3",
				text:           TTSTestTextTechnical,
				voiceType:      "tertiary",
				format:         "mp3",
				responseFormat: bifrost.Ptr("json"),
				expectChunks:   1,
			},
		}

		for _, tc := range streamRoundTripCases {
			t.Run(tc.name, func(t *testing.T) {
				// Step 1: Generate TTS audio
				voice := GetProviderVoice(testConfig.Provider, tc.voiceType)
				ttsRequest := &schemas.BifrostRequest{
					Provider: testConfig.Provider,
					Model:    testConfig.SpeechSynthesisModel,
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

				ttsResponse, err := client.SpeechRequest(ctx, ttsRequest)
				require.Nilf(t, err, "TTS generation failed for stream round-trip test: %v", err)
				require.NotNil(t, ttsResponse.Speech)
				require.NotNil(t, ttsResponse.Speech.Audio)
				require.Greater(t, len(ttsResponse.Speech.Audio), 0, "TTS returned empty audio")

				// Save temp audio file
				tempDir := os.TempDir()
				audioFileName := filepath.Join(tempDir, "stream_roundtrip_"+tc.name+"."+tc.format)
				writeErr := os.WriteFile(audioFileName, ttsResponse.Speech.Audio, 0644)
				require.NoError(t, writeErr, "Failed to save temp audio file")

				// Register cleanup
				t.Cleanup(func() {
					os.Remove(audioFileName)
				})

				t.Logf("🔄 Generated TTS audio for stream round-trip: %s (%d bytes)", audioFileName, len(ttsResponse.Speech.Audio))

				// Step 2: Test streaming transcription
				streamRequest := &schemas.BifrostRequest{
					Provider: testConfig.Provider,
					Model:    testConfig.TranscriptionModel,
					Input: schemas.RequestInput{
						TranscriptionInput: &schemas.TranscriptionInput{
							File:           ttsResponse.Speech.Audio,
							Language:       bifrost.Ptr("en"),
							Format:         bifrost.Ptr(tc.format),
							ResponseFormat: tc.responseFormat,
						},
					},
					Params: MergeModelParameters(&schemas.ModelParameters{
						Temperature: bifrost.Ptr(0.0), // More deterministic output
					}, testConfig.CustomParams),
					Fallbacks: testConfig.Fallbacks,
				}

				// Test streaming response
				responseChannel, err := client.TranscriptionStreamRequest(ctx, streamRequest)
				require.Nilf(t, err, "Transcription stream failed: %v", err)
				require.NotNil(t, responseChannel, "Response channel should not be nil")

				var fullTranscriptionText string
				var chunkCount int
				var lastResponse *schemas.BifrostStream

				// Create a timeout context for the stream reading
				streamCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
				defer cancel()

				// Read streaming chunks
				for {
					select {
					case response, ok := <-responseChannel:
						if !ok {
							// Channel closed, streaming complete
							goto streamComplete
						}

						require.NotNil(t, response, "Stream response should not be nil")

						// Check for errors in stream
						if response.BifrostError != nil {
							t.Fatalf("Error in stream: %v", response.BifrostError)
						}

						require.NotNil(t, response.Transcribe, "Transcribe data should be present in stream")

						// Collect transcription chunks
						if response.Transcribe.Text != "" {
							if response.Transcribe.BifrostTranscribeStreamResponse != nil &&
								response.Transcribe.BifrostTranscribeStreamResponse.Delta != nil {
								// This is a delta chunk
								fullTranscriptionText += *response.Transcribe.BifrostTranscribeStreamResponse.Delta
							} else {
								// This is a complete text chunk
								fullTranscriptionText += response.Transcribe.Text
							}
							chunkCount++
							t.Logf("Received transcription chunk %d: '%s'", chunkCount, response.Transcribe.Text)
						}

						// Validate stream response structure
						assert.Equal(t, "audio.transcription.chunk", response.Object)
						assert.Equal(t, testConfig.TranscriptionModel, response.Model)
						assert.Equal(t, testConfig.Provider, response.ExtraFields.Provider)

						lastResponse = response

					case <-streamCtx.Done():
						t.Fatal("Stream reading timed out")
					}
				}

			streamComplete:
				// Validate streaming results
				assert.GreaterOrEqual(t, chunkCount, tc.expectChunks, "Should receive minimum expected chunks")
				assert.NotNil(t, lastResponse, "Should have received at least one response")

				// Validate round-trip: check if transcribed text contains key words from original
				require.NotEmpty(t, fullTranscriptionText, "Transcribed text should not be empty")

				// Normalize for comparison (lowercase, remove punctuation)
				originalWords := strings.Fields(strings.ToLower(tc.text))
				transcribedWords := strings.Fields(strings.ToLower(fullTranscriptionText))

				// Check that at least 50% of original words are found in transcription
				foundWords := 0
				for _, originalWord := range originalWords {
					// Remove punctuation for comparison
					cleanOriginal := strings.Trim(originalWord, ".,!?;:")
					if len(cleanOriginal) < 3 { // Skip very short words
						continue
					}

					for _, transcribedWord := range transcribedWords {
						cleanTranscribed := strings.Trim(transcribedWord, ".,!?;:")
						if strings.Contains(cleanTranscribed, cleanOriginal) || strings.Contains(cleanOriginal, cleanTranscribed) {
							foundWords++
							break
						}
					}
				}

				// Expect at least 50% word match for successful round-trip
				minExpectedWords := len(originalWords) / 2
				assert.GreaterOrEqual(t, foundWords, minExpectedWords,
					"Stream round-trip failed: original='%s', transcribed='%s', found %d/%d words",
					tc.text, fullTranscriptionText, foundWords, len(originalWords))

				t.Logf("✅ Stream round-trip successful: '%s' → TTS → SST → '%s' (%d chunks, found %d/%d words)",
					tc.text, fullTranscriptionText, chunkCount, foundWords, len(originalWords))
			})
		}
	})
}

// RunTranscriptionStreamAdvancedTest executes advanced streaming transcription test scenarios
func RunTranscriptionStreamAdvancedTest(t *testing.T, client *bifrost.Bifrost, ctx context.Context, testConfig config.ComprehensiveTestConfig) {
	if !testConfig.Scenarios.TranscriptionStream {
		t.Logf("Transcription streaming not supported for provider %s", testConfig.Provider)
		return
	}

	t.Run("TranscriptionStreamAdvanced", func(t *testing.T) {
		t.Run("JSONStreaming", func(t *testing.T) {
			// Generate audio for streaming test
			audioData, _ := GenerateTTSAudioForTest(ctx, t, client, testConfig.Provider, testConfig.SpeechSynthesisModel, TTSTestTextBasic, "primary", "mp3")

			// Test streaming with JSON format
			request := &schemas.BifrostRequest{
				Provider: testConfig.Provider,
				Model:    testConfig.TranscriptionModel,
				Input: schemas.RequestInput{
					TranscriptionInput: &schemas.TranscriptionInput{
						File:           audioData,
						Language:       bifrost.Ptr("en"),
						Format:         bifrost.Ptr("mp3"),
						ResponseFormat: bifrost.Ptr("json"),
					},
				},
				Params:    MergeModelParameters(&schemas.ModelParameters{}, testConfig.CustomParams),
				Fallbacks: testConfig.Fallbacks,
			}

			responseChannel, err := client.TranscriptionStreamRequest(ctx, request)
			require.Nilf(t, err, "JSON streaming failed: %v", err)

			var receivedResponse bool
			streamCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			for {
				select {
				case response, ok := <-responseChannel:
					if !ok {
						goto verboseStreamComplete
					}

					if response.BifrostError != nil {
						t.Fatalf("Error in verbose stream: %v", response.BifrostError)
					}

					if response.Transcribe != nil {
						receivedResponse = true

						// Check for verbose_json specific fields
						if response.Transcribe.BifrostTranscribeStreamResponse != nil {
							t.Logf("Stream type: %v", response.Transcribe.BifrostTranscribeStreamResponse.Type)
							if response.Transcribe.BifrostTranscribeStreamResponse.Delta != nil {
								t.Logf("Delta: %s", *response.Transcribe.BifrostTranscribeStreamResponse.Delta)
							}
						}
					}

				case <-streamCtx.Done():
					t.Fatal("Verbose stream reading timed out")
				}
			}

		verboseStreamComplete:
			assert.True(t, receivedResponse, "Should receive at least one response")
			t.Logf("✅ Verbose JSON streaming successful")
		})

		t.Run("MultipleLanguages_Streaming", func(t *testing.T) {
			// Generate audio for language streaming tests
			audioData, _ := GenerateTTSAudioForTest(ctx, t, client, testConfig.Provider, testConfig.SpeechSynthesisModel, TTSTestTextBasic, "primary", "mp3")

			// Test streaming with different language hints (only English for now)
			languages := []string{"en"}

			for _, lang := range languages {
				t.Run("StreamLang_"+lang, func(t *testing.T) {
					langCopy := lang
					request := &schemas.BifrostRequest{
						Provider: testConfig.Provider,
						Model:    testConfig.TranscriptionModel,
						Input: schemas.RequestInput{
							TranscriptionInput: &schemas.TranscriptionInput{
								File:     audioData,
								Language: &langCopy,
							},
						},
						Params:    MergeModelParameters(&schemas.ModelParameters{}, testConfig.CustomParams),
						Fallbacks: testConfig.Fallbacks,
					}

					responseChannel, err := client.TranscriptionStreamRequest(ctx, request)
					require.Nilf(t, err, "Streaming failed for language %s: %v", lang, err)

					var receivedData bool
					streamCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
					defer cancel()

					for {
						select {
						case response, ok := <-responseChannel:
							if !ok {
								goto langStreamComplete
							}

							if response.BifrostError != nil {
								t.Fatalf("Error in stream for language %s: %v", lang, response.BifrostError)
							}

							if response.Transcribe != nil {
								receivedData = true
							}

						case <-streamCtx.Done():
							t.Fatalf("Stream timed out for language %s", lang)
						}
					}

				langStreamComplete:
					assert.True(t, receivedData, "Should receive transcription data for language %s", lang)
					t.Logf("✅ Streaming successful for language: %s", lang)
				})
			}
		})

		t.Run("WithCustomPrompt_Streaming", func(t *testing.T) {
			// Generate audio for custom prompt streaming test
			audioData, _ := GenerateTTSAudioForTest(ctx, t, client, testConfig.Provider, testConfig.SpeechSynthesisModel, TTSTestTextTechnical, "tertiary", "mp3")

			// Test streaming with custom prompt for context
			request := &schemas.BifrostRequest{
				Provider: testConfig.Provider,
				Model:    testConfig.TranscriptionModel,
				Input: schemas.RequestInput{
					TranscriptionInput: &schemas.TranscriptionInput{
						File:     audioData,
						Language: bifrost.Ptr("en"),
						Prompt:   bifrost.Ptr("This audio contains technical terms, proper nouns, and streaming-related vocabulary."),
					},
				},
				Params: MergeModelParameters(&schemas.ModelParameters{
					Temperature: bifrost.Ptr(0.1),
				}, testConfig.CustomParams),
				Fallbacks: testConfig.Fallbacks,
			}

			responseChannel, err := client.TranscriptionStreamRequest(ctx, request)
			require.Nilf(t, err, "Custom prompt streaming failed: %v", err)

			var chunkCount int
			streamCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			for {
				select {
				case response, ok := <-responseChannel:
					if !ok {
						goto promptStreamComplete
					}

					if response.BifrostError != nil {
						t.Fatalf("Error in prompt stream: %v", response.BifrostError)
					}

					if response.Transcribe != nil && response.Transcribe.Text != "" {
						chunkCount++
					}

				case <-streamCtx.Done():
					t.Fatal("Prompt stream reading timed out")
				}
			}

		promptStreamComplete:
			assert.Greater(t, chunkCount, 0, "Should receive at least one transcription chunk")
			t.Logf("✅ Custom prompt streaming successful: %d chunks received", chunkCount)
		})
	})
}
