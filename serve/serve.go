package serve

import (
	"fmt"
	"os"

	"github.com/twharmon/forge/config"
	"github.com/twharmon/forge/devserver"
)

func Start() error {
	os.Setenv("DEBUG", "true")
	cfg, err := config.Get()
	if err != nil {
		return fmt.Errorf("serve.Start: %w", err)
	}
	server, err := devserver.New(cfg)
	if err != nil {
		return fmt.Errorf("serve.Start: %w", err)
	}
	defer server.Shutdown()
	return server.Run()
}
