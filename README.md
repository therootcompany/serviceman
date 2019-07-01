# go-serviceman

A cross-platform service manager.

Goal:

```bash
serviceman install [options] [interpreter] <service> [-- [options]]
```

```bash
serviceman install --user ./foo-app -- -c ./
```

```bash
serviceman install --user /usr/local/bin/node ./whatever.js -- -c ./
```
