package gstool

import "github.com/fsnotify/fsnotify"

type FileWatchStruct struct {
	DirPath   string
	Fun       func(event fsnotify.Event)
	closeChan chan int
	isStop    bool
}

func NewFileWatch(dirPath string, fun func(event fsnotify.Event)) FileWatchStruct {
	return FileWatchStruct{DirPath: dirPath, Fun: fun, closeChan: make(chan int, 1)}
}
func (h *FileWatchStruct) Start() error {
	watcher, watcherErr := fsnotify.NewWatcher()
	if watcherErr != nil {
		return watcherErr
	}
	defer func(watcher *fsnotify.Watcher) {
		h.isStop = true
		_ = watcher.Close()
	}(watcher)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				h.Fun(event)
			case watchErr, ok := <-watcher.Errors:
				if !ok {
					return
				}
				h.closeChan <- 1
				FmtPrintlnLogTime(`监听`+h.DirPath+`报错：%s`, watchErr.Error())
			}
		}
	}()

	watchErr := watcher.Add(h.DirPath)
	if watchErr != nil {
		return watchErr
	}
	<-h.closeChan
	return nil
}

func (h *FileWatchStruct) Stop() {
	if h.isStop {
		return
	}
	h.closeChan <- 1
}
