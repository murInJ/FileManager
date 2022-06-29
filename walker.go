package FileManager

import (
	"fmt"
	"github.com/fatih/color"
	"io/ioutil"
)

type walker struct {
	fileManager *FileManager
	currentNode *node
}

func newWalker(manager *FileManager, currentNode *node) *walker {
	return &walker{
		fileManager: manager,
		currentNode: currentNode,
	}
}

func (w *walker) appendFile(name string) {
	w.fileManager.writeLock.Lock()
	w.fileManager.FileList = append(w.fileManager.FileList, name)
	w.fileManager.writeLock.Unlock()
}

func (w *walker) walk() {

	fileList, _ := ioutil.ReadDir(w.currentNode.name)
	w.currentNode.expand(fileList)
	validFileList, validDirList := w.currentNode.validChildrenList()

	for _, file := range validFileList {
		w.appendFile(file.name)
		if w.fileManager.isdebug {
			fmt.Printf("%s add path %s\n",
				color.New(color.FgHiCyan).Sprintf("FileWalker:"),
				color.New(color.FgYellow).Sprintf(file.name))
		}
	}

	for _, dir := range validDirList {
		w.appendFile(dir.name)
		if w.fileManager.isdebug {
			fmt.Printf("%s add path %s\n",
				color.New(color.FgHiCyan).Sprintf("FileWalker:"),
				color.New(color.FgYellow).Sprintf(dir.name))
		}
		w.fileManager.walkerCount.Add(1)
		walker := newWalker(w.fileManager, dir)
		go walker.walk()
	}
	w.fileManager.walkerCount.Done()
}
