# Working on this repo

## Workflow

- Work incrementally: each PR should be a small, self-contained change that
  is easy to review.
- Keep `docs/roadmap.md` up to date as milestones progress.
- Keep PR titles and descriptions up to date: when a PR changes after it's
  opened (e.g. in response to review feedback), update them to match.
- Don't merge PRs, even once it's agreed that one should be merged: the
  user merges them.

## Code comments

Keep comments within the scope of the code they describe. A comment should
explain what the code does and why, without relying on details of other parts
of the system (such as where it's deployed or who calls it), so that it doesn't
become stale when those other parts change.

## Branch names

Name branches after their purpose, in the form `<type>/<short-description>`,
using kebab-case. For example:

- `feat/healthz-endpoint`
- `ci/github-actions`
- `fix/diary-date-timezone`
- `docs/roadmap`

Use this convention instead of generated branch names (such as
`claude/<random-words>`). Pick the name when starting the work, and use a new
branch for each PR.

## Dependencies

Only use dependency versions that were published at least 14 days ago (Go
modules, GitHub Actions, npm packages, Terraform providers and container
images). This is enforced where the tooling supports it: pnpm's
`minimumReleaseAge` and Dependabot's cooldown.

The one exception is the distroless runtime image in the `Dockerfile`. Its
images have no publication date, so Dependabot's cooldown can't apply, and its
tag is republished often enough that the newest digest would rarely be 14 days
old. Its digest updates can be merged as they come, since the image only
contains CA certificates, time zone data and user account files.
