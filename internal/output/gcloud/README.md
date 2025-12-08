## gcloud

`dotevn` allows you to output credentials in an ephemeral `application_default_credentials.json` file:

```json
{
  "client_id": "...",
  "client_secret": "...",
  "quota_project_id": "your google project",
  "refresh_token": "...",
  "type": "authorized_user"
}
```

To configure it you need to get the secret from any of the supported sources and
fill the required values:

```yaml
- output: dotenv
  values:
    client_id:
      static:
        value: "YOURS3BUCKET"
      type: static

    client_secret:
      bash:
        command: op read op://Employee/GCLOUD_CLIENT_SECRET/password
      type: bash
    
    ...
```

It will create an ephemeral .netrc file and keep it in the `$GOOGLE_APPLICATION_CREDENTIALS` env var which `gcloud` cli tool will
directly read.

Once the environment gets closed, the ephemeral `application_default_credentials.json` file will get deleted.
