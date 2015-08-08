package extdirect

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	. "github.com/ahmetalpbalkan/go-linq"
	lgr "github.com/nbgo/logger"
	"reflect"
	"net/http"
	"strings"
	"encoding/json"
	"fmt"
	"time"
	"github.com/zenazn/goji/web"
	"errors"
	"github.com/Sirupsen/logrus"
)

var providerDebug = false
var l = lgr.Create()

type testLogger struct {}
func (this *testLogger) Print(v ...interface{}) {
	if err, errOk := v[0].(error); errOk {
		l.Error(err)
	} else {
		l.Debug(v[0])
	}
}

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

type Db struct {
	C *web.C
	R *http.Request
}
func (this Db) GetRecords(r *GetDataRequest) (string, error) {
	result := fmt.Sprintf("model=%v page=%v start=%v limit=%v sort=%v", r.Model, r.Page, r.Start, r.Limit, r.Sort)
	return result, nil
}
func (this Db) Test() string {
	time.Sleep(30 * time.Millisecond)
	var result string
	if this.C != nil {
		result += this.C.URLParams["test"]
	}
	if this.R != nil {
		result += this.R.Host
	}
	return result
}
func (this Db) TestEcho1(s string) (string, error) {
	time.Sleep(30 * time.Millisecond)
	return s, nil
}
func (this Db) TestEcho2(s string, n int, n2 int8, n3 int16, n4 int32, n5 int, s2 string) string {
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

func TestExtDirect(t *testing.T) {
	SetLogger(&testLogger{})
	logrus.SetLevel(logrus.FatalLevel)

	Convey("Default provider serialization", t, func() {
		value, err := Provider.Json()
		So(err, ShouldBeNil)
		So(value, ShouldEqual, "{\"type\":\"remoting\",\"url\":\"/directapi\",\"namespace\":\"DirectApi\",\"timeout\":30000,\"actions\":{}}")
	})

	Convey("Action registration", t, func() {
		provider := NewProvider()
		provider.LogMode(providerDebug)
		provider.RegisterAction(reflect.TypeOf(Db{}))
		Convey("One registered action with name 'Db' expected", func() {
			So(len(provider.Actions), ShouldEqual, 1)
			So(provider.Actions, ShouldContainKey, "Db")
			Convey("and has 8 methods", func() {
				So(len(provider.Actions["Db"]), ShouldEqual, 8)
				Convey("test", func() {
					method, exists, err := From(provider.Actions["Db"]).FirstBy(func(x T) (bool, error) {
						if m, ok := x.(DirectMethod); ok {
							return m.Name == "test", nil
						} else {
							return false, nil
						}
					})
					So(err, ShouldBeNil)
					So(exists, ShouldBeTrue)
					So(method.(DirectMethod).Name, ShouldEqual, "test")
					Convey("with no arguments", func() {
						So(method.(DirectMethod).Len, ShouldBeZeroValue)
					});
				})
				Convey("getRecords", func() {
					method, exists, err := From(provider.Actions["Db"]).FirstBy(func(x T) (bool, error) {
						if m, ok := x.(DirectMethod); ok {
							return m.Name == "getRecords", nil
						} else {
							return false, nil
						}
					})
					So(err, ShouldBeNil)
					So(exists, ShouldBeTrue)
					So(method.(DirectMethod).Name, ShouldEqual, "getRecords")
					Convey("with 1 argument", func() {
						So(method.(DirectMethod).Len, ShouldEqual, 1)
					});
				})
				Convey("testEcho1", func() {
					method, exists, err := From(provider.Actions["Db"]).FirstBy(func(x T) (bool, error) {
						if m, ok := x.(DirectMethod); ok {
							return m.Name == "testEcho1", nil
						} else {
							return false, nil
						}
					})
					So(err, ShouldBeNil)
					So(exists, ShouldBeTrue)
					So(method.(DirectMethod).Name, ShouldEqual, "testEcho1")
					Convey("with 1 argument", func() {
						So(method.(DirectMethod).Len, ShouldEqual, 1)
					});
				})
				Convey("testEcho2", func() {
					method, exists, err := From(provider.Actions["Db"]).FirstBy(func(x T) (bool, error) {
						if m, ok := x.(DirectMethod); ok {
							return m.Name == "testEcho2", nil
						} else {
							return false, nil
						}
					})
					So(err, ShouldBeNil)
					So(exists, ShouldBeTrue)
					So(method.(DirectMethod).Name, ShouldEqual, "testEcho2")
					Convey("with 7 arguments", func() {
						So(method.(DirectMethod).Len, ShouldEqual, 7)
					});
				})
			})
		})
		Convey("Action with methods serialization", func() {
			jsonText, err := provider.Json()
			So(err, ShouldBeNil)
			So(jsonText, ShouldEqual, `{"type":"remoting","url":"/directapi","namespace":"DirectApi","timeout":30000,"actions":{"Db":[{"name":"getRecords","len":1},{"name":"test","len":0},{"name":"testEcho1","len":1},{"name":"testEcho2","len":7},{"name":"testException1","len":0},{"name":"testException2","len":0},{"name":"testException3","len":0},{"name":"testException4","len":0}]}}`)
			javaScript, err2 := provider.JavaScript()
			So(err2, ShouldBeNil)
			So(javaScript, ShouldEqual, `Ext.ns("DirectApi");DirectApi.REMOTE_API={"type":"remoting","url":"/directapi","namespace":"DirectApi","timeout":30000,"actions":{"Db":[{"name":"getRecords","len":1},{"name":"test","len":0},{"name":"testEcho1","len":1},{"name":"testEcho2","len":7},{"name":"testException1","len":0},{"name":"testException2","len":0},{"name":"testException3","len":0},{"name":"testException4","len":0}]}}`)
		})
	})

	Convey("Request with single action call", t, func() {
		provider := NewProvider()
		provider.LogMode(providerDebug)
		provider.RegisterAction(reflect.TypeOf(Db{}))
		reqs := mustDecodeTransaction(strings.NewReader(`{"action":"Db","method":"test","data":null,"type":"rpc","tid":1}`))
		Convey("has one parsed request with correct fields", func() {
			So(len(reqs), ShouldEqual, 1)
			So(reqs[0].Action, ShouldEqual, "Db")
			So(reqs[0].Method, ShouldEqual, "test")
			So(reqs[0].Data, ShouldBeNil)
			So(reqs[0].Tid, ShouldEqual, 1)
			So(reqs[0].Type, ShouldEqual, "rpc")
			Convey("which is processed into 1 response with correct fields", func() {
				var resps = provider.processRequests(nil, nil, reqs)
				So(len(resps), ShouldEqual, 1)
				So(resps[0].Action, ShouldEqual, "Db")
				So(resps[0].Method, ShouldEqual, "test")
				So(resps[0].Result, ShouldBeEmpty)
				So(resps[0].Tid, ShouldEqual, 1)
				So(resps[0].Type, ShouldEqual, "rpc")
				So(resps[0].Message, ShouldBeEmpty)
				Convey("which is serialized correctly", func() {
					s, err := json.Marshal(resps)
					So(err, ShouldBeNil)
					// Response is always array even for single request
					So(string(s), ShouldEqual, `[{"type":"rpc","tid":1,"action":"Db","method":"test","message":"","result":""}]`)
				})
			})
		})
	})

	Convey("Request with multiple actions call", t, func() {
		provider := NewProvider()
		provider.LogMode(providerDebug)
		provider.RegisterAction(reflect.TypeOf(Db{}))
		reqs := mustDecodeTransaction(strings.NewReader(`[{"action":"Db","method":"testEcho1","data":["Hello!"],"type":"rpc","tid":1},{"action":"Db","method":"testEcho2","data":["Hello", 1, 2, 3, 4, null, null],"type":"rpc","tid":2}]`))
		Convey("has 2 parsed requests with correct fields", func() {
			So(len(reqs), ShouldEqual, 2)
			So(reqs[0].Action, ShouldEqual, "Db")
			So(reqs[0].Method, ShouldEqual, "testEcho1")
			dataArray, dataArrayOk := reqs[0].Data.([]interface{})
			So(dataArrayOk, ShouldBeTrue)
			So(len(dataArray), ShouldEqual, 1)
			So(dataArray[0], ShouldEqual, "Hello!")
			So(reqs[0].Tid, ShouldEqual, 1)
			So(reqs[0].Type, ShouldEqual, "rpc")

			So(reqs[1].Action, ShouldEqual, "Db")
			So(reqs[1].Method, ShouldEqual, "testEcho2")
			dataArray, dataArrayOk = reqs[1].Data.([]interface{})
			So(dataArrayOk, ShouldBeTrue)
			So(len(dataArray), ShouldEqual, 7)
			So(reqs[1].Tid, ShouldEqual, 2)
			So(reqs[1].Type, ShouldEqual, "rpc")
			Convey("which is concurrently processed into 2 responses with correct fields", func() {
				t1 := time.Now()
				resps := provider.processRequests(nil, nil, reqs)
				t2 := time.Now()
				So(t2.Sub(t1), ShouldBeLessThan, 35 * time.Millisecond)
				So(len(resps), ShouldEqual, 2)
				So(resps[0].Message, ShouldBeEmpty)
				So(resps[0].Type, ShouldEqual, "rpc")
				So(resps[0].Action, ShouldEqual, "Db")
				So(resps[0].Method, ShouldEqual, "testEcho1")
				So(resps[0].Result, ShouldEqual, "Hello!")
				So(resps[0].Tid, ShouldEqual, 1)

				So(resps[1].Message, ShouldBeEmpty)
				So(resps[1].Type, ShouldEqual, "rpc")
				So(resps[1].Action, ShouldEqual, "Db")
				So(resps[1].Method, ShouldEqual, "testEcho2")
				So(resps[1].Result, ShouldEqual, "Hello12340")
				So(resps[1].Tid, ShouldEqual, 2)
				Convey("which is serialized correctly", func() {
					s, err := json.Marshal(resps)
					So(err, ShouldBeNil)
					So(string(s), ShouldContainSubstring, `{"type":"rpc","tid":1,"action":"Db","method":"testEcho1","message":"","result":"Hello!"}`)
					So(string(s), ShouldContainSubstring, `{"type":"rpc","tid":2,"action":"Db","method":"testEcho2","message":"","result":"Hello12340"}`)
				})
			})
		})
	})

	Convey("Exception methods call", t, func() {
		provider := NewProvider()
		provider.LogMode(providerDebug)
		provider.RegisterAction(reflect.TypeOf(Db{}))
		reqs := mustDecodeTransaction(strings.NewReader(`[{"action":"Db","method":"testException1","data":null,"type":"rpc","tid":1},{"action":"Db","method":"testException2","data":null,"type":"rpc","tid":2},{"action":"Db","method":"testException3","data":null,"type":"rpc","tid":3},{"action":"Db","method":"testException4","data":null,"type":"rpc","tid":4}]`))
		Convey("processed with 4 responses", func() {
			resps := provider.processRequests(nil, nil, reqs)
			So(len(resps), ShouldEqual, 4)
			Convey("with null results each", func() {
				for _, resp := range resps {
					So(resp.Result, ShouldBeNil)
				}
			})
			Convey("with exception type each", func() {
				for _, resp := range resps {
					So(resp.Type, ShouldEqual, "exception")
				}
			})
			Convey("with correct message each", func() {
				for _, resp := range resps {
					So(resp.Message, ShouldContainSubstring, "Error example #")
				}
			})
		})
	})

	Convey("Get records", t, func() {
		provider := NewProvider()
		provider.LogMode(providerDebug)
		provider.RegisterAction(reflect.TypeOf(Db{}))
		reqs := mustDecodeTransaction(strings.NewReader(`{"action":"Db","method":"getRecords","data":[{"page":1,"start":0,"limit":25,"sort":[{"property":"text","direction":"ASC"}]}],"type":"rpc","tid":1}`))
		Convey("processed with correct result", func() {
			resps := provider.processRequests(nil, nil, reqs)
			So(len(resps), ShouldEqual, 1)
			So(resps[0].Message, ShouldBeEmpty)
			So(resps[0].Type, ShouldEqual, "rpc")
			So(resps[0].Result, ShouldEqual, `model= page=1 start=0 limit=25 sort=[{text ASC}]`)
		})
	})

	Convey("Context setting", t, func() {
		provider := NewProvider()
		provider.LogMode(providerDebug)
		provider.RegisterAction(reflect.TypeOf(Db{}))
		reqs := mustDecodeTransaction(strings.NewReader(`{"action":"Db","method":"test","data":null,"type":"rpc","tid":1}`))
		resps := provider.processRequests(&web.C{URLParams:map[string]string{"test":"test1"}}, &http.Request{Host: "test2"}, reqs)
		So(len(resps), ShouldEqual, 1)
		So(resps[0].Message, ShouldBeEmpty)
		So(resps[0].Type, ShouldEqual, "rpc")
		So(resps[0].Result, ShouldEqual, "test1test2")
	})
}