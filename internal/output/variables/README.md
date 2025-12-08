## variables

`variables` allows you to output credentials as variables that can be rendered later
rendered in hooks.

To configure it you need to get the secret from any of the supported sources and
fill the required values:

```yaml
codex: # Codex APP
- output: variables
  pre:
    - type: command
      command: "echo {{ $.OPENAI_API_KEY }} | codex login --with-api-key"
  post:
    - type: command
      command: codex logout
  values:
    OPENAI_API_KEY:
      bash:
        command: op read op://Employee/OpenAI/password
      type: bash
```
