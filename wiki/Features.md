## Environ

The most common use case for `chasky` is to generate an environ (through the Shell) that
include a set of environment variables to be able to use a certain tool. However
along with the environment variables we can leverage some other features.

### Hooks

Hooks allow to run arbitrary actions before/after an environ is run. This is useful
to run certain actions that are needed to run the desired tool.

For example:

```yaml
codex: # Codex APP
- output: variables # Keep the secrets in the variables
  pre: # Before the environ is created
    - type: command
      # Render the variable OPENAI_API_KEY before execute the command
      command: "echo {{ $.OPENAI_API_KEY }} | codex login --with-api-key" 
  post: # After the environ is closed
    - type: command
      command: codex logout
  values:
    OPENAI_API_KEY:
      bash:
        command: op read op://Employee/OpenAI/password
      type: bash
```

#### Pre

Runs arbitrary actions before the environ is created. This is useful to execute logins and set configs.

#### Post

Runs arbitrary actions after the environ is closed. This is useful to execute logouts to avoid idle sessions.

### Inline environs

Creates an environment and runs a command without exporting the environ to the shell.

```bash
$ chasky my_app -- echo "I am ${MY_USER_ENV_VAR}"
```

## Migrating secrets

A good way to start migrating your secrets into chasky environments is to onboard them into a keyring or other password manager.

```console
$ chasky import keyring MY_KEY=MY_VALUE
```
