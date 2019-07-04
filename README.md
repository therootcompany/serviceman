# go-serviceman

A cross-platform service manager.

Because debugging launchctl, systemd, etc absolutely sucks!

...and I wanted a reasonable way to install [Telebit](https://telebit.io) on Windows.
(see more in the **Why** section below)

<details>
<summary>User Mode Services</summary>
  * `sytemctl --user` on Linux
  * `launchctl` on MacOS
  * `HKEY_CURRENT_USER/.../Run` on Windows
</details>
<details>
<summary>System Services</summary>
  * `sudo sytemctl` on Linux
  * `sudo launchctl` on MacOS
  * _not yet implemented_ on Windows
</details>

## Contents

- Install
- Usage
- Build
- Examples
  - compiled programs
  - scripts
  - bash
  - node
  - python
  - ruby
- Logging
- Windows
- Debugging
- Why
- Legal

# Install

Download `serviceman` for

- [MacOS (64-bit darwin)](https://rootprojects.org/serviceman/dist/darwin/amd64/serviceman)
- [Windows 10 (64-bit)](https://rootprojects.org/serviceman/dist/windows/amd64/serviceman.exe)
- [Windows 10 (32-bit)](https://rootprojects.org/serviceman/dist/windows/386/serviceman.exe)
- [Linux (64-bit)](https://rootprojects.org/serviceman/dist/linux/amd64/serviceman)
- [Linux (32-bit)](https://rootprojects.org/serviceman/dist/linux/386/serviceman)
- [Raspberry Pi 4 (64-bit armv8)](https://rootprojects.org/serviceman/dist/linux/armv8/serviceman)
- [Raspberry Pi 3 (armv7)](https://rootprojects.org/serviceman/dist/linux/armv7/serviceman)
- [Raspberry Pi 2 (armv6)](https://rootprojects.org/serviceman/dist/linux/armv6/serviceman)
- [Raspberry Pi Zero (armv5)](https://rootprojects.org/serviceman/dist/linux/armv5/serviceman)

# Usage

```bash
serviceman add [options] [interpreter] <service> -- [service options]
```

```bash
serviceman add --help
```

```bash
serviceman version
```

# Examples

**Compiled Apps**

Normally you might run your program something like this:

```bash
dinglehopper --port 8421
```

Adding a service for that program with `serviceman` would look like this:

> **serviceman add** dinglehopper **--** --port 8421

`serviceman` will find `dinglehopper` in your PATH, but if you have
any arguments with relative paths, you should switch to using absolute paths.

```bash
dinglehopper --config ./conf.json
```

becomes

> **serviceman add** dinglehopper **--** --config **/Users/aj/dinglehopper/conf.json**

<details>
<summary>Using with scripts</summary>

Although your text script may be executable, you'll need to specify the interpreter
in order for `serviceman` to configure the service correctly.

For example, if you had a bash script that you normally ran like this:

```bash
./snarfblat.sh --port 8421
```

You'd create a system service for it like this:

> serviceman add **bash** ./snarfblat.sh **--** --port 8421

`serviceman` will resolve `./snarfblat.sh` correctly because it comes
before the **--**.

**Background Information**

An operating system can't "run" text files (even if the executable bit is set).

Scripts require an _interpreter_. Often this is denoted at the top of
"executable" scripts with something like one of these:

```bash
#!/usr/bin/env ruby
```

```bash
#!/usr/bin/python
```

However, sometimes people get fancy and pass arguments to the interpreter,
like this:

```bash
#!/usr/local/bin/node --harmony --inspect
```

</details>

<details>
<summary>Using with node.js</summary>

If normally you run your node script something like this:

```bash
node ./demo.js --foo bar --baz
```

Then you would add it as a system service like this:

> **serviceman add** node ./demo.js **--** --foo bar --baz

It is important that you specify `node ./demo.js` and not just `./demo.js`

See **Using with scripts** for more detailed information.

</details>

<details>
<summary>Using with python</summary>

If normally you run your python script something like this:

```bash
python ./demo.py --foo bar --baz
```

Then you would add it as a system service like this:

> **serviceman add** python ./demo.py **--** --foo bar --baz

It is important that you specify `python ./demo.py` and not just `./demo.py`

See **Using with scripts** for more detailed information.

</details>

<details>
<summary>Using with ruby</summary>

If normally you run your ruby script something like this:

```bash
ruby ./demo.rb --foo bar --baz
```

Then you would add it as a system service like this:

> **serviceman add** ruby ./demo.rb **--** --foo bar --baz

It is important that you specify `ruby ./demo.rb` and not just `./demo.rb`

See **Using with scripts** for more detailed information.

</details>

# Logging

When you run `serviceman add` it will either give you an error or
will print out the location where logs will be found.

By default it's one of these:

```txt
~/.local/share/<NAME>/var/log/<NAME>.log
```

```txt
/var/log/<NAME>/var/log/<NAME>.log
```

You set it with one of these:

- `--logdir <path>` (cli)
- `"logdir": "<path>"` (json)
- `Logdir: "<path>"` (go)

If anything about the logging sucks, tell me... unless they're your logs
(which they probably are), in which case _you_ should fix them.

That said, my goal is that it shouldn't take an IT genius to interpret
why your app failed to start.

# Peculiarities of Windows

Windows doesn't have a userspace daemon launcher.
This means that if your application crashes, it won't automatically restart.

However, `serviceman` handles this by not directly adding your application
to `HKEY_CURRENT_USER/.../Run`, but rather installing a copy of _itself_
instead, which runs your application and automatically restarts it whenever it
exits.

If the application fails to start `serviceman` will retry continually,
but it does have an exponential backoff of up to 1 minute between failed
restart attempts.

See the bit on `serviceman run` in the **Debugging** section down below for more information.

# Debugging

One of the most irritating problems with all of these launchers is that they're
terrible to debug - it's often difficult to find the logs, and nearly impossible
to interpret them, if they exist at all.

The config files generate by `serviceman` are simple, template-generated and
tested, and therefore gauranteed to work - **_if_** your
application runs with the parameters given, which is big 'if'.

`serviceman` tries to make sure that all necessary files and folders
exist and give clear error messages if they don't (be sure to check the logs,
mentioned above).

There's also a `run` utility that can be used to test that the parameters
you've given are being interpreted correctly (absolute paths and such).

```bash
serviceman run --config ./conf.json
```

Where `conf.json` looks something like

**For Binaries**:

```json
{
	"title": "Demo",
	"exec": "/Users/aj/go-demo/demo",
	"argv": ["--foo", "bar", "--baz", "qux"]
}
```

**For Scripts**:

Scripts can't be run directly. They require a binary `interpreter` - bash, node, ruby, python, etc.

If you're running from the folder containing `./demo.js`,
and `node.exe` is in your PATH, then you can use executable
names and relative paths.

```json
{
	"title": "Demo",
	"interpreter": "node.exe",
	"exec": "./bin/demo.js",
	"argv": ["--foo", "bar", "--baz", "qux"]
}
```

That's equivalent to this:

```json
{
	"title": "Demo",

	"name": "demo",

	"exec": "node.exe",
	"argv": ["./bin/demo.js", "--foo", "bar", "--baz", "qux"]
}
```

Making `add` and `run` take the exact same arguments is on the TODO list.
The fact that they don't is an artifact of `run` being created specifically
for Windows.

If you have gripes about it, tell me. It shouldn't suck. That's the goal anyway.

# Building

```bash
git clone https://git.coolaj86.com/coolaj86/go-serviceman.git
```

```bash
pushd ./go-serviceman
```

```bash
go generate -mod=vendor ./...
```

**Windows**:

```bash
go build -mod=vendor -ldflags "-H=windowsgui" -o serviceman.exe
```

**Linux, MacOS**:

```bash
go build -mod=vendor -o /usr/local/bin/serviceman
```

# Why

I created this for two reasons:

1. Too often I just run services in `screen -xRS foo` because systemd `.service` files are way too hard to get right and even harder to debug. I make stupid typos or config mistakes and get it wrong. Then I get a notice 18 months later from digital ocean that NYC region 3 is being rebooted and to expect 5 seconds of downtime... and I don't remember if I remembered to go back and set up that service with systemd or not.
2. To make it easier for people to install [Telebit](https://telebit.io) on Windows.

<!-- {{ if .Legal }} -->

# Legal

[serviceman](https://git.coolaj86.com/coolaj86/go-serviceman) |
MPL-2.0 |
[Terms of Use](https://therootcompany.com/legal/#terms) |
[Privacy Policy](https://therootcompany.com/legal/#privacy)

Copyright 2019 AJ ONeal.

<!-- {{ end }} -->
