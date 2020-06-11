package epubee

import (
	"testing"
	"fmt"
	"bytes"
	"log"
)

func TestNewIP(t *testing.T) {
	var (
		buf    bytes.Buffer
		logger = log.New(&buf, "INFO: ", log.Lshortfile)

		infof = func(info string) {
			logger.Output(2, info)
		}
	)

	fmt.Println(NewIP())
	infof("Hello world")

	fmt.Print(&buf)
}
func TestNewID(t *testing.T) {
	s := NewIP()
	fmt.Println(s)
	fmt.Println(NewID(s))
}
func TestQueryBook(t *testing.T) {
	fmt.Println(QueryBook("道德经"))
}
func TestAddBook(t *testing.T) {
	books,_ := QueryBook("操作系统")
	fmt.Println(AddBook(&(*books)[0],IP,ID))
}
func TestMyBook(t *testing.T) {
	books,_ := QueryBook("操作系统")
	fmt.Println(AddBook(&(*books)[0],IP,ID))
	fmt.Println(AddBook(&(*books)[3],IP,ID))
	fmt.Println(MyBook(IP,ID))
}
func TestReadBook(t *testing.T) {
	NextContext()
	books,_ := QueryBook("操作系统")
	fmt.Println(AddBook(&(*books)[0],IP,ID))
	fmt.Println(AddBook(&(*books)[3],IP,ID))
	myBooks,_ := MyBook(IP,ID)
	fmt.Println(myBooks)
	s,_ := ReadBook((*myBooks)[0].ReadUrl,IP,ID)
	fmt.Println(*s)
	s,_ = ReadBook((*myBooks)[1].ReadUrl,IP,ID)
	fmt.Println(*s)
}
func TestSearch(t *testing.T) {
	fmt.Println(Search("深入理解计算机系统",nil))
}
func TestRetrieve(t *testing.T) {
	Retrieve("深入理解计算机系统",true,"test.txt",false)
}