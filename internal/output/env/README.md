## env

`env` allows you to output credentials as env vars in your shell:

```bash
S3_BUCKET="YOURS3BUCKET"
SECRET_KEY="YOURSECRETKEYGOESHERE"
```

To configure it you need to get the secret from any of the supported sources and
fill the required values:

```yaml
- output: env
  values:
    S3_BUCKET:
      static:
        value: "YOURS3BUCKET"
      type: static

    SECRET_KEY:
      bash:
        command: op read op://Employee/S3_ACCESS/password
      type: bash
```

By default the values are injected directly into the spawned process environment.
When [`--fifo`](./Features#on-demand-files-via-pipes---fifo) is enabled the secrets
are not placed in the environment (where they would be readable through
`/proc/<pid>/environ`); instead they are written to an on-demand env file whose
path is exposed in the `$ENV_FILE` env var:

```bash
$ set -a && eval "$(cat "$ENV_FILE")" && set +a
```

Read the file with `cat` rather than sourcing it directly (`. "$ENV_FILE"`):
`$ENV_FILE` is a pipe, and some shells read 0 bytes because they size the file up
front (a pipe reports a size of 0).
