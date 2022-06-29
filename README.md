# FileWalker
go file walker

## contributor
MurInj

## function
- file walker
- file state watch

## install
```shell
go get github.com/murInJ/FileManager
```

## quick start
you can get file list and watch your path
```go
    fm := FileWalker.NewFileManager("your path")
	//fm.SetDebug(true)
	fm.GetFileList()
	fmt.Println(len(fm.FileList))
```