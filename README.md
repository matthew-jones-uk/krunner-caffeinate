# KRunner Backend for Caffeination

This application provides a Krunner backend that allows the user to "caffeinate" (inhibit sleep/locking) for a user specified length of time.

## Usage

In KRunner simply do `caffeinate` or `caff` followed by the length of time you'd like.

The format of time is a decimal number followed by a unit suffix. Units supported are "ns", "us", "ms", "s", "m", and "h".

For example

```
caffeinate 4h2m
```

or

```
caff 10s5m2ms
```
