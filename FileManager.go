package asynchronousFileIO

type FileManager interface {
	Load(fileName string, t uint64)
	Save(data AFIOInterface, t uint64)
	GetData(fileName string) AFIOInterface
	saveWorker(data AFIOInterface, t uint64)
	loadWorker(fileName string, t uint64)
}

type saveCmd struct {
	t uint64
	data AFIOInterface
}

type loadCmd struct {
	t uint64
	fileName string
}