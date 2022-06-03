package theme

import (
	"fmt"
	"os"
	"path"

	"github.com/go-git/go-git/v5"
	"github.com/urfave/cli/v2"
)

func Add(c *cli.Context) error {
	themeName := path.Base(c.Args().First())
	themePath := path.Join("themes", themeName)
	if _, err := os.Stat(themePath); err == nil {
		return fmt.Errorf("theme %s already added", themeName)
	}
	_, err := git.PlainClone(themePath, false, &git.CloneOptions{
		URL:   c.Args().First(),
		Depth: 1,
	})
	if err != nil {
		return fmt.Errorf("theme.Add: %w", err)
	}
	if err := os.RemoveAll(path.Join(themePath, ".git")); err != nil {
		return fmt.Errorf("theme.Add: %w", err)
	}
	return nil
}

func Remove(c *cli.Context) error {
	themeName := path.Base(c.Args().First())
	themePath := path.Join("themes", themeName)
	if _, err := os.Stat(themePath); err != nil {
		return fmt.Errorf("theme %s not found", themeName)
	}
	if err := os.RemoveAll(themePath); err != nil {
		return fmt.Errorf("theme.Remove: %w", err)
	}
	return nil
}

func Update(c *cli.Context) error {
	if err := Remove(c); err != nil {
		return fmt.Errorf("theme.Update: %w", err)
	}
	if err := Add(c); err != nil {
		return fmt.Errorf("theme.Update: %w", err)
	}
	return nil
}
