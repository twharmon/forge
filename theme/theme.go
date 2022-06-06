package theme

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/urfave/cli/v2"
)

func Add(c *cli.Context) error {
	themeUrl := c.Args().First()
	themeName := getThemeName(themeUrl)
	themePath := path.Join("themes", themeName)
	if _, err := os.Stat(themePath); err == nil {
		return fmt.Errorf("theme %s already added", themeName)
	}
	if err := exec.Command("git", "clone", themeUrl, themePath, "--depth=1").Run(); err != nil {
		return fmt.Errorf("theme.Add: %w", err)
	}
	if err := os.RemoveAll(path.Join(themePath, ".git")); err != nil {
		return fmt.Errorf("theme.Add: %w", err)
	}
	return nil
}

func Remove(c *cli.Context) error {
	themeName := getThemeName(c.Args().First())
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

func getThemeName(url string) string {
	base := path.Base(url)
	return strings.TrimSuffix(base, ".git")
}
