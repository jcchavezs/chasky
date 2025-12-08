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
