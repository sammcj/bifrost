<!-- The pattern we follow here is to keep the changelog for the latest version -->
<!-- Old changelogs are automatically attached to the GitHub releases -->

- feat: Use all keys for list models request
- refactor: Cohere provider to use completeRequest and response pooling for all requests
- chore: Added id, object, and model fields to Chat Completion responses from Bedrock and Cohere providers
- feat: Add request level control for adding extra headers, url path, skipping key selection, and sending back raw response
- feat: Added support for gzip decompression of response bodies from all providers
- feat: Added support for anthropic's meta data and thinking signature fields
- feat: Moved all streaming calls to use fasthttp client for efficiency
- refactor: Moved all convertors from schemas/providers to separate provider packages in providers directory