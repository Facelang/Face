package prog

import (
	"fmt"
)

type FilePos struct {
	Filename          string
	Col, Line, Offset int
}

func (i *FilePos) String() string {
	return fmt.Sprintf("行: %d, 列: %d, 文件名：%s", i.Line+1, i.Col+1, i.Filename)
}
