# build stage
FROM golang:buster as backend-builder
ARG BUILD_VER=v0.1.1-dev
ENV GO111MODULE=on
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-X github.com/lncapital/torq/build.overrideBuildVer=$BUILD_VER" cmd/torq/torq.go

# frontend build stage
FROM node:buster-slim as frontend-builder
WORKDIR /app
COPY web/package*.json ./
RUN npm install --legacy-peer-deps && npm cache clean --force;
COPY web/. .
RUN TSX_COMPILE_ON_ERROR=true ESLINT_NO_DEV_ERRORS=true npm run build

# final stage
FROM debian:buster-slim
COPY --from=backend-builder /app/torq /app/
COPY --from=frontend-builder /app/build /app/web/build
RUN useradd -ms /bin/bash torq
RUN apt-get -y update && apt-get -y --no-install-recommends install ca-certificates bash && rm -rf /var/lib/apt/lists/*;
RUN update-ca-certificates
ENV GIN_MODE=release
WORKDIR /app
USER torq
ENTRYPOINT ["./torq"]
