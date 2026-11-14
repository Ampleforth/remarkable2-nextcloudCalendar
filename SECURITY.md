# Security and local data

- Connect only to your own trusted HTTPS Nextcloud installation. TLS validation
  remains enabled; redirects do not forward credentials to other origins.
- Use a dedicated Nextcloud app password, not the account's main password.
- The app password is stored as plaintext in a tablet-local file with permissions
  `600`. The app does not claim encrypted secret storage.
- Saved passwords are not loaded into QML or returned by the connection API.
  Password fields and the custom keyboard use masked display.
- The HTTP API binds only to loopback. Mutations require a custom request header;
  browser-origin requests are rejected. This is not a sandbox against other local
  processes running as the tablet user.
- Cached calendar/task data and pending edits are private user data. Do not share
  the state file, configuration, environment files, authentication links or backups.
- The distribution includes no account settings or cache. The example config and
  tests contain only explicit placeholders and dummy credentials.
- Publishing a private installation's config or adding secrets to Git history
  requires credential revocation; deleting the latest file does not erase history.

For a security report, contact the repository maintainer privately through GitHub.
Do not attach credentials, full cache files, or private event/task contents.

Lasso captures contain recognized handwriting and notebook metadata. Treat the
intake directory and consumed-ID ledger as private data. Discard dismisses a
capture; it does not securely erase its spool file. `launcher.json` contains only
the meeting-notes shortcut boolean. Calendar/list color preferences contain local
collection identifiers and color values. Neither setting changes server data.
