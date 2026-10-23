# Icons

The SVG files in this directory come from the
[dashboard-icons](https://github.com/homarr-labs/dashboard-icons) collection by
the Homarr Labs contributors, which is available under the
[Apache License 2.0](https://github.com/homarr-labs/dashboard-icons/blob/main/LICENSE).

All product names, trademarks and registered trademarks are the property of
their respective owners. The icons are used here only to identify the
applications they stand for, and do not imply endorsement by or affiliation with
those projects.

Each file is named after the catalogue entry it belongs to (`jellyfin.svg` for
`jellyfin` in `../catalog.json`). To add an app, add its icon here and an entry
to `catalog.json`. Icons over 50 KB are skipped to keep the binary small, and the whole set is capped at 3 MB. A test checks that the two lists match and that every icon
is a plain, inert SVG: no scripts, event handlers, external links or embedded
images.
