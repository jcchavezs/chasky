# Chasky

Chasky is a secrets dealer. Declare the secrets you need on every application as environment variables or other setups (e.g. ephemeral .env or .netrc files) and use your apps without the security concern of leaking secrets on terminal history or abandoned `.env` files.

## Installation

```console
go install github.com/jcchavezs/chasky/cmd/chasky@latest
```

## Getting started

First you need to declare your secrets in `~/.chasky.yaml` under the following syntax:

```yaml
myapp: # This comment will show up as the description of the environ
- output: env # defines how to output the credentials
  values:
    GITHUB_TOKEN: # Name of the value inside the output
      type: bash # Type of source
      bash: # Source config
        command: gh auth token # The command to be executed whose output will populate the variable
    JIRA_TOKEN:
      type: bash
      bash: 
        command: op item get op://Employee/my_jira_access/password
    #...
```

## Quick links

- [Features](./Features)
- [Sources](./Sources)
- [Outputs](./Outputs)
