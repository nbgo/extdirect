package main

import (
	"github.com/zenazn/goji"
	"github.com/hypebeast/gojistatic"
	"github.com/nbgo/extdirect"
	"reflect"
	"net/http"
	"github.com/zenazn/goji/web"
	"fmt"
	"time"
	"errors"
)

type GetDataRequest struct {
	Page int
	Start int
	Limit int
	Sort []SortDescriptor
	Filter []FilterDescriptor
	Model string
}

type SortDescriptor struct {
	Property  string
	Direction string
}

type FilterDescriptor struct {
	Property string
	Value    bool
}

type GetDataResponse struct {
	Total int `json:"total"`
	Records []interface{} `json:"records"`
}

type User struct {
	Id int `json:"id"`
	Text string `json:"text"`
}

type Db struct {
	Ctx *web.C
	Req *http.Request
}
func (this Db) GetRecords(r *GetDataRequest) (*GetDataResponse, error) {
	fmt.Printf("Hello from GetRecords(): model=%v start=%v limit=%v\n", r.Model, r.Start, r.Limit)
	result := &GetDataResponse{
		Total: 2,
		Records: []interface{} {
			&User{1, "Bob"},
			&User{2, "Alice"},
		},
	}
	return result, nil
}
func (this Db) Test() {
	fmt.Println("Hello from Test()")
	time.Sleep(30 * time.Millisecond)
}
func (this Db) TestEcho1(s string) string {
	fmt.Println("Hello from TestEcho1()")
	time.Sleep(30 * time.Millisecond)
	return s
}
func (this Db) TestEcho2(s string, n int, n2 int8, n3 int16, n4 int32, n5 int, s2 string) string {
	fmt.Println("Hello from TestEcho2()")
	time.Sleep(30 * time.Millisecond)
	return fmt.Sprintf("%v%v%v%v%v%v%v", s, n, n2, n3, n4, n5, s2)
}
func (this Db) TestException1() {
	panic("Error example #1")
}
func (this Db) TestException2() error {
	return errors.New("Error example #2")
}
func (this Db) TestException3() (string, error) {
	return "test", errors.New("Error example #3")
}
func (this Db) TestException4() {
	panic(errors.New("Error example #4"))
}

func main() {
	extdirect.Provider.RegisterAction(reflect.TypeOf(Db{}))
	goji.Get(extdirect.Provider.Url, extdirect.Api(extdirect.Provider))
	goji.Post(extdirect.Provider.Url, extdirect.ActionsHandlerCtx(extdirect.Provider))
	goji.Use(gojistatic.Static("public", gojistatic.StaticOptions{SkipLogging:true}))
	goji.Serve()
}