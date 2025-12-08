## bash

`bash` is a common source, it allows you to declare how to generate the value
using a bash command. For example:

```yaml
GITHUB_TOKEN:
  type: bash
  bash:
    command: gh auth token
```

It is useful to obtain/generate secrets using CLI tools.
