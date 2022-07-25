package FileManager

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"io/ioutil"
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
	changeFileMap  map[string]fsnotify.Event
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
		changeFileMap:  make(map[string]fsnotify.Event),
	}
	go FileManager.onWatchFile()

	return FileManager
}

func (m *FileManager) GetFileList() {
	m.FileList = []string{}
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

			if event.Op.String() == "CREATE" {
				m.addFile(event.Name)
			} else if event.Op.String() == "REMOVE" || event.Op.String() == "RENAME" {
				m.removeFile(event.Name)
			}

			// 打印监听事件
			m.changeFileMap[event.Name] = event
			if m.isdebug {
				fmt.Printf("%s %s %s\n",
					color.New(color.FgHiCyan).Sprintf("FileWalker:"),
					color.New(color.FgGreen).Sprintf(event.Op.String()),
					color.New(color.FgYellow).Sprintf(event.Name))

			}
		case _, ok := <-m.fileWatcher.Errors:
			if !ok {
				return
			}
		}
	}

}

func (m *FileManager) GetChangeFileList() []fsnotify.Event {
	var l []fsnotify.Event
	for k := range m.changeFileMap {
		l = append(l, m.changeFileMap[k])
	}
	return l
}

func (m *FileManager) CleanChangeFileList() {
	m.changeFileMap = make(map[string]fsnotify.Event)
}

func (m *FileManager) addFile(path string) {
	m.FileList = append(m.FileList, path)
}

func (m *FileManager) removeFile(path string) {
	j := 0
	for _, val := range m.FileList {
		if val == path {
			m.FileList[j] = val
			j++
		}
	}
	m.FileList = m.FileList[:j]
}

func (m *FileManager) ExportFileList_JSON(path ...string) {
	name := "FileList.json"
	if len(path) != 0 {
		name = path[0] + "\\" + name
	} else {
		name = m.fileTree.name + "\\" + name
	}
	jsonObject, _ := json.Marshal(m.FileList)
	_ = ioutil.WriteFile(name, jsonObject, 0666)
}
