# drone-grafana-docker

Drone plugin that uses Docker-in-Docker to build and publish Grafana Docker images.

## Build

You will need to install [Mage](https://magefile.org) in order to build this project.

Build the Docker image with the following command:

```console
mage
```

## Usage

> Notice: Be aware that this Docker plugin requires privileged capabilities, otherwise the integrated Docker daemon is 
not able to start.
