package epubee

import (
	"math/rand"
	"strings"
	"fmt"
	"galen.demo.go/crawler/support"
	"time"
	"bytes"
	"encoding/json"
	"sync"
	"github.com/PuerkitoBio/goquery"
	"regexp"
	"errors"
	"math"
	"go/ast"
	"strconv"
	"net/http"
	"galen.demo.go/Set"
	"crypto/md5"
	"encoding/binary"
	"io/ioutil"
	"os"
)

type Book struct {
	Title       string
	Author      string
	Level       string
	LevelNum    float64
	Price       string
	ReadUrl     string
	Size        string
	Bid         string
	Format      string
	DownLoadUrl string
}

func (b *Book) String() string {
	return b.Title + "   " + b.Author + "   " + b.Level + "   " + b.Price + "   " + "   " + b.Format + "   " + b.Size  + "\n" + b.DownLoadUrl + "\n" + b.ReadUrl;
}

const max = 255
var random = rand.New(rand.NewSource(time.Now().UnixNano()))
var ips = [4]int{random.Intn(max), random.Intn(max), random.Intn(max), random.Intn(max)}

var (
	IP,ID string
	count int
	nextMu sync.Mutex
)

var ErrTimeOut = errors.New("time out")
var ErrResponse = errors.New("response error")
var ErrDecode = errors.New("decode error")

func nextIp(index int) {
	if ips[index] < max {
		ips[index] = ips[index] + 1
	} else if index == 0 {
		ips[0] = 0
	} else {
		nextIp(index - 1)
		return
	}
	for ; index < 3; index++ {
		ips[index+1] = 0;
	}
}

func init()  {
//	NextContext()
}

func Used()  {
	if count == 2 {
		NextContext()
	}else {
		count++
	}
}

func NextContext()  {
	nextContextCircle(10)
}
func nextContextCircle(retry int)  {
	nextMu.Lock()
	defer nextMu.Unlock()
	IP = NewIP()
	id,err := NewID(IP)
	if err != nil {
		if retry > 0 {
			nextContextCircle(retry - 1)
		}else{
			panic(err)
		}
	}
	if len(id) < 2 {
		nextContextCircle(retry - 1)
	}
	ID = id
	count = 0
}

func NewIP() string {
	nextIp(3)
	return strings.Replace(strings.Trim(fmt.Sprint(ips), "[]"), " ", ".", -1)
}
func handleRequestError(resp * http.Response,err *error) error{
	if err != nil {
		fmt.Println(strconv.Itoa(resp.StatusCode) + ":%s",err)
	}
	if resp.StatusCode < 300 && resp.StatusCode >199 {
		return nil
	} else if resp.StatusCode < 500 && resp.StatusCode >399 {
		return ErrTimeOut
	}else{
		return ErrResponse
	}
}
func NewID(ip string) (string,error) {
	resp, err := support.Request("POST", "http://cn.epubee.com/keys/genid_with_localid.asmx/genid_with_localid", "{localid:'0'}", nil, map[string]string{
		"X-Forwarded-For":  ip,
		"Content-Type":     "application/json",
		"X-Requested-With": "XMLHttpRequest",
	}, 0, 0)
	err = handleRequestError(resp,&err)
	if err != nil {
		return "",err
	}
	s,origin,err := decodeBodyD(resp)
	if err != nil {
		fmt.Errorf(*origin + ":%s",err)
		return "",ErrResponse
	}
	ss,_ :=  (*s).([]interface{})
	sss,_ := ss[0].(map[string]interface{})
	return  strconv.FormatFloat(sss["ID"].(float64),'f',-1,64),nil
}

func decodeBodyD(resp * http.Response)  ( r *interface{},origin *string, err error){
	defer func() {
		if p := recover(); p != nil{
			err = errors.New(fmt.Sprint(p))
		}
	}()
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	resp.Body.Close()
	str := string(buf.String())
	origin = & str
	var bodyResult map[string]interface{}
	json.Unmarshal(buf.Bytes(), &bodyResult)
	s := bodyResult["d"]
	r = &s
	return
}

