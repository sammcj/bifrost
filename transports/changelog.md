<!-- The pattern we follow here is to keep the changelog for the latest version -->
<!-- Old changelogs are automatically attached to the GitHub releases -->

- chore: version update framework to 1.1.18 and core to 1.2.16
- feat: Use all keys for list models request
- fix: handled panic when using gemini models with openai integration responses API requests
- chore: Added id, object, and model fields to Chat Completion responses from Bedrock and Cohere providers
- feat: Added control for sending back raw response using `x-bf-send-back-raw-response` header
- feat: Added support for gzip decompression of response bodies from all providers
- feat: Added support for Anthropic MAX account authentication in integration
- feat: Added support for Anthropic meta data and thinking signature fields in integration
- feat: Adds support for dynamic plugins. Note that dynamic plugins are in beta
- feat: Adds auth support for dashboard, inference APIs and dashboard APIs.