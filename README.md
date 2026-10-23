# Bifrost

A lightweight, self-hosted application dashboard in the spirit of [Heimdall](https://github.com/linuxserver/Heimdall): one start page that links you to all your apps and services.

## Status

Design phase. No application code yet.

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

## Out of scope for the MVP

- Status checks (the online/offline dot and its text on tiles). The wireframes show them, but they are deferred.
- Widgets (live per-app stats such as Jellyfin streams or Sonarr queue size). They depend on the same background worker, so they wait with status checks.

## Open decisions

- How status checks and widgets should run once they are added (proposal: a plain HTTP request every 60 seconds, configurable per app).
- Exact command and flag names for the operations subcommands.
