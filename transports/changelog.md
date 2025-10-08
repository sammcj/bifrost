<!-- The pattern we follow here is to keep the changelog for the latest version -->
<!-- Old changelogs are automatically attached to the GitHub releases -->

- Feat: A new config called `Enable LiteLLM Fallback` that enables text_completion calls to fall back to chat_completions calls for the Groq provider. This is an anti-pattern, but we are adding this to help users migrate from LiteLLM easily. Reach out to us if you want us to enable any other quirky patterns LiteLLM has.