func QueryBook(bookName string) (bookList *[]Book,err error) {
	resp, err := support.Request("POST", "http://cn.epubee.com/keys/get_ebook_list_search.asmx/getSearchList", "{skey:'" + bookName + "'}", nil, map[string]string{
		"Content-Type":     "application/json",
		"X-Requested-With": "XMLHttpRequest",
	}, 0, 0)

	err = handleRequestError(resp,&err)
	if err != nil {
		return nil,err
	}
	s,origin,err := decodeBodyD(resp)
	if err != nil {
		fmt.Errorf(*origin + ":%s",err)
		return nil,ErrResponse
	}
	defer func() {
		if p := recover(); p != nil{
			fmt.Errorf(fmt.Sprint(p))
			err = ErrDecode
		}
	}()
	var books = make([]Book,len((*s).([]interface{})))
	bookList = &books
	for i,boo := range  (*s).([]interface{}) {
		boom := boo.(map[string]interface{})
		title := boom["Title"].(string);
		index := strings.LastIndex(title,"[")
		var book Book
		if index > -1 {
			book = Book{
				Title: title[0:index],
				Bid:boom["BID"].(string),
				Format:title[index + 2 : len(title) - 1],
			}
		}else {
			book = Book{
				Title: title[0:],
				Bid:boom["BID"].(string),
			}
		}
		books[i] = book
	}
	return
}
func AddBook(book *Book,ip string,id string) (err error) {
	resp, err := support.Request("POST", "http://cn.epubee.com/app_books/addbook.asmx/online_addbook", "{bookid:'" + book.Bid + "',uid:" + id + ",act:'search'}", nil, map[string]string{
		"X-Forwarded-For": ip,
		"Content-Type": "application/json",
		"X-Requested-With": "XMLHttpRequest",
		"Accept-Language": "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2",
		"Accept-Encoding": "gzip, deflate",
	}, 0, 0)
	err = handleRequestError(resp,&err)
	if err != nil {
		return err
	}
	s,origin,err := decodeBodyD(resp)
	if err != nil {
		fmt.Errorf(*origin + ":%s",err)
		return ErrResponse
	}
	if(len((*s).([]interface{})) != 0){
		fmt.Errorf(*origin)
		return ErrResponse
	}
	return nil
}
func MyBook(ip string,id string) (books *[]Book,err error) {
	resp,err := support.Request("GET", "http://cn.epubee.com/files.aspx", "", map[string]string{
		"identify": id,
	}, map[string]string{
		"X-Forwarded-For": ip,
	}, 0, 0)
	err = handleRequestError(resp,&err)
	if err != nil {
		return nil,err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	resp.Body.Close()
	/*fmt.Println(buf.String())
	strings.NewReader(buf.String())*/

	doc, err := goquery.NewDocumentFromReader(buf)
	if err != nil {
		fmt.Errorf(buf.String() + ":%s",err)
		return nil,ErrResponse
	}
	defer func() {
		if p := recover(); p != nil{
			fmt.Errorf(fmt.Sprint(p))
			err = ErrDecode
		}
	}()
	sel  := "form table tr td div#centerContent div#centerRight div table.gv tr.parent"
	sele := doc.Find(sel)
	var bookList = make([]Book,(*sele).Size())
	books = &bookList
	(*sele).Each(func(i int, selection *goquery.Selection) {
		 title := selection.Find("td div.listEbook div.contentshow span#gvBooks_lblTitle_" + strconv.Itoa(i)).Text();
		 format := strings.Split(strings.TrimSpace(selection.Find("td div.listEbook div.contentshow span#gvBooks_lblExtensions_" + strconv.Itoa(i)).Text())," ")[0][1:]
		 size := selection.Find("td div.listEbook div.contentshow div table#gvBooks_gvBooks_child_" + strconv.Itoa(i) + " tr td.gvchild_first div.book_child span.list-filesize_k").Text();
		 reader,_ := selection.Find("td div.listEbook div.contentshow div table#gvBooks_gvBooks_child_" + strconv.Itoa(i) + " tr td.list_reader a.child_send").Attr("href");
		 reader = "http://cn.epubee.com/" + reader
		 bookList[i] = Book{
		 	Title:title,
		 	Format:format,
		 	Size:size,
		 	ReadUrl:reader,
		 }
	})
	return
}

func ReadBook(readUrl string,ip string,id string) (s *string,err error) {
	defer func() {
		if recover() != nil {
			err = errors.New("未知" + fmt.Sprint(recover()))
		}
	}()
	resp,err := support.Request("GET", readUrl, "", map[string]string{
		"identify": id,
	}, map[string]string{
		"X-Forwarded-For": ip,
		"Accept-Language": "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2",
		"Accept-Encoding": "gzip, deflate",
		"Upgrade-Insecure-Requests": "1",
	}, 0, 0)
	if err != nil || resp.StatusCode != 302 {
		fmt.Errorf(strconv.Itoa(resp.StatusCode) + ":%s",err)
		if resp.StatusCode < 500 || resp.StatusCode >499 {
			return nil,ErrTimeOut
		}else{
			return nil,ErrResponse
		}
	}
	defer func() {
		if p := recover(); p != nil{
			fmt.Errorf(fmt.Sprint(p))
			err = ErrDecode
		}
	}()
	location,_ := resp.Location()
	ss := string( regexp.MustCompile("&.*").ReplaceAll([]byte(strings.ReplaceAll(location.String(),"?book=", "")), []byte(".epub")))
	s = &ss
	return
}
func tryDo(action func() error,count int)  bool{
	err := action()
	if err == nil{
		return true
	}else if err == ErrTimeOut{
		time.Sleep(time.Duration(2)*time.Second)
	}else if err == ErrResponse{
		NextContext()
	}else if err == ErrDecode{
		panic(err)
	}else{
		panic(err)
	}

	if count <= 0 {
		return false
	}
	return tryDo(action,count - 1)
}

func tryQueryBook(bookName string,count int) (*[]Book,bool)  {
	var bookList *[]Book
	var err error
	if !tryDo(func() error {
		bookList,err = QueryBook(bookName)
		return err
	},count){
		return bookList, false
	}

	if len(*bookList) == 0 {
		index := math.Max(float64(strings.LastIndex(bookName,"(")),float64(strings.LastIndex(bookName,"（")))
		if index > 1 {
			return tryQueryBook(bookName[int(index):],count)
		}
	}
	return bookList,true
}

func Search(bookName string,filter ast.Filter) (*[]Book,bool) {
	fmt.Println("查询：" + bookName)
	bookList,ok := tryQueryBook(bookName,3)
	if !ok {
		return nil, false
	}
	if bookList == nil {
		return nil, true
	}
	var result = make([]Book,len(*bookList))
	for i,book := range *bookList  {
		if filter == nil || filter(book.Title) {
			fmt.Println("获取：" + bookName)
			b,ok := fillBook(&book)
			if ok {
				result[i] = *b
			}else {
				book.DownLoadUrl = "read timeOut"
				result[i] = book
			}
		}
	}
	return &result,true
}
func fillBook(book *Book) (*Book,bool) {
	defer Used()
	var err error
	if !tryDo(func() error {
		fmt.Println("添加：" + book.Title)
		err = AddBook(book,IP,ID)
		return err
	},3){
		return nil,false
	}
	var myBooks *[]Book
	if !tryDo(func() error {
		fmt.Println("我的：" + book.Title)
		myBooks,err = MyBook(IP,ID)
		return err
	},3){
		return nil, false
	}

	for _,myBook := range *myBooks {
		if strings.EqualFold(book.Title,myBook.Title) {
			book.ReadUrl = myBook.ReadUrl
			book.Size = myBook.Size
			book.Format = myBook.Format
			var location *string
			if !tryDo(func() error {
				fmt.Println("阅读：" + book.Title)
				location,err = ReadBook(myBook.ReadUrl,IP,ID)
				return err
			},3){
				return nil, false
			}else {
				book.DownLoadUrl = *location
				return book,true
			}
		}
	}
	return book,false
}

func Retrieve(bookName string,distinct bool,fileName string,download bool){
	running := true
	var lock sync.Mutex
	go func() {
		s := ""
		lock.Lock()
		if running{
			s = s + "·"
			fmt.Println("running"+s)
		}
		lock.Unlock()
		time.Sleep(time.Duration(3) * time.Second)
	}()
	defer func() {
		if p:= recover();p!=nil{
			fmt.Errorf("异常:%s",p)
		}
	}()
	nameSet := Set.New()
	hash := md5.New()
	var filter func(name string) bool
	if distinct {
		filter = func(name string) bool {
			return nameSet.Add(int(binary.LittleEndian.Uint32(hash.Sum([]byte(name)))))
		}
	}
	start := time.Now()
	fmt.Println("开始查询")
	books,ok := Search(bookName,filter)
	fmt.Println("查询结束.耗时" + time.Since(start).String())
	if ok {
		rs := "";
		for _,book := range *books {
			rs += book.String();
			rs += "\n";
		}
		fmt.Println(rs)
		ioutil.WriteFile(fileName,[]byte(rs),os.ModePerm)
	}else {
		fmt.Errorf("失败")
	}
}