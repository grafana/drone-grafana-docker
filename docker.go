package docker

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type (
	// Daemon defines Docker daemon parameters.
	Daemon struct {
	}

	// Login defines Docker login parameters.
	Login struct {
		Username string // Docker registry username
		Password string // Docker registry password
	}

	// Build defines Docker build parameters.
	Build struct {
		Remote string // Git remote URL
		Name   string // Docker build using default named tag
		Squash bool   // Docker build squash
	}

	// Plugin defines the Docker plugin parameters.
	Plugin struct {
		Login   Login  // Docker login configuration
		Build   Build  // Docker build configuration
		Daemon  Daemon // Docker daemon configuration
		Dryrun  bool   // Docker push is skipped
		Cleanup bool   // Docker purge is enabled
	}
)

// Exec executes the plugin step.
func (p Plugin) Exec() error {
	log.Debug().Msgf("Starting Docker daemon")
	p.startDaemon()

	const maxAttempts = 15

	// poll the docker daemon until it is started. This ensures the daemon is
	// ready to accept connections before we proceed.
	i := 0
	for i = 0; i < maxAttempts; i++ {
		log.Debug().Msgf("Polling Docker daemon to see if it's ready")
		cmd := commandInfo()
		if err := cmd.Run(); err == nil {
			break
		}
		time.Sleep(time.Second * 1)
	}
	if i >= maxAttempts {
		return fmt.Errorf("docker daemon didn't come up on time")
	}

	if !p.Dryrun {
		if p.Login.Username == "" || p.Login.Password == "" {
			env := os.Environ()
			log.Error().Str("environment", strings.Join(env, ", ")).Msg("Username or password not in environment")
			return fmt.Errorf("registry credentials must be provided")
		}

		// Log into the Docker registry
		cmd := commandLogin(p.Login)
		if output, err := cmd.CombinedOutput(); err != nil {
			log.Error().Err(err).Msg("Failed logging into Docker registry")
			return fmt.Errorf("error authenticating: %w\n%s", err, output)
		}
		log.Info().Msg("Successfully logged into Docker registry")
	}

	var cmds []*exec.Cmd
	cmds = append(cmds, commandVersion())
	cmds = append(cmds, commandInfo())
	// Enable execution of Docker images for other architectures
	cmds = append(cmds, exec.Command("docker", "run", "--privileged", "--rm",
		"docker/binfmt:a7996909642ee92942dcd6cff44b9b95f08dad64"))
	// TODO: Take --ubuntu into account
	cmds = append(cmds, exec.Command("./bin/grabpl", "build-docker"))

	if p.Cleanup {
		cmds = append(cmds, commandRmi(p.Build.Name))
		cmds = append(cmds, commandPrune())
	}

	// Execute all commands in batch mode
	for _, cmd := range cmds {
		log.Debug().Msgf("Executing %q: %s", cmd.Path, strings.Join(cmd.Args, ", "))

		output, err := cmd.CombinedOutput()
		if err != nil {
			switch {
			case isCommandPrune(cmd.Args):
				log.Warn().Str("output", string(output)).Msg("Could not prune system containers. Ignoring...")
			case isCommandRmi(cmd.Args):
				log.Warn().Str("output", string(output)).Msgf("Could not remove image %q. Ignoring...", cmd.Args[2])
			default:
				return fmt.Errorf("Command failed: %w\n%s", err, output)
			}
		}
	}

	return nil
}

// commandLogin creates the docker login command.
func commandLogin(login Login) *exec.Cmd {
	log.Info().Str("username", login.Username).Msgf("Logging into Docker registry")
	return exec.Command(
		dockerExe, "login",
		"-u", login.Username,
		"-p", login.Password,
	)
}

func commandVersion() *exec.Cmd {
	return exec.Command(dockerExe, "version")
}

func commandInfo() *exec.Cmd {
	return exec.Command(dockerExe, "info")
}

// commandDaemon is a helper function to create the docker daemon command.
func commandDaemon(daemon Daemon) *exec.Cmd {
	args := []string{
		"--host=unix:///var/run/docker.sock",
		// Required for making manifests
		"--experimental",
	}

	return exec.Command(dockerdExe, args...)
}

func isCommandPrune(args []string) bool {
	return len(args) > 3 && args[2] == "prune"
}

func commandPrune() *exec.Cmd {
	return exec.Command(dockerExe, "system", "prune", "-f")
}

func isCommandRmi(args []string) bool {
	return len(args) > 2 && args[1] == "rmi"
}

func commandRmi(tag string) *exec.Cmd {
	return exec.Command(dockerExe, "rmi", tag)
}
