# Chasky

Chasky is a secrets dealer. Declare the secrets you need on every application as environment variables or other setups (e.g. ephemeral netrc files) and use your apps without the security concern of leaking secrets on terminal history or abandoned `.env` or `.envrc` files.

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
$ chasky myapp

Generating env vars for "myapp"...
```

or inject the values on the fly:

```console
chasky myapp -- <command> <arg1> <arg2> ...
```

### Migrating secrets

A good way to start migrating your secrets into chasky environments is to onboard them into a keyring or other password manager.

```console
chasky import keyring MY_KEY=MY_VALUE
```

## Motivation

When dealing with multiple apps, having to inject environment variables everytime is cumbersome and also using `.env` files imposes a security risk as those files store the values in plain text in your filesystem. Password managers like `1password` or `lastpass` are good to store secrets but not to deliver them to apps, different namings across apps and multiple vaults can be hard to maintain as once one secrets expires you need to change it in multiple vaults.

`chasky` allows you to declare the secrets in one place and maintain the secrets somewhere else so you could create a token, store in 1password and inject to independently. **This is, store the secret in the way that makes more sense for you and deliver into apps in the way that makes more sense for the apps.**
