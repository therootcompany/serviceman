# go-serviceman

Cross-platform service management made easy.

> sudo serviceman add --name foo ./serve.js --port 3000

> Success: "foo" started as a "launchd" SYSTEM service, running as "root"

## Why?

Because it sucks to debug launchctl, systemd, etc.

Also, I wanted a reasonable way to install [Telebit](https://telebit.io) on Windows.
(see more in the **More Why** section below)

## Features

-   Unprivileged (User Mode) Services with `--user` (_Default_)
    -   [x] Linux (`sytemctl --user`)
    -   [x] MacOS (`launchctl`)
    -   [x] Windows (`HKEY_CURRENT_USER/.../Run`)
-   Privileged (System) Services with `--system` (_Default_ for `root`)
    -   [x] Linux (`sudo sytemctl`)
    -   [x] MacOS (`sudo launchctl`)
    -   [ ] Windows (_not yet implemented_)

# Table of Contents

-   Usage
-   Install
-   Examples
    -   compiled programs
    -   scripts
    -   bash
    -   node
    -   python
    -   ruby
-   Logging
-   Debugging
-   Windows
-   Building
-   More Why
-   Legal

# Usage

The basic pattern of usage:

```bash
sudo serviceman add --name "foobar" [options] [interpreter] <service> [--] [service options]
sudo serviceman start <service>
sudo serviceman stop <service>
serviceman version
```

And what that might look like:

```bash
sudo serviceman add --name "foo" foo.exe -c ./config.json
```

You can also view the help:

```
serviceman add --help
```

# System Services VS User Mode Services

User services start **on login**.

System services start **on boot**.

The **default** is to register a _user_ services. To register a _system_ service, use `sudo` or run as `root`.

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
move serviceman.exe %userprofile%\bin\serviceman.exe
reg add HKEY_CURRENT_USER\Environment /v PATH /d "%PATH%;%userprofile%\bin"
```

**All Others**

```
chmod a+x ./serviceman
sudo mv ./serviceman /usr/local/bin/
```

# Examples

```bash
sudo serviceman add --name <name> <program> [options] [--] [raw options]

# Example
sudo serviceman add --name "gizmo" gizmo --foo bar/baz
```

Anything that looks like file or directory will be **resolved to its absolute path**:

```bash
# Example of path resolution
gizmo --foo /User/me/gizmo/bar/baz
```

Use `--` to prevent this behavior:

```bash
# Complex Example
sudo serviceman add --name "gizmo" gizmo -c ./config.ini -- --separator .
```

For native **Windows** programs that use `/` for flags, you'll need to resolve some paths yourself:

```bash
# Windows Example
serviceman add --name "gizmo" gizmo.exe .\input.txt -- /c \User\me\gizmo\config.ini /q /s .
```

In this case `./config.ini` would still be resolved (before `--`), but `.` would not (after `--`)

<details>
<summary>Compiled Programs</summary>

Normally you might your program somewhat like this:

```bash
gizmo run --port 8421 --config envs/prod.ini
```

Adding a service for that program with `serviceman` would look like this:

```bash
sudo serviceman add --name "gizmo" gizmo run --port 8421 --config envs/prod.ini
```

serviceman will find `gizmo` in your PATH and resolve `envs/prod.ini` to its absolute path.

</details>

<details>
<summary>Using with scripts</summary>

```bash
./snarfblat.sh --port 8421
```

Although your text script may be executable, you'll need to specify the interpreter
in order for `serviceman` to configure the service correctly.

This can be done in two ways:

1. Put a **hashbang** in your script, such as `#!/bin/bash`.
2. Prepend the **interpreter** explicitly to your command, such as `bash ./dinglehopper.sh`.

For example, suppose you had a script like this:

`iamok.sh`:

```bash
while true; do
  sleep 1; echo "Still Alive, Still Alive!"
done
```

Normally you would run the script like this:

```bash
./imok.sh
```

So you'd either need to modify the script to include a hashbang:

```bash
#!/usr/bin/env bash
while true; do
  sleep 1; echo "I'm Ok!"
done
```

Or you'd need to prepend it with `bash` when creating a service for it:

```bash
sudo serviceman add --name "imok" bash ./imok.sh
```

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

Serviceman understands all 3 of those approaches.

</details>

<details>
<summary>Using with node.js</summary>

If normally you run your node script something like this:

```bash
pushd ~/my-node-project/
npm start
```

Then you would add it as a system service like this:

```bash
sudo serviceman add npm start
```

If normally you run your node script something like this:

```bash
pushd ~/my-node-project/
node ./serve.js --foo bar --baz
```

Then you would add it as a system service like this:

```bash
sudo serviceman add node ./serve.js --foo bar --baz
```

It's important that any paths start with `./` and have the `.js`
so that serviceman knows to resolve the full path.

```bash
# Bad Examples
sudo serviceman add node ./demo # Wouldn't work for 'demo.js' - not a real filename
sudo serviceman add node demo   # Wouldn't work for './demo/' - doesn't look like a directory
```

See **Using with scripts** for more detailed information.

</details>

<details>
<summary>Using with python</summary>

If normally you run your python script something like this:

```bash
pushd ~/my-python-project/
python ./serve.py --config ./config.ini
```

Then you would add it as a system service like this:

```bash
sudo serviceman add python ./serve.py --config ./config.ini
```

See **Using with scripts** for more detailed information.

</details>

<details>
<summary>Using with ruby</summary>

If normally you run your ruby script something like this:

```bash
pushd ~/my-ruby-project/
ruby ./serve.rb --config ./config.yaml
```

Then you would add it as a system service like this:

```bash
sudo serviceman add ruby ./serve.rb --config ./config.yaml
```

See **Using with scripts** for more detailed information.

</details>

## Hints

-   If something goes wrong, read the output **completely** - it'll probably be helpful
-   Run `serviceman` from your **project directory**, just as you would run it normally
    -   Otherwise specify `--name <service-name>` and `--workdir <project directory>`
-   Use `--` in front of arguments that should not be resolved as paths
    -   This also holds true if you need `--` as an argument, such as `-- --foo -- --bar`

```
# Example of a / that isn't a path
# (it needs to be escaped with --)
sudo serviceman add dinglehopper config/prod -- --category color/blue
```

# Logging

### Linux

```bash
sudo journalctl -xef --unit <NAME>
sudo journalctl -xef --user-unit <NAME>
```

### Mac, Windows

When you run `serviceman add` it will either give you an error or
will print out the location where logs will be found.

By default it's one of these:

```txt
~/.local/share/<NAME>/var/log/<NAME>.log
```

```txt
/opt/<NAME>/var/log/<NAME>.log
```

You set it with one of these:

-   `--logdir <path>` (cli)
-   `"logdir": "<path>"` (json)
-   `Logdir: "<path>"` (go)

If anything about the logging sucks, tell me... unless they're your logs
(which they probably are), in which case _you_ should fix them.

That said, my goal is that it shouldn't take an IT genius to interpret
why your app failed to start.

# Debugging

-   `serviceman add --dryrun <normal options>`
-   `serviceman run --config <special config>`

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

# More Why

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
<!-- {{ end }} -->
