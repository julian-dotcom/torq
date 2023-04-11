# frontend build stage
FROM node:buster-slim as frontend-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm install --legacy-peer-deps && npm cache clean --force
COPY web/. .
RUN TSX_COMPILE_ON_ERROR=true ESLINT_NO_DEV_ERRORS=true npm run build

# build stage
FROM golang:buster as backend-builder
ENV GO111MODULE=on
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
COPY --from=frontend-builder /app/web/build /app/web/build
RUN CGO_ENABLED=0 GOOS=linux go generate build/version.go
RUN CGO_ENABLED=0 GOOS=linux go build cmd/torq/torq.go

# final stage
FROM debian:buster-slim
COPY --from=backend-builder /app/torq /app/
RUN useradd -ms /bin/bash torq
RUN apt-get -y update && apt-get -y --no-install-recommends install ca-certificates bash && rm -rf /var/lib/apt/lists/*;
RUN update-ca-certificates
ENV GIN_MODE=release
WORKDIR /app
USER torq
ENTRYPOINT ["./torq"]
