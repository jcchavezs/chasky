## Environ

The most common use case for `chasky` is to generate an environ (through the Shell) that
inject a set of environment variables to be consumed by a tool. However
along with the environment variables we can leverage some other features:

### Hooks

Hooks allow to run arbitrary actions before/after an environ is run. This is useful
to run actions that are pre-requisites in the usage of the desired tool.

For example:

```yaml
codex: # (OpenAI) Codex CLI
- output: variables # Keep the secrets in the variables
  pre: # Before the environ is created
    - type: command
      # Render the variable OPENAI_API_KEY before execute the command
      command: "echo {{ $.OPENAI_API_KEY }} | codex login --with-api-key" 
  post: # After the environ is closed
    - type: command
      # Logout from codex
      command: codex logout
  values:
    OPENAI_API_KEY:
      bash: # Read the secret from 1Password (op)
        command: op read op://Employee/OpenAI/password
      type: bash
```

#### Pre

Runs arbitrary actions before the environ is created. This is useful to execute logins and set configs.

#### Post

Runs arbitrary actions after the environ is closed. This is useful to execute logouts to avoid orphaned sessions.

### Inline environs

Creates an environment and runs a command without exporting the environ to the shell.

```bash
chasky my_app -- echo "I am ${MY_USER_ENV_VAR}"
```

### On-demand files via pipes (`--fifo`)

Some outputs (`dotenv`, `netrc`, `gcloud`) materialize the secrets as a file so a
tool can read them (`--env-file`, `--netrc-file`, `GOOGLE_APPLICATION_CREDENTIALS`, ...).
By default that file is a temporary file written to disk, which means the secrets
briefly live in plain text on your filesystem.

Passing `--fifo` makes `chasky` expose an on-demand pipe at that path instead:
the contents are delivered only when the reader process actually opens the file,
so nothing is ever written to disk. In every mode the path is still handed to the
process through the usual env var (`$DOTENV_FILE`, `$NETRC_FILE`,
`$GOOGLE_APPLICATION_CREDENTIALS`, ...).

`--fifo-mode` selects how the pipe is exposed:

- `named-pipe` *(default)* — a [UNIX named pipe (FIFO)](https://en.wikipedia.org/wiki/Named_pipe)
  on the filesystem (`0600`, inside a `0700` directory). Reachable by any process
  of the **same user** (a FIFO carries no reader identity), and readable more
  than once.
- `fd` — an anonymous pipe inherited by the spawned process and exposed as
  `/dev/fd/N`. Reachable by **only the spawned process** (and its children), and
  delivered a single time.

```bash
# Default: a named pipe, readable more than once by same-user processes.
chasky my_app --fifo -- docker run --env-file "$DOTENV_FILE" ...

# Hardened: an inherited descriptor, reachable only by the spawned process.
chasky my_app --fifo --fifo-mode=fd -- docker run --env-file "$DOTENV_FILE" ...
```

Use `fd` when you want the secret scoped to exactly the process you launch:
because an anonymous pipe has no name on the filesystem, no other process can
open it. The trade-off is that it is delivered a single time (a second read sees
EOF) and the consuming tool must be able to read a `/dev/fd/N` path.

The named pipe is removed once the environment is closed.

> [!NOTE]
> Both modes rely on UNIX pipe semantics, so `--fifo` is only supported on
> unix-based systems (Linux, macOS, ...).

## Migrating secrets

A good way to start migrating your secrets into chasky environments is to onboard them into a keyring or other password manager.

```console
$ chasky import keyring MY_KEY=MY_VALUE

Credentials successfully imported into keyring.

To use them in a given environment, type `chasky edit` and add:

---
# ...
- values:
  - MY_KEY:
      type: keyring
      keyring:
        key: com.github.jcchavezs.pakay-my_key
```
