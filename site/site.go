package site

import (
	"fmt"
	"path"

	"github.com/twharmon/forge/utils"
	"github.com/urfave/cli/v2"
)

func New(c *cli.Context) error {
	name := c.Args().First()
	if err := utils.Mkdir(path.Join(name, "content")); err != nil {
		return fmt.Errorf("newsite.Create: %w", err)
	}
	if err := utils.Mkdir(path.Join(name, "public")); err != nil {
		return fmt.Errorf("newsite.Create: %w", err)
	}
	if err := utils.Mkdir(path.Join(name, "themes")); err != nil {
		return fmt.Errorf("newsite.Create: %w", err)
	}
	if err := utils.WriteFile(path.Join(name, ".gitignore"), []byte(fmt.Sprintf("%s/build\n", name))); err != nil {
		return fmt.Errorf("newsite.Create: %w", err)
	}
	if err := utils.WriteFile(path.Join(name, "config.yml"), []byte("Port: 8080\n")); err != nil {
		return fmt.Errorf("newsite.Create: %w", err)
	}
	if err := utils.WriteFile(path.Join(name, "content", "index.md"), []byte("---\n---\n# Hello, World!\n")); err != nil {
		return fmt.Errorf("newsite.Create: %w", err)
	}
	fmt.Printf("\ncd %s\nforge serve\n\n", name)
	return nil
}
