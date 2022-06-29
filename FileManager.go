package FileManager

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"log"
	"sync"
)

type FileManager struct {
	fileTree       *node
	FileList       []string
	isdebug        bool
	writeLock      sync.Mutex
	walkerCount    sync.WaitGroup
	controlChannel chan int
	fileWatcher    *fsnotify.Watcher
	changeFileMap  map[string]int
}

func (m *FileManager) Close() {
	m.fileWatcher.Close()
}

func (m *FileManager) SetDebug(debug bool) {
	m.isdebug = debug
}

func NewFileManager(root string) *FileManager {
	watcher, err := fsnotify.NewWatcher()
	err = watcher.Add(root)
	if err != nil {
		log.Fatal(err)
	}
	FileManager := &FileManager{
		fileTree:       newNode(root, true),
		controlChannel: make(chan int, 50000),
		fileWatcher:    watcher,
		changeFileMap:  make(map[string]int),
	}
	go FileManager.onWatchFile()

	return FileManager
}

func (m *FileManager) GetFileList() {
	m.walkerCount.Add(1)
	walker := newWalker(m, m.fileTree)
	go walker.walk()

	m.walkerCount.Wait()
}

func (m *FileManager) onWatchFile() {

	for {
		select {
		case event, ok := <-m.fileWatcher.Events:
			if !ok {
				return
			}
			// 打印监听事件
			m.changeFileMap[event.Name] = 1
			if m.isdebug {
				fmt.Printf("%s %s %s\n",
					color.New(color.FgHiCyan).Sprintf("FileWalker:"),
					color.New(color.FgYellow).Sprintf(event.Name),
					color.New(color.FgGreen).Sprintf(event.Op.String()))
			}
		case _, ok := <-m.fileWatcher.Errors:
			if !ok {
				return
			}
		}
	}

}

func (m *FileManager) GetChangeFileList() []string {
	var l []string
	for k := range m.changeFileMap {
		l = append(l, k)
	}
	return l
}

func (m *FileManager) CleanChangeFileList() {
	m.changeFileMap = make(map[string]int)
}
