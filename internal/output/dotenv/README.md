## dotenv

`dotenv` allows you to output credentials in an ephemeral `.env` file:

```bash
S3_BUCKET="YOURS3BUCKET"
SECRET_KEY="YOURSECRETKEYGOESHERE"
```

To configure it you need to get the secret from any of the supported sources and
fill the required values:

```yaml
- output: dotenv
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

Once the environment gets closed, the ephemeral `.env` file will get deleted.
