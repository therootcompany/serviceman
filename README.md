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

```bash
serviceman run --config conf.json
```

```json
{
    "interpreter": "/Program Files (x86)/node/node.exe",
    "exec": "/Users/aj/demo/demo.js",
    "argv": ["--foo", "bar", "--baz", "qux"]
}
```

```bash
go generate -mod=vendor ./...
go build -mod=vendor -ldflags "-H=windowsgui"
.\\go-serviceman node ./demo.js -- --foo bar --baz qux
```