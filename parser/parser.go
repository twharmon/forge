package parser

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"syscall"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/twharmon/forge/utils"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

type Parser struct {
	md       goldmark.Markdown
	minifier *minify.M
}

type Config struct {
	Minify     bool
	Extensions []string
}

func New(cfg Config) *Parser {
	var exts []goldmark.Extender
	for _, ext := range cfg.Extensions {
		switch ext {
		case "footnote":
			exts = append(exts, extension.Footnote)
		case "linkify":
			exts = append(exts, extension.NewLinkify(extension.WithLinkifyAllowedProtocols([][]byte{[]byte("https:")})))
		default:
			fmt.Printf("warning: ignoring unknown extension %s\n", ext)
		}
	}
	p := &Parser{
		md: goldmark.New(goldmark.WithExtensions(exts...)),
	}
	if cfg.Minify {
		p.minifier = minify.New()
		p.minifier.AddFunc("text/css", css.Minify)
		p.minifier.AddFunc("text/html", html.Minify)
		p.minifier.AddFunc("application/javascript", js.Minify)
	}
	return p
}

func (p *Parser) Markdown(data []byte, buf io.Writer) error {
	if err := p.md.Convert(data, buf); err != nil {
		return fmt.Errorf("parser.Parse: %w", err)
	}
	return nil
}

func (p *Parser) Minify(mediatype string, w io.Writer, r io.Reader) error {
	if err := p.minifier.Minify(mediatype, w, r); err != nil {
		return fmt.Errorf("parser.Minify: %w", err)
	}
	return nil
}

func (p *Parser) CopyDirectory(scrDir string, dest string) error {
	entries, err := ioutil.ReadDir(scrDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(scrDir, entry.Name())
		destPath := filepath.Join(dest, entry.Name())
		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			return err
		}
		stat, ok := fileInfo.Sys().(*syscall.Stat_t)
		if !ok {
			return fmt.Errorf("failed to get raw syscall.Stat_t data for '%s'", sourcePath)
		}
		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := utils.Mkdir(destPath); err != nil {
				return err
			}
			if err := p.CopyDirectory(sourcePath, destPath); err != nil {
				return err
			}
		default:
			if err := p.copyFile(sourcePath, destPath); err != nil {
				return err
			}
		}
		if err := os.Lchown(destPath, int(stat.Uid), int(stat.Gid)); err != nil {
			return err
		}
		isSymlink := entry.Mode()&os.ModeSymlink != 0
		if !isSymlink {
			if err := os.Chmod(destPath, entry.Mode()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *Parser) copyFile(srcFile string, dstFile string) error {
	out, err := os.Create(dstFile)
	if err != nil {
		return fmt.Errorf("utils.copyFile: %w", err)
	}
	defer out.Close()
	in, err := os.Open(srcFile)
	if err != nil {
		return fmt.Errorf("utils.copyFile: %w", err)
	}
	defer in.Close()
	if p.minifier != nil {
		switch path.Ext(srcFile) {
		case ".css":
			err = p.Minify("text/css", out, in)
		case ".html":
			err = p.Minify("text/html", out, in)
		case ".js":
			err = p.Minify("application/javascript", out, in)
		default:
			_, err = io.Copy(out, in)
		}
	} else {
		_, err = io.Copy(out, in)
	}
	if err != nil {
		return fmt.Errorf("utils.copyFile: %w", err)
	}
	return nil
}
