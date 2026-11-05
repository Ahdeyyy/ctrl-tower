# ctrl-tower

A small desktop app for your day: events, reminders and a focus timer in one
place. It lives in the system tray and runs on Windows, macOS and Linux.

## Features

- **Today and Upcoming** — your agenda for today and the days ahead.
- **Events and reminders** — with location, link and notes, and repeats
  (daily, weekdays, weekly, monthly). Edit or skip a single occurrence.
- **Quick add** — type plain text like `Design sync tomorrow 3pm 45m zoom.us/j/123`
  or `remind me to call the bank at 4`.
- **Alerts** — desktop notifications before things start, plus an overlay
  when a meeting begins, with a one-click Join for Zoom, Meet, Teams and others.
  Snooze or dismiss.
- **Focus timer** — Pomodoro-style sessions with presets, breaks and long
  breaks. Optional Do Not Disturb while focusing, and a heads-up when a
  session would run into a meeting.
- **Stats** — focus minutes per day, time by tag and your streak.
- **Tray and global shortcut** — click the tray icon or press
  `Ctrl+Alt+Space` (configurable) for a quick panel from any app.
- **Local data** — everything is stored in a single `data.json` in your
  config folder (Settings → Data shows where).

Keyboard shortcuts in the app: `Ctrl+N` new event, `Ctrl+Shift+N` new
reminder, `Ctrl+1`–`5` switch pages, `Space` start/pause the timer on Focus.

## Run locally

You need [Go](https://go.dev/dl/) 1.27 or later. The app is built with
[MyGo](https://github.com/egoist/mygo), which is pulled in as a Go tool, so
there is nothing else to install. On Linux you also need GTK 3.

```sh
git clone https://github.com/Ahdeyyy/ctrl-tower.git
cd ctrl-tower
go tool mygo dev
```

This starts the app and rebuilds it when you change the code. To start with
sample data, set `CTRL_TOWER_DEV=1` first.

Run the tests:

```sh
CGO_ENABLED=0 go test ./...
```

## Build

```sh
go tool mygo build
```

The app and its installer land in `build/<os>-<arch>/`. To build for other
platforms, pass `-platform`, for example:

```sh
go tool mygo build -platform windows/amd64,linux/amd64,darwin/universal
```

The Windows installer needs `makensis` (NSIS) on your PATH.

## Releases

Pushing a tag that matches the version in `mygo.json` (e.g. `v0.1.0`) builds
every platform and publishes a GitHub release.
