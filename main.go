package main

import (
	"log"
	"os"

	"github.com/twharmon/forge/build"
	"github.com/twharmon/forge/serve"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Version: "v0.0.5",
		Name:    "forge",
		Usage:   "Static site generator",
		Commands: []*cli.Command{
			{
				Name:  "build",
				Usage: "Build the site",
				Action: func(c *cli.Context) error {
					return build.All()
				},
			},
			{
				Name:  "serve",
				Usage: "Start the development server",
				Action: func(c *cli.Context) error {
					return serve.Start()
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
