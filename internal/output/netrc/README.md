## netrc

`netrc` allows you to output credentials in the `.netrc` file format:

```netrc
machine api.github.com login jcchavezs password gho_I2d3DoDZxxxxxxxxxxvBmqL4un1pLZ30ERXU
```

To configure it you need to get the secret from any of the supported sources and
fill the required values:

```yaml
- output: netrc
  values:
    machine:
      static:
        value: "api.github.com"
      type: static

    login:
      bash:
        command: "gh api /user | jq .login"
      type: bash

    password:
      bash:
        command: "gh auth token"
      type: bash
```

It will create an ephemeral .netrc file and keep it in the `$NETRC_FILE` env var to be used like:

```bash
$ curl --netrc-file $NETRC_FILE ....
```

Once the environment gets closed, the ephemeral `.netrc` file will get deleted.
