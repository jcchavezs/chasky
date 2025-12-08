## keyring

`keyring` allows to access variables from the keyring system. It supports OS X, Linux/BSD (dbus) and Windows.

```yaml
my_value:
  type: keyring
  keyring:
    key: "my_key"
```
