package utils

import (
	"fmt"
	"io/ioutil"
	"os"
)

// func CopyDirectory(scrDir string, dest string, prsr *parser.Parser) error {
// 	entries, err := ioutil.ReadDir(scrDir)
// 	if err != nil {
// 		return err
// 	}
// 	for _, entry := range entries {
// 		sourcePath := filepath.Join(scrDir, entry.Name())
// 		destPath := filepath.Join(dest, entry.Name())
// 		fileInfo, err := os.Stat(sourcePath)
// 		if err != nil {
// 			return err
// 		}
// 		stat, ok := fileInfo.Sys().(*syscall.Stat_t)
// 		if !ok {
// 			return fmt.Errorf("failed to get raw syscall.Stat_t data for '%s'", sourcePath)
// 		}
// 		switch fileInfo.Mode() & os.ModeType {
// 		case os.ModeDir:
// 			if err := Mkdir(destPath); err != nil {
// 				return err
// 			}
// 			if err := CopyDirectory(sourcePath, destPath, prsr); err != nil {
// 				return err
// 			}
// 		default:
// 			if err := copyFile(sourcePath, destPath, prsr); err != nil {
// 				return err
// 			}
// 		}
// 		if err := os.Lchown(destPath, int(stat.Uid), int(stat.Gid)); err != nil {
// 			return err
// 		}
// 		isSymlink := entry.Mode()&os.ModeSymlink != 0
// 		if !isSymlink {
// 			if err := os.Chmod(destPath, entry.Mode()); err != nil {
// 				return err
// 			}
// 		}
// 	}
// 	return nil
// }

// func copyFile(srcFile string, dstFile string, prsr *parser.Parser) error {
// 	out, err := os.Create(dstFile)
// 	if err != nil {
// 		return fmt.Errorf("utils.copyFile: %w", err)
// 	}
// 	defer out.Close()
// 	in, err := os.Open(srcFile)
// 	if err != nil {
// 		return fmt.Errorf("utils.copyFile: %w", err)
// 	}
// 	defer in.Close()
// 	switch path.Ext(srcFile) {
// 	case ".css":
// 		err = prsr.Minify("text/css", out, in)
// 	case ".html":
// 		err = prsr.Minify("text/html", out, in)
// 	case ".js":
// 		err = prsr.Minify("application/javascript", out, in)
// 	default:
// 		_, err = io.Copy(out, in)
// 	}
// 	if err != nil {
// 		return fmt.Errorf("utils.copyFile: %w", err)
// 	}
// 	return nil
// }

func Mkdir(dirs ...string) error {
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("utils.Mkdir: %w", err)
		}
	}
	return nil
}

func WriteFile(path string, data []byte) error {
	return ioutil.WriteFile(path, data, 0755)
}
