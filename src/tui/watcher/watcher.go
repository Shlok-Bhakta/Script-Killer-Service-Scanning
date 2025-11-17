package watcher

import (
	"context"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

type FileChangeMsg struct {
	Path string
}

type Watcher struct {
	watcher   *fsnotify.Watcher
	targetDir string
}

func New(targetDir string) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Watcher{
		watcher:   w,
		targetDir: targetDir,
	}, nil
}

func (w *Watcher) Start(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		if err := w.addWatches(w.targetDir); err != nil {
			return err
		}

		debounce := time.NewTimer(500 * time.Millisecond)
		debounce.Stop()
		pending := false

		for {
			select {
			case <-ctx.Done():
				return nil
			case event, ok := <-w.watcher.Events:
				if !ok {
					return nil
				}

				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove) != 0 {
					if !pending {
						pending = true
						debounce.Reset(500 * time.Millisecond)
					}
				}

			case <-debounce.C:
				if pending {
					pending = false
					return FileChangeMsg{Path: w.targetDir}
				}

			case err, ok := <-w.watcher.Errors:
				if !ok {
					return nil
				}
				if err != nil {
					continue
				}
			}
		}
	}
}

func (w *Watcher) addWatches(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if filepath.Base(path) == ".git" ||
			filepath.Base(path) == "node_modules" ||
			filepath.Base(path) == "vendor" ||
			filepath.Base(path) == "__pycache__" {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return w.watcher.Add(path)
		}

		return nil
	})
}

func (w *Watcher) Close() error {
	return w.watcher.Close()
}
