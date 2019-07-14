# serviceman

A cross-platform service manager

```bash
serviceman add --name "my-project" node ./serve.js --port 3000
serviceman stop my-project
serviceman start my-project
```

Works with launchd (Mac), systemd (Linux), or standalone (Windows).

## Meta Package

This is a meta-package to fetch and install the correction version of
[go-serviceman](https://git.rootprojects.org/root/serviceman)
for your architecture and platform.

```bash
npm install serviceman
```

## How does it work?

1. Resolves executable from PATH, or hashbang (ex: `#!/usr/bin/env node`)
2. Resolves file and directory paths to absolute paths (ex: `/Users/me/my-project/serve.js`)
3. Creates a template `.plist` (Mac), `.service` (Linux), or `.json` (Windows) file
4. Calls `launchd` (Mac), `systemd` (Linux), or `serviceman-runner` (Windows) to enable/start/stop/etc
