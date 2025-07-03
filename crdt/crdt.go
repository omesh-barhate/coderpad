package crdt

import "fmt"

type CRDT interface {
	Insert(pos int, val string) (string, error)
	Delete(pos int) string
}

func TestCRDT(c CRDT) {
	fmt.Println(c.Insert(1, "a"))
}
