# Development

## Architecture

The Go backend runs as AppLoad's `backend/entry`. It binds an HTTP API to loopback
port 38766 and connects to the Unix SEQPACKET lifecycle socket supplied by AppLoad.
The Qt Quick frontend communicates through JSON. Closing the lifecycle connection
or receiving AppLoad's termination packet exits the backend.

CalDAV discovery follows the current-user principal, calendar home and collection
properties. Event queries use a bounded date window; task queries are unbounded
so undated/overdue tasks remain available. Raw iCalendar resources and ETags are
retained. Offline edits overlay the cache and are persisted before acknowledgment.
Creates use stable URLs with `If-None-Match: *`; updates/deletions use `If-Match`.

The cache is atomically replaced as a complete JSON document. Account identity,
display preferences, selected calendars, and the outbox persist across restarts.
Account switches are blocked while edits are pending. Successful account switches
clear the previous account's downloaded data before using the new connection.

## Source layout

- `main.go`: HTTP routes, process setup, AppLoad lifecycle.
- `caldav.go`, `sync.go`: discovery, resource downloads, conditional writes.
- `items.go`: iCalendar parsing, event expansion, task metadata and local edits.
- `connection.go`, `model.go`: private configuration and persistent state.
- `qml/`: UI, timeline geometry, gestures, keyboard, and dropdowns.
- `scripts/`: build, packaging, optional browser login, install and privacy checks.
- `examples/`: fictional read-only desktop demo; no account is contacted.

## Tests

```sh
go test -race ./...
go vet ./...
QT_QPA_PLATFORM=offscreen QT_QPA_PLATFORMTHEME=generic QT_QUICK_BACKEND=software \
  qmltestrunner -input tests/qml -import tests/qml/imports
go build -o /tmp/nextcloud-calendar-host .
python3 tests/lifecycle.py /tmp/nextcloud-calendar-host
python3 scripts/audit.py .
```

QML tests use a test-only AppLoad stub. These tests do not establish compatibility
with every tablet firmware or replace physical touch testing.

The live write test is skipped by default. To run it deliberately, set
`CALENDAR_LIVE_CONFIG` to a private config file and `CALENDAR_LIVE_COLLECTION` to a
writable test collection URL that supports both events and tasks. Then run:

```sh
go test -run '^TestLiveRoundtrip$' -v
```

It creates uniquely identified temporary resources and attempts cleanup. Inspect
any cleanup failure before running it again. Never commit the private config.

## Fictional screenshots

`examples/Demo.qml` overrides all backend requests with fictional fixtures. Render
it with the test import path and the software/offscreen Qt platform. See
`scripts/screenshots.sh`. The app's runtime bundle does not include these fixtures.

## Packaging

```sh
./scripts/build.sh
./scripts/package.sh 0.1.0
```

The archive contains only the ARMv7 app bundle, installer, docs, example config,
screenshots and license notices. Configuration, state, caches, editor backups,
Git history, and build-host paths are excluded. The complete corresponding source
is this repository; include it or provide the same-version source when redistributing
a binary under the GPL.

## Limitations

- Portrait grayscale UI; no event drag/resize.
- Manual sync; no always-on scheduled polling of Nextcloud.
- Recurring resources and invitations are read-only. RDATE-only recurrence,
  THISANDFUTURE exceptions, embedded custom timezone definitions and unusual
  multi-component resources need additional compatibility work.
- Recurrence expansion is capped at 200,000 iterations and 5,000 visible instances
  per resource per requested range. Server responses are capped at 32 MiB.
- The cache is a single-process store. Do not launch a second standalone backend
  with the same cache while the AppLoad app is running.
- Large all-day lists expand the all-day section and leave less space for the grid.
- File-based timezone changes require closing/reopening the app.
- AppLoad 0.5.3 can log an upstream null-unloading warning when a frontend closes,
  even though the backend exits normally.
