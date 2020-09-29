//+build mage

package main

import (
	"fmt"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const version = "0.3.2"
const imageName = "grafana/drone-grafana-docker"

// Build builds the Docker image.
func Build() error {
	if err := sh.RunV("docker", "build", "-t", imageName, "."); err != nil {
		return err
	}

	return sh.RunV("docker", "tag", imageName, fmt.Sprintf("%s:%s", imageName, version))
}

// Publish publishes the Docker image.
func Publish() error {
	mg.Deps(Build)

	if err := sh.RunV("docker", "push", fmt.Sprintf("%s:%s", imageName, version)); err != nil {
		return err
	}

	return sh.RunV("docker", "push", imageName)
}

func Lint() error {
	if err := sh.RunV("golangci-lint", "run", "./..."); err != nil {
		return err
	}

	return nil
}

var Default = Build
