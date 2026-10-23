# Bifrost

A lightweight, self-hosted application dashboard in the spirit of [Heimdall](https://github.com/linuxserver/Heimdall): one start page that links you to all your apps and services.

## Status

Milestones 0 to 6 are in place: the server runs, accounts work (first-run setup, sign in and out, sessions, adding and removing users, password resets), and you can add, edit, favourite, delete, search and reorder apps on your dashboard. You can also choose a theme and a logo, and add popular apps from a built-in list of 30. The rest of the polish is still to come. See the build order below.

## Running it

Bifrost is distributed as a Docker image. Until images are published, build one yourself:

```sh
docker build -t bifrost .
docker run -d --name bifrost -p 8080:8080 -v bifrost-data:/data bifrost
```

Then open http://localhost:8080. The first visit asks you to create the administrator account. Whoever opens the page first creates it, so do this before exposing Bifrost to anyone else. Everything Bifrost stores lives in the `/data` volume.

### Users

- The first account is the administrator. Administrators add and remove users from **Settings > Users**. Everyone can change their own password from **Settings > Account**.
- Adding a user, or resetting someone's password, makes a random temporary password that is shown once. The person must choose their own the next time they sign in. A reset also signs them out everywhere.
- Removing a user deletes their account and all of their apps. You cannot remove or reset your own account from the Users screen, which also means there is always at least one administrator.
- The list shows when each person was last active.

### Locked out?

There is no password reset on the sign-in page. Reset it from the machine that runs Bifrost:

```sh
docker exec bifrost /bifrost reset-password <username>
```

It prints a temporary password and signs that user out everywhere. They must choose a new password the next time they sign in.

### Your dashboard

- Each user has their own apps. Nothing is shared between users.
- **Add app** is two steps. Step 1 offers 30 popular self-hosted apps, each with its logo, which you can search or filter by category. Pick one and step 2 fills in its name and category, with a hint of the port it usually uses. You only add the address you open it at. Or choose **Custom app** to start from a blank form. Editing an app skips step 1.
- A custom app has no preset to take a name from, so it starts blank. Add an app with a name, an address and a category. Type a new category name to create one, or leave it empty for "Other". A category typed in a different case from an existing one joins it.
- Addresses without a scheme get `http://`. Only `http` and `https` are accepted, and addresses with a username or password in them are refused.
- Starred apps also appear in a Favourites row at the top.
- Tiles show the app's logo if it came from the list, and otherwise its first letter on a coloured background. You can switch a logo off in the app's details.
- Category tabs above the apps show one category at a time. **All** shows everything.
- The search box filters your apps by name, category or address as you type. Press `/` to jump to it and Escape to clear it. Enter opens the best match, or searches the web with DuckDuckGo when nothing matches.
- Drag a tile to change its place within its category. On a touch screen, press and hold first, so the page can still scroll. To move an app without dragging, open it with the pencil and use **Move earlier** or **Move later**. Dragging is off while a search is showing, and your order is saved.

### Appearance

- **Settings > Appearance** sets how Bifrost looks. It is saved per user and follows you to every device you sign in on.
- **Mode** is Light, Dark or System. System follows the device and switches with it, including at sunset.
- Light and dark each remember their own theme. Seven are built in: Bifrost Light and Bifrost Dark, Dracula, Nord, Gruvbox Dark, Solarized Light and Solarized Dark.
- **Logo**: the app icon (default), bare arches, or the five-colour spectrum. It shows in your top bar and as your browser tab icon. App icon and Arches take your theme's accent colour. The tab icon is always the arches on a tile, because the other two do not read well at 16 px.
- The sign-in screen has no user yet, so it always uses the default Bifrost look for the device. Signing out forgets your theme in that browser.

### Sessions and passwords

- Passwords are hashed with argon2id. Browsers get a random session token in an `HttpOnly`, `SameSite=Lax` cookie, and only a hash of the token is stored.
- Without "Keep me signed in", a session ends 12 hours after its last use. With it, 30 days.
- The cookie is marked `Secure` when the request arrives over HTTPS (directly, or through a reverse proxy that sends `X-Forwarded-Proto: https`). Plain HTTP on a home network works too.
- Sign-in attempts are rate limited per address and per username. The limits are kept in memory, so a restart clears them. Behind a reverse proxy every request appears to come from the proxy's address, so the per-address limit is shared.

Settings, all optional:

| Variable | Default | Meaning |
|---|---|---|
| `BIFROST_ADDR` | `:8080` | Address the server listens on |
| `BIFROST_DATA_DIR` | `/data` | Where the database and uploads are kept |

## Development

You need Go 1.26 and Node 22. `make help` lists the targets:

| Command | What it does |
|---|---|
| `make dev` | Go server on :8080 and the Vite dev server on :5173, with hot reload |
| `make build` | Builds the frontend, then `bin/bifrost` with it embedded |
| `make test` | Go tests with the race detector |
| `make lint` | `gofmt`, `go vet` and `svelte-check` |
| `make size` | Checks the built frontend against the 100 KB gzipped budget |
| `make image` | Builds the Docker image |

Project layout:

- `cmd/bifrost`: the binary and its subcommands (`serve`, `reset-password`, `healthcheck`, `version`)
- `internal/config`, `internal/db`, `internal/store`, `internal/auth`, `internal/server`: configuration, SQLite with migrations, data access, passwords and rate limiting, HTTP handlers
- `internal/catalog`: the built-in list of 30 apps (`catalog.json`) and their icons. A test checks that every entry has an icon and every icon is a plain, inert SVG with no scripts, event handlers or external links. The icons come from [dashboard-icons](https://github.com/homarr-labs/dashboard-icons) (Apache-2.0) and belong to their projects. See `internal/catalog/icons/NOTICE.md`
- `internal/themes`: the colour themes. Each is a small JSON file in `internal/themes/bundled`, and a test fails if any theme is hard to read (text and secondary text 4.5:1, accent and status colours 3:1)
- `frontend/src/theme.css` holds the two Bifrost themes for painting the page before a user's theme arrives. A test keeps it identical to the theme files, so change them together
- `internal/webui`: embeds the built frontend
- `frontend`: the Svelte app. `npm run build` writes into `internal/webui/dist`
- Database migrations are numbered SQL files in `internal/db/migrations`, applied in order at start-up

CI (`.github/workflows/ci.yml`) is switched off to save GitHub Actions minutes, and can be started by hand from the Actions tab. When on, it runs the Go checks and tests, and the frontend type check, build and size budget. It never builds the Docker image. So before opening a pull request, run `make lint test size` yourself, and run `make image` when you change the Dockerfile.

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
- A small starter catalogue of about 30 popular apps with icons. The full catalogue comes later.
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
- The full preset catalogue (300+ apps).
- Importing your own theme files, and custom CSS.
- Settings sections for General, Search (including the search provider) and Integrations.
- Installing Bifrost as an app (PWA) with the per-user icon.
- Uploading custom icons and fetching favicons.
- Export and import. Until then, back up by copying the `/data` volume.

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
| 6 | Starter catalogue | Embedded list of about 30 apps with icons, and the two-step add flow. |
| 7 | Responsive and touch polish | Review at every breakpoint and orientation, bottom sheets on phones, touch behaviour for edit and reorder. |
| 8 | Release | Multi-architecture image, accessibility pass, end-to-end tests of login, add app and reorder, and user docs. |

### Risks to watch

- Drag-and-drop on touch screens. Prototype it on a real phone in step 4.
- SQLite allows one writer at a time. Use WAL mode and a busy timeout from the start.
- The first-run endpoint must stop working as soon as the first admin exists.
- Session cookies need the secure flags, requests that change data need protection against cross-site requests, and sign-in attempts should be rate limited.

## Open decisions

- How status checks and widgets should run once they are added (proposal: a plain HTTP request every 60 seconds, configurable per app).
- Exact flag names for the operations subcommands (`reset-password` takes only a username for now).
