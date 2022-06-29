package FileManager

import (
	"io/fs"
)

type node struct {
	name     string
	children []*node
	isDir    bool
}

func newNode(name string, isdir bool) *node {
	return &node{
		name:  name,
		isDir: isdir,
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
		if child.isDir {
			dirList = append(dirList, child)
		} else {
			fileList = append(fileList, child)
		}
	}
	return fileList, dirList
}
