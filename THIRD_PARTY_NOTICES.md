# Third-party notices

The application as a whole is distributed under GPL-3.0-only. Compatible upstream
copyright and license notices are retained alongside it.

| Component | License | Notice |
|---|---|---|
| Adapted Calendar toolbar code by Nicko Guyer, 2026 | MIT | `LICENSE.upstream.md` |
| github.com/emersion/go-ical | MIT | `licenses/go-ical-MIT.txt` |
| github.com/teambition/rrule-go | MIT | `licenses/rrule-go-MIT.txt` |
| Go standard library/runtime | BSD-3-Clause | `licenses/Go-BSD-3-Clause.txt` |

Dependency versions are pinned in `go.mod` and `go.sum`. The corresponding source
is available at https://github.com/emersion/go-ical and
https://github.com/teambition/rrule-go; Go source is available at https://go.dev/dl/.

XOVI, AppLoad, the tablet's Qt runtime and firmware are prerequisites, not included
in this distribution. Their licenses and installation instructions remain with
their respective upstream projects. The app's icon is a project-created calendar
mark incorporating a recognizable Nextcloud motif; brand names and marks are not
an endorsement by their owners.

The optional `packaging/xovi/nextcloudCalendarSend.qmd` handwriting capture hook
is adapted from [jdkruzr/reTaskable](https://github.com/jdkruzr/reTaskable), commit
`90fa23d2bb3ffad1f89dc045bc2c983174124667`, `app/xovi/retaskableSend.qmd`,
under GPL-3.0-only. It uses the reMarkable cloud handwriting recognition path;
this adaptation queues captures for review in Nextcloud Calendar.

Credit to **jdkruzr**, the author of reTaskable, for the original non-destructive
lasso handwriting capture and recognition dialog. The capture hook retains the
upstream GPL-3.0-only license; its full license is also included as
`licenses/reTaskable-GPL-3.0-only.txt`.
