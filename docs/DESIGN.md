# Design notes

Why Bifrost is built the way it is, what is in it, and what comes later.

## Goals

- Use as few resources as possible, both at idle and under load.
- Local first: minimal login, no password-reset email, no single sign-on.
- Works on phones, tablets and desktops, in portrait and landscape.
- Themable, with Dracula, Nord, Gruvbox, Solarized and others, and room for user-made themes.

## Architecture

| Layer | Choice |
|---|---|
| Backend | Go, built as one static binary |
| Database | SQLite (WAL mode, pure-Go driver), stored in a `/data` volume |
| Frontend | Svelte 5 + Vite + TypeScript, as a single-page app (no SvelteKit) |
| Styling | Plain CSS with theme variables, no UI component library |
| Distribution | Docker only, multi-architecture (`amd64` and `arm64`) |

### Backend

- One Go binary serves the JSON API and the built frontend, which is embedded in the binary.
- SQLite holds users, sessions, apps and preferences. Nothing else runs next to it.
- Every user has a private dashboard. Apps belong to one user and are never shared between users.
- Passwords are hashed with argon2id. Sessions are stored in SQLite.
- The same binary has subcommands for operations: `healthcheck` (no `curl` in the image) and `reset-password <username>`. Because the sign-in page has no reset, a locked-out user is helped from the host with `docker exec`.
- The preset app catalogue is embedded in the binary. Custom icons and user-added themes live under `/data`.

### Frontend

- Built with Vite and embedded into the Go binary at image build time. Node is needed only to build, never to run.
- The API is JSON from the start, so screens can be added or replaced one at a time.
- Drag-to-reorder tiles, bottom sheets, dialogs, and live theme and logo preview are client-side.
- A bundle-size budget (target under 100 KB gzipped) is checked in CI to protect the lightweight goal.

### Theming

- Components never use raw colours. They read ten tokens: `bg`, `surface`, `border`, `text`, `muted`, `accent`, `onAccent`, `ok`, `bad` and `iconBg`.
- A theme is a small file of those ten values, plus a `mode` of light or dark.
- Light and dark each remember their own theme per user, and System follows the operating system.
- Themes are checked on import: `text`, `muted` and `onAccent` must reach 4.5:1 contrast. A theme that fails shows a warning instead of breaking the page. A missing token falls back to the default theme of the same mode.
- Logo (app icon, arches or spectrum) is also a per-user choice. The installed app icon is fixed when the app is installed.

### Docker

- Minimal base image (distroless or `scratch`), running as a non-root user.
- One exposed port and one volume at `/data`.
- Configuration through environment variables.

## MVP scope

**Accounts**
- First-run setup: create the first admin account.
- Sign in and sign out, including "Keep me signed in".
- Users screen: add a user, change a password, remove a user.
- `reset-password` command on the host.

**Dashboard**
- Tiles grouped by category, a Favourites row and category tabs.
- Add, edit and delete an app: name, URL, category and icon.
- A catalogue of about 580 popular apps with icons. It will keep growing.
- Categories are created on the fly when adding an app. There is no screen to manage them.
- Apps without a preset or uploaded icon show their first letter on a coloured tile.
- Drag-to-reorder tiles.
- Search over your apps. When nothing matches, Enter searches the web with DuckDuckGo. The provider is fixed for now.

**Look and feel**
- Responsive layouts for phone, tablet and desktop, in portrait and landscape.
- Light and dark mode, with six bundled themes (Bifrost, Dracula, Nord, Gruvbox, Solarized Dark and Solarized Light), chosen per user.
- Per-user logo choice: app icon (default), arches or spectrum.

**Packaging**
- Multi-architecture Docker image with a `/data` volume and a health check.

## Later

- Status checks (the online/offline dot and its text on tiles). The wireframes show them, but they are deferred.
- Widgets (live per-app stats such as Jellyfin streams or Sonarr queue size). They depend on the same background worker, so they wait with status checks.
- Importing your own theme files, and custom CSS.
- Settings sections for General, Search (including the search provider) and Integrations.
- Installing Bifrost as an app (PWA) with the per-user icon.
- Uploading custom icons and fetching favicons.

## Build order

Each step ends with something that can be run and tried. Layouts are built mobile-first from step 2, so step 7 is a review, not a rewrite.

| # | Milestone | Result |
|---|---|---|
| 0 | Skeleton and pipeline | Go server with an embedded Svelte page, SQLite with migrations, multi-stage Dockerfile, `/data` volume, `healthcheck` subcommand, and CI that lints, tests, builds the image and checks the bundle size. |
| 1 | Accounts | First-run setup, sign in and sign out, sessions and the `reset-password` command. Also the Svelte shell, router and API client. |
| 2 | Dashboard core | Add, edit and delete custom apps, grouped by category, with a Favourites row and first-letter icons. Each app belongs to one user. |
| 3 | Users screen | Add a user, change a password, remove a user. Needed to test that dashboards are private. |
| 4 | Organise and search | Drag-to-reorder with the order saved, category tabs, search over apps and the DuckDuckGo fallback. |
| 5 | Themes and logo | Mode switch, six bundled themes, logo choice and per-user preferences. A test checks the contrast rule on every bundled theme. |
| 6 | Starter catalogue | Embedded list of apps with icons (30 at first, about 580 now), and the two-step add flow. |
| 7 | Responsive and touch polish | Review at every breakpoint and orientation, a bottom bar and sheets on phones, touch behaviour for edit and reorder. Done. |
| 8 | Release | Multi-architecture image, accessibility pass, end-to-end tests of login, add app and reorder, and user docs. Done. |

### Risks to watch

- Drag-and-drop on touch screens. Prototype it on a real phone in step 4.
- SQLite allows one writer at a time. Use WAL mode and a busy timeout from the start.
- The first-run endpoint must stop working as soon as the first admin exists.
- Session cookies need the secure flags, requests that change data need protection against cross-site requests, and sign-in attempts should be rate limited.

## Open decisions

- How status checks and widgets should run once they are added (proposal: a plain HTTP request every 60 seconds, configurable per app).
- Exact flag names for the operations subcommands (`reset-password` takes only a username for now).

## Sessions and passwords

- Passwords are hashed with argon2id. Browsers get a random session token in an `HttpOnly`, `SameSite=Lax` cookie, and only a hash of the token is stored.
- Without "Keep me signed in", a session ends 12 hours after its last use. With it, 30 days.
- The cookie is marked `Secure` when the request arrives over HTTPS (directly, or through a reverse proxy that sends `X-Forwarded-Proto: https`).
- Sign-in attempts are rate limited per address and per username. The limits are kept in memory, so a restart clears them.
- Removing a user deletes their account and all of their apps. You cannot remove or reset your own account from the Users screen, so there is always at least one administrator.
- Importing a file runs every app through the same checks as adding one by hand. Only `http` and `https` addresses are accepted.
