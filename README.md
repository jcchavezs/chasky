# Chasky

Chasky is a secrets dealer. Declare the secrets you need on every tool
as environment variables or other setups and use your tooling without the security concern
of leaking secrets on `.env` or `.envrc` files.

## Installation

```console
go install github.com/jcchavezs/chasky/cmd/chasky@latest
```

## Getting started

First you need to declare your secrets in `~/.chasky.yaml` under the following syntax:

```yaml
mytool:
  - output: env # defines how to output the credentials
    values:
      GITHUB_TOKEN:
        type: bash
        bash: 
          command: gh auth token
      JIRA_TOKEN:
        type: bash
        bash: 
          command: op item get op://Employee/my_jira_access/password
      #...
```

then inject the values using

```console
$ chasky mytool

Generating env vars for "mytool"...
```

## Rationale

When dealing with multiple tools, having to inject environment variables is cumbersome and using `.env` files imposes a security risk as those files store the values in plain text in your filesystem. Password managers like `1password` or lastpass are good to store secrets but not to deliver them to apps, different namings across tools, different vaults can be hard to maintain as once one secrets expires you need to change it in multiple vaults.

`chasky` allows you to declare the secrets in one place and maintain the secrets somewhere else so you could create a token, store in 1password and inject to independently.
