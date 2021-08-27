# evdev-trigger

`evdev-trigger` is a command-line tool that triggers commands by input events of evdev (Linux event devices).

- This will work even if the device is disconnected and reconnected.
- You can specify the interval between the next execution of the command.

## Installation

```bash
# clone repository
git clone https://github.com/hareku/evdev-trigger.git
cd evdev-trigger

# build (requires golang and build-essential)
make

# put the built binary to bin path
cp .build/evdev-trigger /usr/local/bin/evdev-trigger
```

## Usage

Create your configuration file like this.

```yaml
# physical id of device
phys: a1:b2:c3:d4:e5:f6
triggers:
  # Key is the input event code to trigger the command.
  115:
    # Command to execute.
    command: ["echo", "Hello"]
    # Optional, a minimum interval between the next execution of the command.
    # Value must be parsable as Golang time.Duration.
    interval: 3s
```

And you can start by `evdev-trigger --config /etc/evdev-trigger/myconf.yml --debug`.

In `--debug` mode, evdev-trigger displays the device connection status and input events to stdout.
If it's not in debug mode, only the results of the command execution will be displayed.
