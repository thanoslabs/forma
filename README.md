# template

[![CI](https://github.com/thanoslabs/template-repo/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/thanoslabs/template-repo/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/thanoslabs/template-repo?sort=semver&display_name=tag)](https://github.com/thanoslabs/template-repo/releases/latest)
[![Docs](https://github.com/thanoslabs/template-repo/actions/workflows/docs.yaml/badge.svg?branch=main)](https://thanoslabs.github.io/template-repo/)
[![Docs site](https://img.shields.io/badge/docs-GitHub%20Pages-blue?logo=github)](https://thanoslabs.github.io/template-repo/)
[![Renovate](https://img.shields.io/badge/renovate-enabled-brightgreen?logo=renovatebot)](https://github.com/thanoslabs/template-repo/pulls?q=is%3Apr+author%3Aapp%2Frenovate)
[![SemVer](https://img.shields.io/badge/semver-2.0.0-blue)](https://semver.org/)
[![License: MIT](https://img.shields.io/github/license/thanoslabs/template-repo)](LICENSE)

Template repository for `thanoslabs` projects.

## What's Included

- **Renovate** — automated dependency updates for Docker, Go modules, and GitHub Actions
- **Releases** — automatic SemVer bumps on merge to `main` via PR labels (`major`/`minor`/`patch`, default `patch`), plus manual dispatch for explicit versions
- **Branch protection** — `main` requires PRs and review
- **CODEOWNERS** — `@thanoslabs/reviewers` auto-requested for review
- **MIT License**
- **.gitignore** — common exclusions for OS, IDE, build outputs, and secrets
- **.tool-versions** — single source of truth for language/tool versions (asdf/mise compatible)
- **Docs** — markdown in `docs/` built into a static site with [nebula-md](https://github.com/thanoslabs/nebula-md) and deployed to GitHub Pages (public repos only)

## Using This Template

1. Click **Use this template** to create a new repository
2. Pin your project's language and tool versions in `.tool-versions`
3. Update `renovate.json` to configure managers and schedules for your project
4. Enable the new repo in the Renovate GitHub App if using hosted Renovate
5. Fill in the artifact build/publish TODO stub in `.github/workflows/release.yaml`
6. For public repos, set **Settings → Pages → Source** to **GitHub Actions** so the docs site can deploy
7. Replace `thanoslabs/template-repo` in the README badges with your repo's `owner/name` (and drop the docs site badge for private repos)

## Renovate

Dependency updates are managed via Renovate. Configuration is in `renovate.json` and `.github/renovate-config.js`.

Enabled managers:
- `asdf` (keeps `.tool-versions` up to date)
- `docker-compose`
- `github-actions`
- `gomod`

Add or remove managers as needed for your project.

## Releases

Releases follow [SemVer](https://semver.org/) with bare `X.Y.Z` tags (no `v` prefix).

- **Automatic**: merging a PR to `main` cuts a release based on its `major`, `minor`, or `patch` label; no label means a `patch` bump.
- **Manual**: run the **Release** workflow with an explicit `X.Y.Z` version to cut a specific release.

Automation lives in `.github/workflows/release.yaml` (versioning, tag, GitHub release). Building and publishing artifacts (images, binaries) is project-specific — fill in the TODO stub in the workflow.

## Docs

Project documentation lives in `docs/` as Obsidian-flavoured markdown (frontmatter, `[[wikilinks]]`, callouts). On every push to `main` that touches `docs/`, `.github/workflows/docs.yaml` calls the reusable [nebula-md](https://github.com/thanoslabs/nebula-md) workflow to build a static site and deploy it to GitHub Pages.

- **Public repos**: the site is built and deployed to GitHub Pages.
- **Private repos**: the workflow is skipped — nothing is published.

You can also trigger it manually via **Actions → Docs → Run workflow**.

## Files

```
.
├── .github/
│   ├── CODEOWNERS           # Auto-request review from @thanoslabs/reviewers
│   ├── renovate-config.js   # Renovate platform config
│   └── workflows/
│       ├── docs.yaml        # Build docs/ with nebula-md, deploy to GitHub Pages (public repos)
│       ├── release.yaml     # Label-driven SemVer release workflow
│       └── renovate.yaml    # Renovate GitHub Action workflow
├── docs/                   # Project documentation (markdown → nebula-md site)
├── .gitignore
├── .tool-versions          # Pinned language/tool versions (asdf/mise)
├── CONTRIBUTING.md
├── LICENSE
├── README.md
├── renovate.json           # Renovate settings
└── SECURITY.md
```
