# go-serviceman

A cross-platform service manager.

Because debugging launchctl, systemd, etc absolutely sucks!

...and I wanted a reasonable way to install [Telebit](https://telebit.io) on Windows.
(see more in the **Why** section below)

## Features

- Unprivileged (User Mode) Services
  - [x] Linux (`sytemctl --user`)
  - [x] MacOS (`launchctl`)
  - [x] Windows (`HKEY_CURRENT_USER/.../Run`)
- Privileged (System) Services
  - [x] Linux (`sudo sytemctl`)
  - [x] MacOS (`sudo launchctl`)
  - [ ] Windows (_not yet implemented_)

# Table of Contents

- Usage
- Install
- Examples
  - compiled programs
  - scripts
  - bash
  - node
  - python
  - ruby
- Logging
- Debugging
- Windows
- Building
- Why
- Legal

# Usage

The basic pattern of usage, and what that might look like:

```
serviceman add [options] [interpreter] <service> -- [service options]
```

```
serviceman add foo.exe
```

```
serviceman add --title "Foo App" node ./foo.js -- --bar
```

You can also view the help and the version:

```
serviceman add --help
```

```
serviceman version
```

# Install

There are a number of pre-built binaries.

If none of them work for you, or you prefer to build from source,
see the instructions for building far down below.

## Downloads

### MacOS

MacOS (darwin): [64-bit Download ](https://rootprojects.org/serviceman/dist/darwin/amd64/serviceman)

```
curl https://rootprojects.org/serviceman/dist/darwin/amd64/serviceman -o serviceman
```

### Windows

<details>
<summary>See download options</summary>
Windows 10: [64-bit Download](https://rootprojects.org/serviceman/dist/windows/amd64/serviceman.exe)

```
powershell.exe $ProgressPreference = 'SilentlyContinue'; Invoke-WebRequest https://rootprojects.org/serviceman/dist/windows/amd64/serviceman.exe -OutFile serviceman.exe
```

**Debug version**:

```
powershell.exe $ProgressPreference = 'SilentlyContinue'; Invoke-WebRequest https://rootprojects.org/serviceman/dist/windows/amd64/serviceman.debug.exe -OutFile serviceman.debug.exe
```

Windows 7: [32-bit Download](https://rootprojects.org/serviceman/dist/windows/386/serviceman.exe)

```
powershell.exe "(New-Object Net.WebClient).DownloadFile('https://rootprojects.org/serviceman/dist/windows/386/serviceman.exe', 'serviceman.exe')"
```

**Debug version**:

```
powershell.exe "(New-Object Net.WebClient).DownloadFile('https://rootprojects.org/serviceman/dist/windows/386/serviceman.debug.exe', 'serviceman.debug.exe')"
```

</details>

### Linux

<details>
<summary>See download options</summary>

Linux (64-bit): [Download](https://rootprojects.org/serviceman/dist/linux/amd64/serviceman)

```
curl https://rootprojects.org/serviceman/dist/linux/amd64/serviceman -o serviceman
```

Linux (32-bit): [Download](https://rootprojects.org/serviceman/dist/linux/386/serviceman)

```
curl https://rootprojects.org/serviceman/dist/linux/386/serviceman -o serviceman
```

</details>

### Raspberry Pi (Linux ARM)

<details>
<summary>See download options</summary>

RPi 4 (64-bit armv8): [Download](https://rootprojects.org/serviceman/dist/linux/armv8/serviceman)

```
curl https://rootprojects.org/serviceman/dist/linux/armv8/serviceman -o serviceman`
```

RPi 3 (armv7): [Download](https://rootprojects.org/serviceman/dist/linux/armv7/serviceman)

```
curl https://rootprojects.org/serviceman/dist/linux/armv7/serviceman -o serviceman
```

ARMv6: [Download](https://rootprojects.org/serviceman/dist/linux/armv6/serviceman)

```
curl https://rootprojects.org/serviceman/dist/linux/armv6/serviceman -o serviceman
```

RPi Zero (armv5): [Download](https://rootprojects.org/serviceman/dist/linux/armv5/serviceman)

```
curl https://rootprojects.org/serviceman/dist/linux/armv5/serviceman -o serviceman
```

</details>

### Add to PATH

**Windows**

```
mkdir %userprofile%\bin
reg add HKEY_CURRENT_USER\Environment /v PATH /d "%PATH%;%userprofile%\bin"
move serviceman.exe %userprofile%\bin\serviceman.exe
```

**All Others**

```
sudo mv ./serviceman /usr/local/bin/
```

# Examples

> **serviceman add** &lt;program> **--** &lt;program options>

<details>
<summary>Compiled Programs</summary>

Normally you might your program somewhat like this:

```
dinglehopper --port 8421
```

Adding a service for that program with `serviceman` would look like this:

> **serviceman add** dinglehopper **--** --port 8421

serviceman will find dinglehopper in your PATH.

</details>

<details>
<summary>Using with scripts</summary>

Although your text script may be executable, you'll need to specify the interpreter
in order for `serviceman` to configure the service correctly.

For example, if you had a bash script that you normally ran like this:

```
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

```
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

## Relative vs Absolute Paths

Although serviceman can expand the executable's path,
if you have any arguments with relative paths
you should switch to using absolute paths.

```
dinglehopper --config ./conf.json
```

```
serviceman add dinglehopper -- --config /Users/me/dinglehopper/conf.json
```

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
	"exec": "/Users/me/go-demo/demo",
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

## Peculiarities of Windows

# Console vs No Console

Windows binaries can be built either for the console or the GUI.

When they're built for the console they can hide themselves when they start.
They must open up a terminal window.

When they're built for the GUI they can't print any output - even if they're started in the terminal.

This is why there's a **Debug version** for the windows binaries -
so that you can get your arguments correct with the one and then
switch to the other.

There's probably a clever way to work around this, but I don't know what it is yet.

# No userspace launcher

Windows doesn't have a userspace daemon launcher.
This means that if your application crashes, it won't automatically restart.

However, `serviceman` handles this by not directly adding your application
to `HKEY_CURRENT_USER/.../Run`, but rather installing a copy of _itself_
instead, which runs your application and automatically restarts it whenever it
exits.

If the application fails to start `serviceman` will retry continually,
but it does have an exponential backoff of up to 1 minute between failed
restart attempts.

See the bit on `serviceman run` in the **Debugging** section up above for more information.

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
