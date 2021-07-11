# KRunner Backend for Caffeination

This application provides a [KRunner](https://docs.kde.org/stable5/en/plasma-desktop/plasma-desktop/krunner.html) backend that allows the user to "caffeinate" (inhibit sleep/locking) for a user specified length of time.

## Usage

In KRunner simply do `caffeinate` or `caff` followed by the length of time you'd like.

The format of time is a decimal number followed by a unit suffix. Units supported are "ns", "us", "ms", "s", "m", and "h".

For example

```sh
caffeinate 4h2m
```

or

```sh
caff 10s5m2ms
```

## How it works

The `.desktop` file placed in `~/.local/share/kservices5/krunner/dbusplugins` adds the plugin to KRunner as a DBus plugin, with details for the DBus path. We request this address in DBus and then listen in the Golang program. When anything is typed in KRunner it is sent (via DBus) to the `Match` method which checks if the syntax is correct for a valid request. If it is, we return a `Match` that will appear as an entry in KRunner. To parse the given duration we use Golang's built-in <code>[ParseDuration](https://golang.org/pkg/time/#ParseDuration)</code>.

When this entry is selected `Run` is called where we call `org.freedesktop.PowerManagement.Inhibit` and `org.freedesktop.ScreenSaver.Inhibit` (and then `.UnInhibit` after the given duration).

You can also add the `.service` file to `~/.local/share/dbus-1/services/` so that the application is autostarted, opposed to autostarting.
