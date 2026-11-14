# Install Nextcloud Calendar 0.2.0

Download the portable archive for your tablet from
[GitHub Releases](https://github.com/Ampleforth/remarkable2-nextcloudCalendar/releases/latest).
The archives include the AppLoad app, installer, matching source, docs and licenses.

| Tablet | Archive suffix | Status |
| --- | --- | --- |
| reMarkable 1 / 2 | `rm1-rm2-armv7` | RM2 tested on-device; RM1 confirmed working by the user |
| Paper Pro / Paper Pro Move | `paper-pro-aarch64` | Experimental; built and tested off-device, no physical hardware validation |

Use the matching architecture: the ARMv7 binary does not run on Paper Pro/Move.
Color support is detected from known Paper Pro/Move model identifiers; RM1, RM2
and unknown devices retain the original grayscale appearance. Matching CPU
architecture alone does not establish AppLoad, firmware or display compatibility.

## Requirements

- Working XOVI, Qt resource rebuilder, and AppLoad for your tablet and firmware.
  The tested RM2 setup uses firmware **3.27.3.0**, Qt **6.8.2**, AppLoad **0.5.3**.
- Root SSH access for installation and Wi-Fi for sync.
- An HTTPS Nextcloud server with Calendar/Tasks and a dedicated app password.

Install XOVI/AppLoad using their upstream instructions:
[XOVI](https://github.com/asivery/xovi),
[AppLoad](https://github.com/asivery/rm-appload).
This release does not replace firmware, XOVI, AppLoad or system libraries.
The older Vellum recipe in `packaging/vellum/VELBUILD` remains pinned to 0.1.8;
use the portable 0.2.0 archive for the features documented here.

## Portable installation

1. Download the correct archive and `SHA256SUMS`; verify its checksum.
2. Extract it and enter its directory, for example:

   ```sh
   tar -xzf remarkable2-nextcloudCalendar-v0.2.0-rm1-rm2-armv7.tar.gz
   cd remarkable2-nextcloudCalendar-v0.2.0-rm1-rm2-armv7
   ```

3. Close Nextcloud Calendar on the tablet, then run:

   ```sh
   ./install.sh SSH_ALIAS_OR_ROOT_AT_TABLET_ADDRESS
   ```

   The installer checks the bundle's architecture against the tablet, requires root
   and AppLoad, backs up the previous app and cache, and installs the new bundle.
   It retains your credentials, pending edits, captures, and display preferences.

4. Close other AppLoad apps, tap AppLoad Reload, and open **Nextcloud Calendar**.
   Optional hooks below need a separate XOVI reload.

On macOS, use `shasum -a 256` instead of `sha256sum`. An aarch64 archive has the
same `nextcloud-calendar/` directory inside; follow the same installation steps.
Do not install an experimental bundle over a working setup without keeping the
backup printed by the installer.

## Connect Nextcloud

1. In Nextcloud, create an app password in Settings → Security → Devices & sessions.
2. Open **Nextcloud Calendar → Settings** and enter the HTTPS server base URL,
   username and app password. Use the base URL, not an individual CalDAV resource.
3. Tap **Save & sync** and select the calendars/task lists you want to display.

A blank password retains the saved password only for the same server and username.
Account changes require resolving pending edits first and clear downloaded data
and account-specific colors. Configuration and cache files are private (`600`),
but stored without encryption. See [SECURITY.md](SECURITY.md).

## Color devices and calendar/list colors

On recognized Paper Pro/Move models, Settings shows **Calendar & task list colors**.
Tap a calendar/list and select its color. The choice persists locally and colors
its timed event blocks, all-day blocks, month labels, and task cards. Selected
controls use blue accents. The chosen color-device logo has a charcoal calendar,
multicolor date squares, and a blue Nextcloud symbol. These settings do not change
Nextcloud's server-side colors. Names and text remain readable independent of color.

RM1/RM2 keep the existing monochrome logo, event fills, task cards and controls;
color settings are hidden. Missing/unknown device identity defaults to grayscale.
Physical color-panel rendering and refresh behavior have not been validated.

## Optional lasso-to-task capture (firmware 3.27.x)

This feature adapts [jdkruzr's reTaskable](https://github.com/jdkruzr/reTaskable)
handwriting capture hook. It uses reMarkable's existing recognition flow without
modifying the notebook. Recognition requires Wi-Fi and a signed-in reMarkable
account; the dialog offers manual text entry if recognition is unavailable.

1. Open Nextcloud Calendar once to create the private intake directory.
2. Copy the bundled `nextcloud-calendar/xovi/nextcloudCalendarSend.qmd` to
   `/home/root/xovi/exthome/qt-resource-rebuilder/` and reload XOVI using your
   normal device procedure. Avoid enabling reTaskable's capture hook alongside it:
   both patch the same selection handlers. Do not enable on other firmware lines.
3. Lasso handwriting, tap the new checkmark, review the text, and tap Create.
   Repeat to queue multiple captures.
4. Open Nextcloud Calendar. **Tasks** stays at the far right. Immediately to its
   left, **Import Lasso Task** appears for one capture, or **Import 3 Lasso Tasks**
   for three, for example. It is hidden when no captures are waiting.
5. Tap Import. The editor shows progress such as **Lasso Task 1 of 3**. Choose
   a list, adjust dates/text, and Save to open the next capture automatically.
   **Discard** dismisses that capture without creating/deleting a calendar task
   or changing the notebook, then opens the next. **Cancel** stops the sequence
   and leaves the current and remaining captures waiting.
6. Tap **Sync** once to upload all saved tasks.

Source notebook name, document UUID and page index are kept in Description.
Jump-back navigation and background sync are not included. Captures remain in
`/home/root/.local/share/nextcloud-calendar/intake/`; a durable consumed-ID ledger
prevents saved or discarded captures from returning after restart. Unsaved captures
are not account-bound: review the active account and destination before saving.
Removing only `nextcloudCalendarSend.qmd` and reloading XOVI removes the hook.
The ARMv7 app's RM1 confirmation does not establish separate RM1 hook compatibility.

## Optional meeting-notes icon shortcut (firmware 3.27.3.0)

Copy `nextcloud-calendar/xovi/nextcloudCalendarLauncher.qmd` into the same resource
rebuilder directory, close the app, and reload XOVI once. The home-screen calendar
icon then opens Nextcloud Calendar. This hook is restricted to firmware 3.27.3.0.

In Settings, **Repurpose Meeting Notes Calendar Icon to open Nextcloud Calendar**
switches between Nextcloud Calendar and the original meeting-note behavior.
The default is enabled. Changes persist and apply within roughly one second,
without another restart. The hook reads only the non-secret `launcher.json`
preference and works while the app is closed. Disabled restores the original
meeting-note action and its visibility rules. Remove only the launcher QMD and
reload XOVI to remove the shortcut entirely. Hook behavior on color hardware has
not been physically verified.

## Timezone and manual configuration

Times default to UTC. Close the app before changing
`/home/root/.config/nextcloud-calendar/config.json`. Add an IANA timezone such as
`"timezone": "America/Phoenix"`, keep permissions `600`, then reopen the app.
See `config.example.json` for the other fields. Never commit this private file.
The 12/24-hour display setting is available in the app.

## Build from source

Requirements: Go 1.25+, Qt 6 `rcc`, librsvg's `rsvg-convert`; Python 3 for packaging
checks and screenshot helpers. QML UI tests use `qmltestrunner`.

```sh
git clone --branch v0.2.0 https://github.com/Ampleforth/remarkable2-nextcloudCalendar.git
cd remarkable2-nextcloudCalendar
./scripts/build.sh
./scripts/package.sh 0.2.0 arm
CALENDAR_GOARCH=arm64 ./scripts/build.sh
./scripts/package.sh 0.2.0 arm64
```

Set `GO`/`RCC` to executable paths if necessary. Builds are static (`CGO_ENABLED=0`)
Linux binaries with trimmed paths. ARMv7 output is `dist/nextcloud-calendar/`;
aarch64 output is `dist/nextcloud-calendar-arm64/`. Packaging produces separate
architecture archives, corresponding source, per-archive checksums and `SHA256SUMS`.
Do not infer tablet validation from an aarch64 build or off-device test pass.

## Updates, backups and troubleshooting

Close the app before updating. The installer prints the backup directory under
`/home/root/.local/share/nextcloud-calendar/backups/`; retain it until verified.
To restore, close the app, restore the previous bundle from that backup, and reload
AppLoad. If rolling back a hook, restore/remove its QMD and reload XOVI too.

- Cannot sync: check Wi-Fi, HTTPS base URL, username and app password.
- No writable list: Sync, then select an account collection supporting tasks/events.
- Conflict: inspect Pending changes in Settings. Retry ordinary errors; discard
  local changes to accept server data when a conflict requires it.
- Import missing: the button appears only with waiting captures. Confirm hook
  installation, firmware, and that capture creation finished.
- Color absent: RM1/RM2 intentionally remain monochrome; unknown models fall back
  to grayscale. Color hardware support remains experimental.
- Bad architecture: use the matching archive; do not rename binaries to bypass checks.

See [development and tests](docs/DEVELOPMENT.md) for validation and limitations.
