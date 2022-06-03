package main

import (
	"log"
	"os"

	"github.com/twharmon/forge/build"
	"github.com/twharmon/forge/serve"
	"github.com/twharmon/forge/site"
	"github.com/twharmon/forge/theme"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Version: "v0.0.7",
		Name:    "forge",
		Usage:   "Static site generator",
		Commands: []*cli.Command{
			{
				Name:  "new",
				Usage: "Create a new site",
				Action: func(c *cli.Context) error {
					return site.New(c)
				},
			},
			{
				Name:  "add-theme",
				Usage: "Add a theme",
				Action: func(c *cli.Context) error {
					return theme.Add(c)
				},
			},
			{
				Name:  "remove-theme",
				Usage: "Remove a theme",
				Action: func(c *cli.Context) error {
					return theme.Remove(c)
				},
			},
			{
				Name:  "serve",
				Usage: "Start the development server",
				Action: func(c *cli.Context) error {
					return serve.Start()
				},
			},
			{
				Name:  "build",
				Usage: "Build the site",
				Action: func(c *cli.Context) error {
					return build.All()
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
