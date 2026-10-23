<p align="center">
  <img src="docs/logo.svg" alt="Bifrost" width="96" height="96">
</p>

<h1 align="center">Bifrost</h1>

<p align="center">
  A small, fast, self-hosted start page for your apps and services.<br>
  One page that links to everything you run. A lightweight take on <a href="https://github.com/linuxserver/Heimdall">Heimdall</a>.
</p>

<p align="center">
  <a href="https://github.com/dbohry/bifrost/actions/workflows/ci.yml"><img alt="CI" src="https://github.com/dbohry/bifrost/actions/workflows/ci.yml/badge.svg"></a>
  <a href="https://github.com/dbohry/bifrost/actions/workflows/release.yml"><img alt="Release" src="https://github.com/dbohry/bifrost/actions/workflows/release.yml/badge.svg"></a>
  <a href="https://hub.docker.com/r/dbohry/bifrost"><img alt="Docker pulls" src="https://img.shields.io/docker/pulls/dbohry/bifrost"></a>
  <a href="https://hub.docker.com/r/dbohry/bifrost"><img alt="Image size" src="https://img.shields.io/docker/image-size/dbohry/bifrost/latest"></a>
  <img alt="Go version" src="https://img.shields.io/github/go-mod/go-version/dbohry/bifrost">
  <a href="https://github.com/dbohry/bifrost/commits/main"><img alt="Last commit" src="https://img.shields.io/github/last-commit/dbohry/bifrost"></a>
</p>

<p align="center">
  <img src="docs/screenshots/dashboard-light.png" alt="Bifrost dashboard, light theme" width="49%">
  <img src="docs/screenshots/dashboard-dark.png" alt="Bifrost dashboard, Dracula theme" width="49%">
</p>

## Features

- **About 390 apps ready to add**, with their logos: Jellyfin, Sonarr, Home Assistant, Immich, Nextcloud and many more. Or add any address yourself.
- **Private dashboards.** Every user has their own apps, favourites and look.
- **Themes.** Light, dark or follow the device. Bifrost, Dracula, Nord, Gruvbox and Solarized are built in.
- **Works on phones, tablets and desktops**, in either orientation.
- **Search** your apps as you type, or fall back to a web search.
- **Backup.** Export your apps and look to a file and import them anywhere.
- **Tiny.** One Go binary and one SQLite file. No other services.

## Run it

You need [Docker](https://docs.docker.com/get-docker/). The image runs on PCs, servers and Raspberry Pi (amd64 and arm64).

```sh
docker run -d --name bifrost --restart unless-stopped \
  -p 8080:8080 -v bifrost-data:/data \
  dbohry/bifrost:latest
```

Or with Docker Compose, using the [`docker-compose.yml`](docker-compose.yml) in this repository:

```sh
docker compose up -d
```

Then open <http://localhost:8080>. The first visit asks you to create the administrator account. **Whoever opens the page first creates it**, so do this before you share the address.

Everything Bifrost stores lives in the `/data` volume.

> `latest` is rebuilt from `main` on every merge. To build the image yourself, run `docker build -t bifrost .` and use `bifrost` instead of `dbohry/bifrost:latest` above.

### Settings

Both are optional.

| Variable | Default | Meaning |
|---|---|---|
| `BIFROST_ADDR` | `:8080` | Address the server listens on |
| `BIFROST_DATA_DIR` | `/data` | Where the database is kept |

### Update

```sh
docker pull dbohry/bifrost:latest
docker stop bifrost && docker rm bifrost
# then run it again with the same command as above
```

Your data stays in the volume, and the database upgrades itself when Bifrost starts. With Compose: `docker compose pull && docker compose up -d`.

### Back up

- **Your own dashboard:** Settings → Backup → Download export.
- **The whole server**, with every user: stop the container and copy the `/data` volume.

### Locked out?

There is no password reset on the sign-in page. Reset it from the machine that runs Bifrost:

```sh
docker exec bifrost /bifrost reset-password <username>
```

It prints a temporary password and signs that user out everywhere. They choose a new one at their next sign-in.

### HTTPS and reverse proxies

Plain HTTP is fine on a home network. Behind a reverse proxy that serves HTTPS, have it send `X-Forwarded-Proto: https` and Bifrost marks its cookie as secure. Every request then appears to come from the proxy, so the sign-in rate limit is shared between all users.

## Using it

- **Add an app.** Press **Add app**. Pick one from the list, enter the address you open it at, and save. Or choose **Custom app**.
- **Reorder.** Drag a tile. On a touch screen, press and hold first. Or open the pencil and use **Move earlier** and **Move later**.
- **Favourites.** Star an app to also show it in a row at the top.
- **Search.** Press `/` to jump to the search box. Enter opens the best match, or searches DuckDuckGo if nothing matches.
- **Touch screens.** The edit buttons are hidden until you press **Edit**. On a phone held upright, the main buttons are in a bar at the bottom.
- **Users.** The first account is the administrator and can add and remove users in Settings → Users. A new user gets a temporary password, shown once, and must choose their own.
- **Look.** Settings → Appearance sets the mode, a theme for light and one for dark, and the logo. It follows you to every device.
- **Esc** leaves Settings.

## More

- [Development](docs/DEVELOPMENT.md): build, test and project layout.
- [Design notes](docs/DESIGN.md): goals, architecture, scope and what comes next.
- Icons come from [dashboard-icons](https://github.com/homarr-labs/dashboard-icons) (Apache-2.0) and belong to their projects. See [`NOTICE.md`](internal/catalog/icons/NOTICE.md).
