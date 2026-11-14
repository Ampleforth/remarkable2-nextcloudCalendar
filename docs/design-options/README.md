# Color design options

These are proposals for selection, not screenshots of implemented color support.
The current grayscale app and logo remain unchanged.

| Option | Treatment | Preview |
| --- | --- | --- |
| A | Soft colored blocks, black text, grayscale navigation | [Preview](app-a.png) |
| B | Bold colored blocks, contrasting text, grayscale navigation | [Preview](app-b.png) |
| C | Soft blocks, blue selected controls, tinted task cards | [Preview](app-c.png) |
| D | Soft blocks plus colored navigation and task section accents | [Preview](app-d.png) |

Each proposal includes per-calendar and per-task-list Settings swatches.
Calendar time blocks use the selected calendar color. The task lists illustrate
where their selected colors could appear; existing task grouping can be retained.
All examples use fictional data. Precise e-ink rendering needs hardware validation.

UI previews were generated with the built-in image_gen tool using the current
fictional week screenshot as a reference. Exact prompts: [prompts.json](prompts.json).

The three logo options are native SVG recolorings of `qml/calendar.svg`, rendered
with rsvg-convert so they remain editable and scale cleanly:

1. [Nextcloud blue](logo-1-blue.svg)
2. [Calendar colors](logo-2-multicolor.svg)
3. [Blue tile](logo-3-tile.svg)

[Logo comparison](logo-options.png). Standalone PNG versions are included beside
the SVG files. No generated image has replaced an app runtime asset.

## Selection and revised logos

App style C (Color accents) selected. Logo selection is pending.
The revised logos preserve the exact existing geometry and keep the Nextcloud
connector and all three circles blue (#0082c9). Only calendar colors vary:

1. [Charcoal + multicolor](logo-v2-1.svg)
2. [Teal calendar](logo-v2-2.svg)
3. [Purple calendar](logo-v2-3.svg)

[Revised comparison](logo-options-v2.png). These are SVG recolorings rendered with
rsvg-convert, not AI-generated raster artwork. Earlier logo proposals are superseded.

## Final selections

- App: C, Color accents.
- Logo: option 1, charcoal calendar with blue, green and amber date squares.
  Nextcloud connector and circles remain #0082c9 blue.
- [Selected logo](logo-selected.png), [editable SVG](logo-selected.svg).
- Color-device runtime source: `qml/calendar-color.svg`; grayscale logo remains
  `qml/calendar.svg`.
