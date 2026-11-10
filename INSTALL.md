# Installation on reMarkable 2

Use the AppLoad bundle for a direct install or the signed local APK if you
already use Vellum. Both downloads are on the [v0.1.8 release page](https://github.com/Ampleforth/remarkable2-nextcloudCalendar/releases/tag/v0.1.8).

## 1. Prepare the tablet and computer

You need a reMarkable 2, root SSH access to the tablet, and an HTTPS Nextcloud
account with at least one calendar or task list. Connect the tablet to Wi-Fi
for sync. Run computer-side commands in a Linux or macOS terminal, or in
Ubuntu under WSL on Windows. The installer uses Bash, OpenSSH (`ssh` and
`scp`), and `tar`.

Install [XOVI and its Qt resource rebuilder](https://github.com/asivery/xovi)
and [ARM32 AppLoad](https://github.com/asivery/rm-appload) builds that match
your tablet firmware. Follow their upstream setup instructions first. Confirm that AppLoad opens before continuing. The app was
tested with RM2 firmware 3.27.3.0, Qt 6.8.2 and AppLoad 0.5.3.

Set up SSH access to the tablet. The commands below use `rm2` as an example SSH
alias; replace it with your own alias or `root@TABLET_ADDRESS`. No particular
network address, server, jump host or SSH key is required by this app. If you
use a USB connection, the tablet normally uses `10.11.99.1` over USB.

From your computer, verify the connection before installing:

```sh
ssh rm2 'uname -m && id -u && test -f /home/root/xovi/extensions.d/appload.so'
```

Expect `armv7l` and `0` on separate lines, followed by a successful exit.
Check your SSH target and AppLoad setup if the command fails.

## 2. Verify and unpack the download

Download the ARMv7 archive and `SHA256SUMS` from the same release. On a computer
with `sha256sum` and `tar`, in the download directory:

```sh
sha256sum -c SHA256SUMS
tar -xzf remarkable2-nextcloudCalendar-v0.1.8-rm2-armv7.tar.gz
cd remarkable2-nextcloudCalendar-v0.1.8-rm2-armv7
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

1. Sign in to Nextcloud in a browser. In your account's **Personal settings →
   Security**, find the devices and sessions section. Create an app password
   named **reMarkable Calendar** and copy it when Nextcloud displays it.
   [Nextcloud session management](https://docs.nextcloud.com/server/latest/user_manual/en/session_management.html)
2. In **Nextcloud Calendar → Settings**, enter your Nextcloud HTTPS base URL,
   including any installation subpath, and your Nextcloud username. For example,
   use `https://cloud.example.com/nextcloud` if that is where you sign in.
3. Enter the app password, tap **Save & sync**, and wait for sync to finish.
   Enable the calendars and task lists you want to see in Settings.
4. In a writable test calendar, create a test event. Tap **Sync** and confirm it
   appears in Nextcloud. If you use Tasks, repeat with a test task. Saving on the
   tablet queues the change. **Sync** sends the queued changes to Nextcloud.

Credentials are saved only on the tablet, at:

```text
/home/root/.config/nextcloud-calendar/config.json
```

The app creates this file with permissions `600`. A saved password is represented
by a placeholder in Settings; it is never returned by the connection API.

## Install a local Vellum package

This is a locally signed package for the reMarkable 2. It requires `rm2` and
`appload>=0.5.3` in Vellum. Install Vellum and compatible AppLoad packages
using [Vellum's setup guide](https://github.com/vellum-dev/vellum). reManager
can set up Vellum and install its dependencies. The local APK install runs in
the tablet terminal. [reManager](https://github.com/rmitchellscott/reManager)

1. Download `nextcloud-calendar-0.1.8-r0.apk`, its `.rsa.pub` public key, and
   `VELLUM-SHA256SUMS` from the same release. Verify with
   `sha256sum -c VELLUM-SHA256SUMS` in your downloads folder.
2. Close Nextcloud Calendar. If you previously used the manual installer, move
   `/home/root/xovi/exthome/appload/nextcloud-calendar` to a backup folder
   outside AppLoad before switching to Vellum. Keep your config and state files.
3. Transfer the APK and its public key to `/home/root/` on the tablet using SSH
   file transfer or reManager's file browser.
4. In the tablet terminal, replace `PACKAGE-KEY.rsa.pub` with the downloaded key
   filename and run:

   ```sh
   cp /home/root/PACKAGE-KEY.rsa.pub /home/root/.vellum/etc/apk/keys/
   /home/root/.vellum/bin/vellum add /home/root/nextcloud-calendar-0.1.8-r0.apk
   ```

5. Reload AppLoad, open Nextcloud Calendar, and connect your account as above.
   Update by installing a newer signed APK. Remove it with
   `vellum del nextcloud-calendar`.

The APK signature and bundle contents were checked locally. The manual bundle
was tested on the tablet. Catalog installation requires a separate Vellum
submission and review. Once listed, the same catalog entry serves reManager.

## Build and install from source

On your computer, install Git, [Go 1.25 or later](https://go.dev/dl/), Qt 6
resource tools, librsvg's `rsvg-convert`, and Python 3. For example, on Ubuntu
or Ubuntu under WSL, install the non-Go build tools with:

```sh
sudo apt update
sudo apt install git openssh-client qt6-base-dev-tools librsvg2-bin python3
```

Clone the release tag, confirm that `rcc` is from Qt 6, and build:

```sh
git clone --branch v0.1.8 https://github.com/Ampleforth/remarkable2-nextcloudCalendar.git
cd remarkable2-nextcloudCalendar
go version
/usr/lib/qt6/libexec/rcc --version
RCC=/usr/lib/qt6/libexec/rcc ./scripts/build.sh
./scripts/package.sh 0.1.8
```

If Qt is installed elsewhere, set `RCC` to the actual Qt 6 compiler path.
The build runs Go tests and cross-compiles the static ARMv7 backend. Check
`dist/nextcloud-calendar/` for the manifest, icon, resources, and
`backend/entry`. The `release/` directory contains the archive and checksums.
Follow the archive installation steps above to deploy it.

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

Download the new archive and checksum, close the running app, and repeat the
install and reload steps. Keep your backup until the new version is
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

- **SSH times out:** check the tablet is awake, verify the address or alias, and
  try USB if Wi-Fi SSH is unavailable.
- **Checksum mismatch:** download the archive and checksum again from the same release.
- **App opens blank or still shows the old UI:** close AppLoad apps and tap
  **Reload**. Verify the bundle has the four expected files and executable backend.
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
