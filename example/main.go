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
	gcontext "github.com/goji/context"
	"golang.org/x/net/context"
)

type GetDataRequest struct {
	Page   int
	Start  int
	Limit  int
	Sort   []SortDescriptor
	Filter []FilterDescriptor
	Model  string
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
	Total   int `json:"total"`
	Records []interface{} `json:"records"`
}

type User struct {
	Id   int `json:"id"`
	Text string `json:"text"`
}

type Db struct {
	Ctx context.Context
	Req *http.Request
}

func (this Db) GetRecords(r *GetDataRequest) (*GetDataResponse, error) {
	fmt.Printf("Hello from GetRecords(): model=%v start=%v limit=%v\n", r.Model, r.Start, r.Limit)
	result := &GetDataResponse{
		Total: 2,
		Records: []interface{}{
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
func (this Db) GetBasicInfo(uid int, foo string) map[string]interface{} {
	return map[string]interface{}{"success": true, "data":map[string]string{"company": "Sencha Inc.", "email":"aaron@sencha.com", "name":foo}}
}
func (this Db) UpdateBasicInfo(data map[string]string) (result *extdirect.DirectFormHandlerResult) {
	result = &extdirect.DirectFormHandlerResult{Success:true}
	if data["email"] == "aaron@sencha.com" {
		result.Success = false
		result.Errors = make(map[string]string, 0)
		result.Errors["email"] = "already exists"
	}
	return
}

func main() {
	extdirect.Provider.RegisterAction(reflect.TypeOf(Db{}))
	goji.Get(extdirect.Provider.URL, extdirect.API(extdirect.Provider))
	goji.Post(extdirect.Provider.URL, func(c web.C, w http.ResponseWriter, r *http.Request) {
		extdirect.ActionsHandlerCtx(extdirect.Provider)(gcontext.FromC(c), w, r)
	})
	goji.Use(gojistatic.Static("public", gojistatic.StaticOptions{SkipLogging:true}))
	goji.Serve()
}