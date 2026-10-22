---
name: chasky-env
description: Run a command with a chasky environment's secrets loaded from its dotenv output, delivered through an anonymous pipe (`--fifo --fifo-mode=fd`) so nothing is written to disk. Use when the user asks to load, use or run something with a chasky environ (e.g. "run X with the `my_app` env").
argument-hint: <environ> [-- command args...]
---

# chasky-env

Loads a chasky environ (`output: env` or `output: dotenv`) and makes its variables
available to a command, without the secrets ever touching the disk.

## How it works

```bash
chasky <env> --fifo --fifo-mode=fd -- <command>
```

- `--fifo` makes the file-backed outputs use a pipe instead of a temp file. With `--fifo`, `output: env` stops exporting the variables directly and exposes them as a file too.
- `--fifo-mode=fd` exposes it as an inherited anonymous pipe (`/dev/fd/N`) that only
  the spawned command (and its children) can open. The path is passed in `$DOTENV_FILE` (for both `output: env` and `output: dotenv`).
- The pipe is **single-read**: the first reader gets the contents, a second read sees EOF.
  So read `$DOTENV_FILE` exactly once, and do it inside the spawned command.
- chasky refuses to run inside another chasky environment (`CHASKY_ENVNAME` set).
- Unix only (Linux, macOS).

## Steps

1. Take the environ name from the user's request (`$ARGUMENTS`). If it is missing, ask the
   user which environ to use.
2. Load the env file once (`eval "$(cat ...)"`, not `.`/`source`: some shells read 0 bytes when they size a pipe up front) inside the spawned process, then exec the target command.
   Each Bash tool call is a fresh shell, so the environ only lives for that one call;
   bundle everything that needs the secrets into the same invocation:

   ```bash
   chasky <env> --fifo --fifo-mode=fd -- sh -c '
     set -a; eval "$(cat "$DOTENV_FILE")"; set +a
     exec "$@"
   ' sh <command> [args...]
   ```

   For several commands, pass `sh -c 'set -a; eval "$(cat "$DOTENV_FILE")"; set +a; cmd1 && cmd2'`
   instead of exec'ing "$@".
3. If the consumer natively takes an env file (e.g. `docker run --env-file`), skip the
   sourcing and pass the path directly. It still gets one read:

   ```bash
   chasky <env> --fifo --fifo-mode=fd -- sh -c 'docker run --env-file "$DOTENV_FILE" image'
   ```

   Note the single quotes: the variable must expand inside the chasky-spawned shell,
   not in the calling shell.
4. Report the command's result to the user.

## Rules

- Never `cat`/`echo`/print `$DOTENV_FILE`, the sourced variables, or `env` output; these
  are secrets and would end up in the transcript. If you need to confirm a variable is
  set, test it without printing it: `[ -n "$VAR" ] && echo "VAR is set"`.
- Never copy the dotenv contents to a file or fall back to the default disk mode
  (omitting `--fifo`) unless the user explicitly asks for it.
- Do not read `$DOTENV_FILE` more than once (e.g. no `source` followed by a second
  `--env-file`); the second read is empty. If two consumers are needed, source once
  and let the second inherit the exported variables.
- If the command fails with an empty environment, the likely causes are: the pipe was
  already read, the consumer cannot read `/dev/fd/N`, or the environ has neither an `env` nor a `dotenv`
  output. Check the environ's config with `chasky edit` (ask the user first) rather
  than switching modes silently.
- `named-pipe` mode (`--fifo` alone) is reachable by any same-user process; only suggest
  it if the user wants multiple reads and accepts that trade-off.
