<!-- The pattern we follow here is to keep the changelog for the latest version -->
<!-- Old changelogs are automatically attached to the GitHub releases -->

- Added specific error handling for timeout scenarios (context.Canceled, context.DeadlineExceeded, fasthttp.ErrTimeout) across all providers
- Created a dedicated error message for timeouts that guides users to adjust the timeout setting
- Added Text Completion Streaming support