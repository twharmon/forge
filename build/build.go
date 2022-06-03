package build

import (
	"bytes"
	"fmt"
	"forge/config"
	"forge/utils"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/yuin/goldmark"
	"gopkg.in/yaml.v3"
)

type statistics struct {
	start   time.Time
	pageCnt int
}

func (s *statistics) Print() {
	fmt.Printf("%d pages built in %dms\n\n", s.pageCnt, time.Since(s.start).Milliseconds())
}

func All() error {
	stats := &statistics{
		start: time.Now(),
	}
	cfg, err := config.Get()
	if err != nil {
		return fmt.Errorf("build.All: %w", err)
	}
	if err := os.RemoveAll("build"); err != nil {
		return fmt.Errorf("build.All: %w", err)
	}
	if err := os.Mkdir("build", 0777); err != nil {
		return fmt.Errorf("build.All: %w", err)
	}
	if err := utils.CopyDirectory("public", "build"); err != nil {
		return err
	}
	if err := utils.CopyDirectory(path.Join("themes", cfg.Theme, "public"), "build"); err != nil {
		return err
	}
	t, err := template.ParseFiles(path.Join("themes", cfg.Theme, "layouts", "base.html"))
	if err != nil {
		return fmt.Errorf("build.All: %w", err)
	}
	t, err = t.ParseGlob(path.Join("themes", cfg.Theme, "layouts/*"))
	if err != nil {
		return fmt.Errorf("build.All: %w", err)
	}
	if err := dir(t, cfg, stats, "content"); err != nil {
		return err
	}
	stats.Print()
	return nil
}

func dir(t *template.Template, cfg *config.Config, stats *statistics, dirName string) error {
	fis, err := ioutil.ReadDir(dirName)
	if err != nil {
		return fmt.Errorf("build.dir: %w", err)
	}
	for _, fi := range fis {
		if fi.IsDir() {
			if err := dir(t, cfg, stats, path.Join(dirName, fi.Name())); err != nil {
				return fmt.Errorf("build.dir: %w", err)
			}
			continue
		}
		stats.pageCnt++
		if err := page(t, cfg, path.Join(dirName, fi.Name())); err != nil {
			return fmt.Errorf("build.dir: %w", err)
		}
	}
	return nil
}

func page(t *template.Template, cfg *config.Config, page string) error {
	pagePath := "build/" + strings.TrimPrefix(page, "content/")
	pageDir, pageName := path.Split(pagePath)
	if !strings.HasPrefix(pageName, "index.") {
		nameParts := strings.Split(pageName, ".")
		pageDir = path.Join(pageDir, nameParts[0])
		pageName = "index.html"
	}
	if err := os.MkdirAll(pageDir, 0777); err != nil {
		return fmt.Errorf("build.page: %w", err)
	}
	f, err := os.Create(path.Join(pageDir, strings.ReplaceAll(pageName, ".md", ".html")))
	if err != nil {
		return fmt.Errorf("build.page: %w", err)
	}
	cfg.Forge.Path = strings.TrimPrefix(pageDir, "build")
	t, err = t.Clone()
	if err != nil {
		return fmt.Errorf("build.page: %w", err)
	}
	pageParams := make(map[string]interface{})
	if strings.HasSuffix(page, ".md") {
		b, err := ioutil.ReadFile(page)
		if err != nil {
			return fmt.Errorf("build.page: %w", err)
		}
		parts := bytes.SplitN(b, []byte("---"), 3)
		if len(parts) != 3 {
			return fmt.Errorf(`build.page: maleformed content; must have front matter surrounded by "---"`)
		}
		if err := yaml.Unmarshal(parts[1], &pageParams); err != nil {
			return fmt.Errorf("build.page: %w", err)
		}
		var buf bytes.Buffer
		if err := goldmark.Convert(parts[2], &buf); err != nil {
			return fmt.Errorf("build.page: %w", err)
		}
		t, err = t.Parse(fmt.Sprintf(`{{ define "body" }}%s{{ end }}`, buf.String()))
		if err != nil {
			return fmt.Errorf("build.page: %w", err)
		}
	} else if strings.HasSuffix(page, ".html") {
		t, err = t.ParseFiles(page)
		if err != nil {
			return fmt.Errorf("build.page: %w", err)
		}
	} else {
		return fmt.Errorf("build.page: invalid content: %s", page)
	}
	if err := t.Execute(f, map[string]interface{}{
		"Theme": cfg.ThemeParams,
		"Page":  pageParams,
		"Forge": cfg.Forge,
	}); err != nil {
		return fmt.Errorf("build.page: %w", err)
	}
	return nil
}
