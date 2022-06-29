package FileWalker

import (
	"io/fs"
	"sync"
)

type node struct {
	name      string
	children  []*node
	isMark    bool
	visitTime int
	isDir     bool
	lock      sync.Mutex
}

func newNode(name string, isdir bool) *node {
	var c []*node
	return &node{
		name:      name,
		children:  c,
		isMark:    false,
		visitTime: 0,
		isDir:     isdir,
	}
}

func (n *node) expand(fileList []fs.FileInfo) {
	if len(n.children) != 0 {
		return
	}

	for _, file := range fileList {
		n.children = append(n.children, newNode(n.name+"/"+file.Name(), file.IsDir()))
	}
}

func (n *node) validChildrenList() ([]*node, []*node) {
	var fileList []*node
	var dirList []*node
	for _, child := range n.children {
		if !child.isMark {
			if child.isDir {
				dirList = append(dirList, child)
			} else {
				fileList = append(fileList, child)
			}
		}
	}
	return fileList, dirList
}

type fileTree struct {
	root *node
}

func newFileTree(node *node) *fileTree {
	return &fileTree{root: node}
}
