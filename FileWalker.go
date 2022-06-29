package FileWalker

import (
	"fmt"
	"github.com/fatih/color"
	"io/ioutil"
	"log"
	"math/rand"
	"time"
)

type FileWalker interface {
}

type FileManager struct {
	fileTree       *fileTree
	fileChannel    chan string
	controlChannel chan int
	readChannel    chan int
	FileList       []string
	isdebug        bool
	maxWorker      int
	activeWorker   int
}

func (m *FileManager) SetDebug(debug bool) {
	m.isdebug = debug
}

func NewFileManager(root string, maxWorker int) *FileManager {
	node := newNode(root, true)
	return &FileManager{
		fileTree:       newFileTree(node),
		fileChannel:    make(chan string, 250),
		controlChannel: make(chan int, 10),
		readChannel:    make(chan int, 5),
		isdebug:        false,
		maxWorker:      maxWorker,
	}
}

func (m *FileManager) getfile() {
	for {
		select {
		case file := <-m.fileChannel:
			m.FileList = append(m.FileList, file)
			if m.isdebug {
				fmt.Printf("%s add path %s\n",
					color.New(color.FgHiCyan).Sprintf("FileWalker:"),
					color.New(color.FgYellow).Sprintf(file))
			}
		case end := <-m.readChannel:
			if end == 0 {
				for file := range m.fileChannel {
					m.FileList = append(m.FileList, file)
					if m.isdebug {
						fmt.Printf("%s add path %s\n",
							color.New(color.FgHiCyan).Sprintf("FileWalker:"),
							color.New(color.FgYellow).Sprintf(file))
					}
				}
				break
			}
		}

	}
}

func (m *FileManager) GetFileList() {
	for i := 0; i < m.maxWorker; i++ {
		m.activeWorker += 1
		walker := newWalker(m)
		walker.walk()
	}
	go m.getfile()
	for {
		select {
		case s := <-m.controlChannel:
			if s == 1 {
				m.activeWorker += 1
				walker := newWalker(m)
				walker.walk()
			}
		default:
			if m.activeWorker == 0 {
				m.readChannel <- 0
				break
			}
		}
	}
}

type walker struct {
	fileManager *FileManager
	currentNode *node
}

func newWalker(manager *FileManager) *walker {
	return &walker{
		fileManager: manager,
	}
}

func (w *walker) walk() {
	w.currentNode = w.fileManager.fileTree.root
	prefix := w.currentNode.name

	for {
		if w.currentNode.isMark {
			w.fileManager.controlChannel <- 0
			w.fileManager.activeWorker -= 1
			return
		}

		w.currentNode.lock.Lock()
		fileList, err := ioutil.ReadDir(prefix)
		if err != nil {
			log.Println("wailker error: ", err.Error())
		}
		w.currentNode.expand(fileList)
		validFileList, validDirList := w.currentNode.validChildrenList()

		for _, file := range validFileList {
			file.isMark = true
			w.fileManager.fileChannel <- file.name
		}

		if len(validDirList) == 0 {
			w.currentNode.isMark = true
			w.fileManager.fileChannel <- prefix
			w.currentNode.lock.Unlock()
			break
		} else if len(validDirList) == 1 {
			w.currentNode.lock.Unlock()
			w.currentNode = validDirList[0]
			prefix = w.currentNode.name
			w.currentNode.visitTime += 1
		} else {
			rand.Seed(time.Now().Unix())
			randomIndex1 := rand.Intn(len(validDirList))
			randomIndex2 := rand.Intn(len(validDirList))
			for randomIndex1 == randomIndex2 {
				randomIndex1 = rand.Intn(len(validDirList))
				randomIndex2 = rand.Intn(len(validDirList))
			}
			w.currentNode.lock.Unlock()
			if validDirList[randomIndex1].visitTime > validDirList[randomIndex2].visitTime {
				w.currentNode = validDirList[randomIndex2]
			} else {
				w.currentNode = validDirList[randomIndex1]
			}
			prefix = w.currentNode.name
			w.currentNode.visitTime += 1
		}
	}
	w.fileManager.controlChannel <- 1
	w.fileManager.activeWorker -= 1
}
