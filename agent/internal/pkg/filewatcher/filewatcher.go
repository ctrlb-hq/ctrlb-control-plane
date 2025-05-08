package filewatcher

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/adapters"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/client"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/pkg/logger"
	"github.com/fsnotify/fsnotify"
)

type FileWatcher struct {
	filePath string
	adapter  adapters.Adapter
	watcher  *fsnotify.Watcher
	done     chan struct{}
	wg       sync.WaitGroup
}

func NewFileWatcher(filePath string, adapter adapters.Adapter) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &FileWatcher{
		filePath: filePath,
		adapter:  adapter,
		watcher:  watcher,
		done:     make(chan struct{}),
	}, nil
}

func (fw *FileWatcher) onFileChange() {
	logger.Logger.Info(fmt.Sprintf("Config modified in path: %s, restarting engine...", fw.filePath))
	fw.adapter.UpdateConfig()
}

func (fw *FileWatcher) onFileRecreated() {
	logger.Logger.Info(fmt.Sprintf("Config recreated: %s", fw.filePath))
	fw.adapter.UpdateConfig()
}

func (fw *FileWatcher) handleFileDeletionError() {
	logger.Logger.Error(fmt.Sprintf("Warning: Config no longer exists: %s", fw.filePath))
}

func (fw *FileWatcher) Start() error {
	if err := fw.watcher.Add(fw.filePath); err != nil {
		return err
	}

	fw.wg.Add(1)
	go fw.watchLoop()

	return nil
}

func (fw *FileWatcher) Stop() {
	close(fw.done)
	fw.wg.Wait()
	fw.watcher.Close()
}

func (fw *FileWatcher) watchLoop() {
	defer fw.wg.Done()

	fileExists := true
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-fw.done:
			return
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}

			client.InformBackendConfigFileChanged(nil)

			switch {
			case event.Op&fsnotify.Write == fsnotify.Write && fileExists:
				fw.onFileChange()

			case event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename:
				fileExists = false
				fw.handleFileDeletionError()
				fw.watcher.Remove(fw.filePath)
			}

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			logger.Logger.Error(fmt.Sprintf("Watcher error: %v", err))

		case <-ticker.C:
			if !fileExists {
				if _, err := os.Stat(fw.filePath); err == nil {
					fileExists = true
					if err := fw.watcher.Add(fw.filePath); err == nil {
						fw.onFileRecreated()
					}
				}
			}
		}
	}
}

func WatchFile(filePath string, adapter adapters.Adapter) error {
	watcher, err := NewFileWatcher(filePath, adapter)
	if err != nil {
		return err
	}

	blockingChan := make(chan struct{})

	go func() {
		<-blockingChan
		watcher.Stop()
	}()

	return watcher.Start()
}
