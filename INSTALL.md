# Installation on reMarkable 2

## 1. Prepare the tablet

Install compatible XOVI, Qt resource rebuilder, and ARM32 AppLoad builds using their
upstream instructions. Confirm that AppLoad opens before continuing. The app was
tested with RM2 firmware 3.27.3.0, Qt 6.8.2 and AppLoad 0.5.3.

Set up SSH access to the tablet. The commands below use `rm2` as an example SSH
alias; replace it with your own alias or `root@TABLET_ADDRESS`. No particular
network address, server, jump host or SSH key is required by this app.

## 2. Verify and unpack the download

Download the ARMv7 archive and `SHA256SUMS` from the same release. On a computer
with `sha256sum` and `tar`, in the download directory:

```sh
sha256sum -c SHA256SUMS
tar -xzf remarkable2-nextcloudCalendar-v0.1.0-rm2-armv7.tar.gz
cd remarkable2-nextcloudCalendar-v0.1.0-rm2-armv7
```

Use the version matching the archive you downloaded. On macOS, use
`shasum -a 256 ARCHIVE_NAME` and compare it to `SHA256SUMS`.

## 3. Install the bundle

Close Nextcloud Calendar on the tablet. Run the included installer from the
extracted folder:

```sh
./install.sh rm2
```

It verifies ARMv7 and AppLoad, stages the bundle, and backs up an existing app and
its offline state before replacing the app directory. It does not change saved
connection credentials, install system dependencies, or restart the tablet UI.
SSH authentication stays in your terminal; do not put passwords in arguments.

For manual installation, copy `nextcloud-calendar/` to:

```text
/home/root/xovi/exthome/appload/nextcloud-calendar/
```

The folder must contain `manifest.json`, `resources.rcc`, `icon.png`, and the
executable `backend/entry`. Keep the bundled license files with it. The backend
must have executable permissions. Close the app before replacing its files.

Open AppLoad and tap **Reload** after all AppLoad apps are closed. Then open
**Nextcloud Calendar**. If the launcher still shows an old icon, close and reopen
AppLoad; its image cache can retain an earlier icon until the tablet UI reloads.

## 4. Enter your connection

In Settings, enter your own HTTPS Nextcloud base URL, username and app password.
Generate the app password in your Nextcloud account's Security settings. Tap
**Save & sync**. Select the calendars and task lists to display.

Credentials are saved only on the tablet, at:

```text
/home/root/.config/nextcloud-calendar/config.json
```

The app creates this file with permissions `600`. A saved password is represented
by a placeholder in Settings; it is never returned by the connection API.

## Timezone and manual configuration

The default timezone is UTC. For another timezone, close the app, copy its config
to your computer, and change only the `timezone` value in a local editor. Use a
valid IANA name, for example `Europe/London`. Copy it back with permissions `600`,
then reopen the app. The existing password remains in this private file, so keep
it out of source control, screenshots, tickets and terminal output.

You can also start from `config.example.json`, replacing all placeholders with
your own settings. Keep the file private on your computer, create the tablet's
config directory with permissions `700`, and install the file with permissions
`600`. Avoid pasting credentials into shell command arguments or shared logs.

An optional browser authorization helper is included with the source repository:

```sh
python3 scripts/connect.py --server https://cloud.example.com --timezone UTC --output /private/path/config.json
```

It obtains a dedicated app password through Nextcloud login flow v2. It prints
only the browser authorization URL and saves credentials to the specified private
file, which you can then transfer to the tablet.

## Updates and removal

Use the same installer for updates. Keep your backup until the new version is
working. Backups are stored under:

```text
/home/root/.local/share/nextcloud-calendar/backups/
```

The cache/outbox is at `~/.local/share/nextcloud-calendar/state.json`. It may contain
unsynced edits. Preserve it before removing or downgrading the app. The installer
never restores or removes your credentials automatically.

To remove the app, close it, remove only its AppLoad bundle directory, then tap
AppLoad Reload. Data and credentials remain until you deliberately remove them.
If you stop using the app, revoke its app password in Nextcloud's Security settings.

## Troubleshooting

- **AppLoad missing:** install compatible XOVI/AppLoad dependencies first.
- **Wrong architecture:** this archive is ARMv7 for RM2, not ARM64.
- **Cannot connect:** verify Wi-Fi, the HTTPS base URL, username, app password,
  trusted TLS certificate chain, and access to the server from the tablet.
- **Read-only list:** the server's advertised CalDAV privileges are honored.
- **Unscheduled tasks:** set Scheduled for (`DTSTART`), or choose Due date sorting.
- **Write conflict:** inspect the pending change in Settings. Discard the local
  edit to keep the server version. A conflict never uses an unconditional overwrite.
- **Time offset:** set the correct IANA timezone and reopen the app.
- **UI errors after a firmware upgrade:** check upstream XOVI/AppLoad compatibility.
  Do not replace stock Qt/glibc libraries to make an incompatible extension load.

The app is an early community release. Desktop tests and one firmware combination
have been exercised; verify your own firmware/server combination using a test
calendar before relying on it for important edits.
