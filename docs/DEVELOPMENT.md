# Development

You need Go 1.26 and Node 22. `make help` lists the targets:

| Command | What it does |
|---|---|
| `make dev` | Go server on :8080 and the Vite dev server on :5173, with hot reload |
| `make build` | Builds the frontend, then `bin/bifrost` with it embedded |
| `make test` | Go tests with the race detector |
| `make lint` | `gofmt`, `go vet` and `svelte-check` |
| `make e2e` | Browser checks of the main journeys and an accessibility scan of every screen and theme, using `bin/bifrost`, so run `make build` first |
| `make size` | Checks the built frontend against the 100 KB gzipped budget |
| `make image` | Builds the Docker image |

Browser checks live in `e2e/` and need Chromium: run `npm install` in that directory once, then `make e2e`. Each check starts its own server on an empty data directory; to run one by hand, start Bifrost on an empty directory and use `BASE=http://127.0.0.1:8080 npm run flows`, `npm run backup` or `npm run a11y`. The accessibility scan (axe-core) runs on a phone and a desktop in all seven themes and must report nothing.

Project layout:

- `cmd/bifrost`: the binary and its subcommands (`serve`, `reset-password`, `healthcheck`, `version`)
- `internal/config`, `internal/db`, `internal/store`, `internal/auth`, `internal/server`: configuration, SQLite with migrations, data access, passwords and rate limiting, HTTP handlers
- `internal/catalog`: the built-in list of about 580 apps (`catalog.json`) and their icons. A test checks that every entry has an icon and every icon is a plain, inert SVG with no scripts, event handlers or external links. The icons come from [dashboard-icons](https://github.com/homarr-labs/dashboard-icons) (Apache-2.0) and belong to their projects. See `internal/catalog/icons/NOTICE.md`
- `internal/themes`: the colour themes. Each is a small JSON file in `internal/themes/bundled`, and a test fails if any theme is hard to read (text and secondary text 4.5:1, accent and status colours 3:1)
- `frontend/src/theme.css` holds the two Bifrost themes for painting the page before a user's theme arrives. A test keeps it identical to the theme files, so change them together
- `internal/webui`: embeds the built frontend
- `frontend`: the Svelte app. `npm run build` writes into `internal/webui/dist`
- Database migrations are numbered SQL files in `internal/db/migrations`, applied in order at start-up

CI (`.github/workflows/ci.yml`) runs the Go checks and tests and the frontend type check, build and size budget on every pull request (not again when it is merged to main; the pull request already ran them, and it can be started by hand from the Actions tab). The Release workflow (`.github/workflows/release.yml`) builds the image for amd64 and arm64 and publishes `dbohry/bifrost:latest` (and `dbohry/bifrost:main-<commit>`, to pin an exact build) to Docker Hub on every merge to main, except changes to docs and browser checks only. It can also be started by hand from the Actions tab. When the image is pushed, a final Deploy job calls a Portainer webhook so the service pulls it. It needs three repository secrets: `DOCKERHUB_USERNAME`, `DOCKERHUB_TOKEN` and `DEPLOY_WEBHOOK` (the Portainer webhook URL). There are no release version tags for now. The browser checks (`make e2e`) are not in CI yet, so run `make e2e` before opening a pull request that changes the interface.
