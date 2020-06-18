package docker

import "github.com/rs/zerolog/log"

const dockerExe = "/usr/local/bin/docker"
const dockerdExe = "/usr/local/bin/dockerd"

// startDaemon starts the Docker daemon.
func (p Plugin) startDaemon() {
	cmd := commandDaemon(p.Daemon)
	go func() {
		log.Debug().Msg("Starting Docker daemon")
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Error().Err(err).Msgf("Docker daemon failed: %s", output)
		}
	}()
}
