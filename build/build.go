package build

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"text/template"
	"time"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/twharmon/forge/config"
	"github.com/twharmon/forge/utils"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"

	"gopkg.in/yaml.v3"
)

type Build struct {
	start    time.Time
	cfg      *config.Config
	markdown goldmark.Markdown
	minify   *minify.M
	template *template.Template
}

func New() (*Build, error) {
	start := time.Now()
	fmt.Println("Loading config...")
	cfg, err := config.Get()
	if err != nil {
		return nil, fmt.Errorf("build.New: %w", err)
	}
	var exts []goldmark.Extender
	for _, ext := range cfg.Markdown.Extensions {
		switch ext {
		case "footnote":
			exts = append(exts, extension.Footnote)
		case "linkify":
			exts = append(exts, extension.NewLinkify(extension.WithLinkifyAllowedProtocols([][]byte{[]byte("https:")})))
		default:
			fmt.Printf("warning: ignoring unknown extension %s\n", ext)
		}
	}
	fmt.Println("Parsing templates...")
	t := template.New("base.html")
	t, err = t.ParseGlob(path.Join("themes", cfg.Theme.Name, "layouts/*"))
	if err != nil {
		return nil, fmt.Errorf("build.New: %w", err)
	}
	b := &Build{
		cfg:      cfg,
		start:    start,
		markdown: goldmark.New(goldmark.WithExtensions(exts...)),
		template: t,
	}
	if !cfg.Forge.Debug {
		b.minify = minify.New()
		b.minify.AddFunc("text/css", css.Minify)
		b.minify.AddFunc("text/html", html.Minify)
		b.minify.AddFunc("application/javascript", js.Minify)
	}
	return b, nil
}

func (b *Build) Run() error {
	if err := os.RemoveAll("build"); err != nil {
		return fmt.Errorf("build.Run: %w", err)
	}
	if err := utils.Mkdir("build"); err != nil {
		return fmt.Errorf("build.Run: %w", err)
	}
	if b.cfg.Forge.Debug {
		fmt.Println("Generating debug assets...")
		if err := utils.WriteFile(path.Join("build", "index.html"), indexHTML); err != nil {
			return fmt.Errorf("build.All: %w", err)
		}
		if err := utils.WriteFile(path.Join("build", "debug.js"), debugJS); err != nil {
			return fmt.Errorf("build.All: %w", err)
		}
	}
	fmt.Println("Processing public directory...")
	if err := b.processPublicDir(path.Join("themes", b.cfg.Theme.Name, "public"), "build"); err != nil {
		return fmt.Errorf("build.Run: %w", err)
	}
	if err := b.processPublicDir("public", "build"); err != nil {
		return fmt.Errorf("build.Run: %w", err)
	}
	fmt.Println("Generating content...")
	if err := b.buildContentDir("content"); err != nil {
		return fmt.Errorf("build.Run: %w", err)
	}
	dur := time.Since(b.start).Round(time.Microsecond * 100)
	fmt.Println("-------------------------------")
	fmt.Printf("Build completed in %s\n\n", dur)
	return nil
}

func (b *Build) processPublicDir(src string, dest string) error {
	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return fmt.Errorf("build.processPublicDir: %w", err)
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())
		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			return fmt.Errorf("build.processPublicDir: %w", err)
		}
		stat, ok := fileInfo.Sys().(*syscall.Stat_t)
		if !ok {
			return fmt.Errorf("failed to get raw syscall.Stat_t data for '%s'", sourcePath)
		}
		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := utils.Mkdir(destPath); err != nil {
				return fmt.Errorf("build.processPublicDir: %w", err)
			}
			if err := b.processPublicDir(sourcePath, destPath); err != nil {
				return fmt.Errorf("build.processPublicDir: %w", err)
			}
		default:
			if err := b.processPublicFile(sourcePath, destPath); err != nil {
				return fmt.Errorf("build.processPublicDir: %w", err)
			}
		}
		if err := os.Lchown(destPath, int(stat.Uid), int(stat.Gid)); err != nil {
			return fmt.Errorf("build.processPublicDir: %w", err)
		}
		isSymlink := entry.Mode()&os.ModeSymlink != 0
		if !isSymlink {
			if err := os.Chmod(destPath, entry.Mode()); err != nil {
				return fmt.Errorf("build.processPublicDir: %w", err)
			}
		}
	}
	return nil
}

