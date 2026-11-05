# Forma

[![Forma CI](https://github.com/thanoslabs/forma/actions/workflows/forma-ci.yaml/badge.svg?branch=main)](https://github.com/thanoslabs/forma/actions/workflows/forma-ci.yaml)
[![Release](https://img.shields.io/github/v/release/thanoslabs/forma?sort=semver&display_name=tag)](https://github.com/thanoslabs/forma/releases/latest)
[![Forma Pages](https://github.com/thanoslabs/forma/actions/workflows/forma-pages.yaml/badge.svg?branch=main)](https://thanoslabs.github.io/forma/)
[![License: MIT](https://img.shields.io/github/license/thanoslabs/forma)](LICENSE)

Forma is a self-hostable virtual office where people can observe, organize, and interact with software agents. Agent harnesses execute work, project providers remain authoritative for work items, and chat providers extend conversations into tools such as Slack or Discord.

The repository currently includes a Go/PostgreSQL service, provider-neutral adapter contracts, and a React/TypeScript 2.5D office rendered in PixiJS.

## Quick start

Install the versions pinned in `.tool-versions` with `mise install` or `asdf install`, then:

```sh
make dev-up
FORMA_ENABLE_DEMO_API=true make run
```

In another terminal:

```sh
cd web
npm ci
npm run dev
```

Open `http://127.0.0.1:5173`. The Vite development server proxies `/api` requests to the Go service on port 8080.

## Development

- `make test` runs the Go test suite.
- `cd web && npm test` runs the frontend tests.
- `cd web && npm run build` type-checks and builds the web application.
- `make dev-down` stops the development database.

The demo API is intentionally unauthenticated and must never be enabled on a public deployment. See `docs/development.md` for details.

## Architecture

Forma treats the visual office as a projection of observed state rather than the authority for agent execution or work-item ownership. Integrations are capability-negotiated and external provider identifiers are always scoped to a connection.

Start with:

- [Architecture](docs/architecture.md)
- [Integration contracts](docs/integrations.md)
- [Roadmap](docs/roadmap.md)
- [Interactive demo](docs/interactive-demo.md)

Releases use bare SemVer tags such as `1.4.2` with no `v` prefix.
