# syntax=docker/dockerfile:1

# Base images are pinned by digest, so that builds are reproducible.

# Build the front end.
FROM node:26.12.0-trixie-slim@sha256:e556a144463cc8dc53afe8de6d2fa0db5c132fd39ffa232c17641bd40df2bf11 AS web
WORKDIR /src/web
COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml web/.npmrc ./
# Install the pnpm version from the packageManager field in package.json.
RUN npm install --global "$(node --print "require('./package.json').packageManager")"
RUN pnpm install --frozen-lockfile
COPY web/ ./
RUN pnpm build

# Build the server, with the front end embedded.
FROM golang:1.27.3-trixie@sha256:c65ecfe613b539ac74c62a4c3b2c27e8f8a6076384c8f27e8a3c0f3be17cc385 AS server
WORKDIR /src
COPY go.* ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
COPY web/web.go web/
COPY --from=web /src/web/build web/build
# A static binary, so that it runs on an image without a C library.
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /diary ./cmd/diary

# The final image contains only the binary: no shell or package manager. The
# nonroot variant runs as an unprivileged user.
FROM gcr.io/distroless/static-debian13:nonroot@sha256:18f4cc26c9ffa7f20a359298be7deadac5c607c8e03a02b8119f1a5fc2f82c99
COPY --from=server /diary /diary
EXPOSE 8080
ENTRYPOINT ["/diary"]