func (b *Build) processPublicFile(srcFile string, dstFile string) error {
	out, err := os.Create(dstFile)
	if err != nil {
		return fmt.Errorf("build.processPublicFile: %w", err)
	}
	defer out.Close()
	var buf bytes.Buffer
	if err := b.execTmpl(srcFile, &buf); err != nil {
		return fmt.Errorf("build.processPublicFile: %w", err)
	}
	if err := b.minifyFile(srcFile, out, &buf); err != nil {
		return fmt.Errorf("build.processPublicFile: %w", err)
	}
	return nil
}

func (b *Build) execTmpl(srcFile string, w io.Writer) error {
	var err error
	switch path.Ext(srcFile) {
	case ".css", ".html", ".js":
		var t *template.Template
		t, err = template.ParseFiles(srcFile)
		if err != nil {
			return fmt.Errorf("build.execTmpl: %w", err)
		}
		err = t.Execute(w, map[string]interface{}{"Theme": b.cfg.Theme.Params})
	default:
		src, err := os.Open(srcFile)
		if err != nil {
			return fmt.Errorf("build.execTmpl: %w", err)
		}
		_, err = io.Copy(w, src)
	}
	if err != nil {
		return fmt.Errorf("build.execTmpl: %w", err)
	}
	return nil
}

func (b *Build) minifyFile(srcFile string, w io.Writer, r io.Reader) error {
	if b.cfg.Forge.Debug {
		_, err := io.Copy(w, r)
		if err != nil {
			return fmt.Errorf("build.minifyFile: %w", err)
		}
		return nil
	}
	var err error
	switch path.Ext(srcFile) {
	case ".css":
		err = b.minify.Minify("text/css", w, r)
	case ".html":
		err = b.minify.Minify("text/html", w, r)
	case ".js":
		err = b.minify.Minify("application/javascript", w, r)
	default:
		_, err = io.Copy(w, r)
	}
	if err != nil {
		return fmt.Errorf("build.minifyFile: %w", err)
	}
	return nil
}

