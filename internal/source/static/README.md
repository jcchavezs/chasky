## static

`static` is the simples source, it allows you to specify the value literally:

```yaml
my_value:
  type: static
  static:
    value: "josecarlos-chavez_atko"
```

It is useful to specify `variables`, values that are not secrets but still relevant as credentials, for example the user e-mail. It is also useful for testing purposes.
