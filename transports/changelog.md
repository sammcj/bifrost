<!-- The pattern we follow here is to keep the changelog for the latest version -->
<!-- Old changelogs are automatically attached to the GitHub releases -->

- Upgrade dependency: core to 1.2.5 and framework to 1.1.5
- Feat: Added Anthropic thinking parameter in responses API.
- Feat: Added Anthropic text completion integration support.
- Fix: Extra fields sent back in streaming responses.
- Feat: Latency for all request types (with inter token latency for streaming requests) sent back in Extra fields.
- Feat: UI websocket implementation generalized.
- Feat: TokenInterceptor interface added to plugins.
- Fix: Middlewares added to integrations route.
