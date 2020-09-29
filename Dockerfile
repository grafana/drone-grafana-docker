FROM golang:1.15.2-alpine3.12 AS build

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 go build -o bin/drone-grafana-docker ./cmd/drone-grafana-docker

FROM docker:19.03.13-dind

ENV DOCKER_HOST=unix:///var/run/docker.sock

COPY --from=build /app/bin/drone-grafana-docker /app/bin/

# Add Gnu tar, since it's used by the build pipeline tool
RUN apk add --no-cache tar

ENTRYPOINT ["/usr/local/bin/dockerd-entrypoint.sh", "/app/bin/drone-grafana-docker"]
