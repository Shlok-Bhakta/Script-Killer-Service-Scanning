package watcher

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

type FileChangeMsg struct {
	Path string
}

type Watcher struct {
	watcher        *fsnotify.Watcher
	targetDir      string
	ignorePatterns []string
}

func New(targetDir string) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	ignorePatterns := []string{".git"}
	gitignorePath := filepath.Join(targetDir, ".gitignore")
	if patterns, err := parseGitignore(gitignorePath); err == nil {
		ignorePatterns = append(ignorePatterns, patterns...)
	}

	return &Watcher{
		watcher:        w,
		targetDir:      targetDir,
		ignorePatterns: ignorePatterns,
	}, nil
}

func parseGitignore(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var patterns []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimSuffix(line, "/")
		patterns = append(patterns, line)
	}

	return patterns, scanner.Err()
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
					if w.shouldIgnore(event.Name) {
						continue
					}
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

func (w *Watcher) shouldIgnore(path string) bool {
	relPath, err := filepath.Rel(w.targetDir, path)
	if err != nil {
		relPath = filepath.Base(path)
	}

	for _, pattern := range w.ignorePatterns {
		if matchPattern(relPath, pattern) || matchPattern(filepath.Base(path), pattern) {
			return true
		}
	}
	return false
}

func (w *Watcher) addWatches(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if !info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(w.targetDir, path)
		if err != nil {
			relPath = filepath.Base(path)
		}

		for _, pattern := range w.ignorePatterns {
			if matchPattern(relPath, pattern) || matchPattern(filepath.Base(path), pattern) {
				return filepath.SkipDir
			}
		}

		return w.watcher.Add(path)
	})
}

func matchPattern(path, pattern string) bool {
	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
		return strings.Contains(path, strings.Trim(pattern, "*"))
	}
	if strings.HasPrefix(pattern, "*") {
		return strings.HasSuffix(path, strings.TrimPrefix(pattern, "*"))
	}
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(path, strings.TrimSuffix(pattern, "*"))
	}

	return path == pattern || filepath.Base(path) == pattern ||
		strings.HasPrefix(path, pattern+"/") || strings.Contains(path, "/"+pattern+"/")
}

func (w *Watcher) Close() error {
	return w.watcher.Close()
}
