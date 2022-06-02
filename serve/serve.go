package serve

import (
	"fmt"
	"io/ioutil"
	"log"
	"main/build"
	"main/config"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"github.com/twharmon/slices"
)

func Start() error {
	os.Setenv("DEBUG", "true")
	if err := build.All(); err != nil {
		fmt.Printf("serve.hot: %s\n", err)
	}
	var mu sync.Mutex
	var listeners []*websocket.Conn
	cfg, err := config.Get()
	if err != nil {
		return fmt.Errorf("serve.Start: %w", err)
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("serve.Start: %w", err)
	}
	defer watcher.Close()
	if err := watchAll(watcher, cfg); err != nil {
		return fmt.Errorf("serve.Start: %w", err)
	}
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove) > 0 {
					fmt.Println("change detected:", event.Name)
					if err := build.All(); err != nil {
						fmt.Printf("serve.hot: %s\n", err)
					} else {
						mu.Lock()
						for i := len(listeners) - 1; i >= 0; i-- {
							if listeners[i].WriteMessage(websocket.TextMessage, []byte("reload")) != nil {
								listeners = slices.Splice(listeners, i, 1)
							}
						}
						mu.Unlock()
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()
	var upgrader = websocket.Upgrader{}
	http.HandleFunc("/hot", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("upgrade failed: ", err)
			return
		}
		defer conn.Close()
		mu.Lock()
		listeners = append(listeners, conn)
		mu.Unlock()
		conn.ReadMessage()
		mu.Lock()
		listeners = slices.Filter(listeners, func(listener *websocket.Conn) bool { return listener != conn })
		mu.Unlock()
	})
	http.HandleFunc("/", handler)
	port := fmt.Sprintf(":%d", cfg.Port)
	url := fmt.Sprintf("http://localhost%s", port)
	fmt.Printf("Listening on %s...\n\n", url)
	go openbrowser(url)
	return http.ListenAndServe(port, nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, path.Join("build", r.URL.Path))
}

func openbrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		return
	}
	if err != nil {
		fmt.Println(err)
	}
}

func watchAll(w *fsnotify.Watcher, cfg *config.Config) error {
	if err := watch(w, "config.yml"); err != nil {
		return fmt.Errorf("serve.watchAll: %w", err)
	}
	if err := watch(w, "content"); err != nil {
		return fmt.Errorf("serve.watchAll: %w", err)
	}
	if err := watch(w, "public"); err != nil {
		return fmt.Errorf("serve.watchAll: %w", err)
	}
	if err := watch(w, path.Join("themes", cfg.Theme)); err != nil {
		return fmt.Errorf("serve.watchAll: %w", err)
	}
	return nil
}

func watch(w *fsnotify.Watcher, dir string) error {
	if err := w.Add(dir); err != nil {
		return fmt.Errorf("serve.watch: %w", err)
	}
	if fi, err := os.Stat(dir); err != nil {
		return fmt.Errorf("serve.watch: %w", err)
	} else if fi.IsDir() {
		fis, err := ioutil.ReadDir(dir)
		if err != nil {
			return fmt.Errorf("serve.watch: %w", err)
		}
		for _, fi := range fis {
			if fi.IsDir() {
				if err := watch(w, path.Join(dir, fi.Name())); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
