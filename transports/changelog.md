<!-- The pattern we follow here is to keep the changelog for the latest version -->
<!-- Old changelogs are automatically attached to the GitHub releases -->

- Added specific error handling for timeout scenarios (context.Canceled, context.DeadlineExceeded, fasthttp.ErrTimeout) across all providers
- Created a dedicated error message for timeouts that guides users to adjust the timeout setting
- Fixed validation in HTTP handlers for embeddings, speech, and text completion requests
- Improved CORS wildcard pattern matching to support domain patterns like *.example.com
- Fixed issues in the logging plugin to properly handle text completion responses
- Enhanced UI form handling for network configuration with proper default values
- Feat: Adds Text Completion Streaming support
