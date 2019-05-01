package asynchronousFileIO

import (
	"io"
)

type AFIOInterface interface {
	Load(r io.Reader) error
	Save(w io.Writer) error
	GetFileName() string
	Refresh(data AFIOInterface)
	Clone() AFIOInterface
}

type DataSource interface {
	ReadFrom(fileName string) (io.Reader, error)
	CloseReader(reader io.Reader) error
	WriteTo(fileName string) (io.Writer, error)
	CloseWrite(writer io.Writer) error
}