func (b *Build) execTmplAndMinify(contentType string, w io.Writer, file string) error {
	t, err := template.ParseFiles(file)
	if err != nil {
		return fmt.Errorf("build.execTmplAndMinify: %w", err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, map[string]interface{}{"Theme": b.cfg.Theme.Params}); err != nil {
		return fmt.Errorf("build.execTmplAndMinify: %w", err)
	}
	if b.cfg.Forge.Debug {
		_, err = io.Copy(w, &buf)
	} else {
		switch contentType {
		case ".css":
			err = b.minify.Minify("text/css", w, &buf)
		case ".html":
			err = b.minify.Minify("text/html", w, &buf)
		case ".js":
			err = b.minify.Minify("application/javascript", w, &buf)
		default:
			_, err = io.Copy(w, &buf)
		}
	}
	if err != nil {
		return fmt.Errorf("build.execTmplAndMinify: %w", err)
	}
	return nil
}

func (b *Build) buildContentDir(dirName string) error {
	fis, err := ioutil.ReadDir(dirName)
	if err != nil {
		return fmt.Errorf("build.buildContentDir: %w", err)
	}
	for _, fi := range fis {
		if fi.IsDir() {
			if err := b.buildContentDir(path.Join(dirName, fi.Name())); err != nil {
				return fmt.Errorf("build.buildContentDir: %w", err)
			}
			continue
		}
		if err := b.buildContentPage(path.Join(dirName, fi.Name())); err != nil {
			return fmt.Errorf("build.buildContentDir: %w", err)
		}
	}
	return nil
}

func (b *Build) buildContentPage(page string) error {
	pagePath := "build/" + strings.TrimPrefix(page, "content/")
	pageDir, pageName := path.Split(pagePath)
	if !strings.HasPrefix(pageName, "index.") {
		nameParts := strings.Split(pageName, ".")
		pageDir = path.Join(pageDir, nameParts[0])
		pageName = "index.html"
	}
	if err := utils.Mkdir(pageDir); err != nil {
		return fmt.Errorf("build.buildContentPage: %w", err)
	}
	f, err := os.Create(path.Join(pageDir, strings.ReplaceAll(pageName, ".md", ".html")))
	if err != nil {
		return fmt.Errorf("build.buildContentPage: %w", err)
	}
	pathname := strings.TrimPrefix(pageDir, "build")
	t, err := b.template.Clone()
	if err != nil {
		return fmt.Errorf("build.buildContentPage: %w", err)
	}
	pageParams := make(map[string]interface{})
	if strings.HasSuffix(page, ".md") {
		pageContents, err := ioutil.ReadFile(page)
		if err != nil {
			return fmt.Errorf("build.page: %w", err)
		}
		parts := bytes.SplitN(pageContents, []byte("---"), 3)
		if len(parts) != 3 && len(parts) != 1 {
			return fmt.Errorf(`build.buildContentPage: maleformed content; front matter must be surrounded by "---"`)
		}
		body := parts[0]
		if len(parts) == 3 {
			body = parts[2]
			if err := yaml.Unmarshal(parts[1], &pageParams); err != nil {
				return fmt.Errorf("build.buildContentPage: %w", err)
			}
		}
		var buf bytes.Buffer
		if err := b.markdown.Convert(body, &buf); err != nil {
			return fmt.Errorf("build.buildContentPage: %w", err)
		}
		t, err = t.Parse(fmt.Sprintf(`{{ define "body" }}%s{{ end }}`, buf.String()))
		if err != nil {
			return fmt.Errorf("build.buildContentPage: %w", err)
		}
	} else if strings.HasSuffix(page, ".html") {
		t, err = t.ParseFiles(page)
		if err != nil {
			return fmt.Errorf("build.buildContentPage: %w", err)
		}
	} else {
		return fmt.Errorf("build.page: invalid content: %s", page)
	}
	data := map[string]interface{}{
		"Theme":    b.cfg.Theme.Params,
		"Page":     pageParams,
		"Forge":    b.cfg.Forge,
		"Pathname": pathname,
	}
	if b.cfg.Forge.Debug {
		if err := t.Execute(f, data); err != nil {
			return fmt.Errorf("build.buildContentPage: %w", err)
		}
	} else {
		var buf bytes.Buffer
		if err := t.Execute(&buf, data); err != nil {
			return fmt.Errorf("build.buildContentPage: %w", err)
		}
		if err := b.minify.Minify("text/html", f, &buf); err != nil {
			return fmt.Errorf("build.buildContentPage: %w", err)
		}
	}
	return nil
}

var debugJS = []byte(`
const ws = new WebSocket('ws://' + window.location.host + '/hot')
ws.onmessage = e => {
	if (e.data === 'reload') return window.location.reload()
	const overlay = document.createElement('div')
	overlay.style = 'position: fixed; left: 0; right: 0; top: 0; bottom: 0; background: #000c; color: #e77; font-size: 18px'
	const msg = document.createElement('div')
	msg.style = 'width: 100%; max-width: 600px; margin: auto; margin-top: 5vh; line-height: 200%; padding: 0 20px;'
	msg.innerText = e.data
	overlay.appendChild(msg)
	document.body.appendChild(overlay)
}
ws.onopen = () => ws.send('loaded')
`)

var indexHTML = []byte(`
<html>
<head>
<script src="/debug.js"></script>
</head>
</html>
`)
