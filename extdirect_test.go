package extdirect

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	. "github.com/ahmetalpbalkan/go-linq"
	nblogger "github.com/nbgo/logger"
	"reflect"
	"net/http"
	"strings"
	"encoding/json"
	"fmt"
	"time"
	"github.com/zenazn/goji/web"
	"github.com/Sirupsen/logrus"
	"errors"
	gcontext "github.com/goji/context"
	"net/http/httptest"
	"io/ioutil"
	. "github.com/jacobsa/oglematchers"
	"golang.org/x/net/context"
	"github.com/nbgo/fail"
	"github.com/nbgo/jsontime"
)

var providerDebug = true
var providerProfile = true

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

type RequestWithTime struct {
	Timestamp *jsontime.RFC3339Nano `json:"timestamp"`
}

type Db struct {
	C                   context.Context
	R                   *http.Request
	UpdateBasicInfoTags DirectMethodTags `formhandler:"true"`
}

func (this Db) GetRecords(r *GetDataRequest) (string, error) {
	result := fmt.Sprintf("model=%v page=%v start=%v limit=%v sort=%v", r.Model, r.Page, r.Start, r.Limit, r.Sort)
	return result, nil
}
func (this Db) Test() string {
	time.Sleep(30 * time.Millisecond)
	var result string
	if this.C != nil {
		gc := gcontext.ToC(this.C)
		if v, ok := gc.URLParams["test"]; ok {
			result += v
		}
		if v, ok := this.C.Value("user").(string); ok {
			result += v
		}
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
	return fail.News("Error example #2")
}
func (this Db) TestException3() (string, error) {
	return "test", fail.News("Error example #3")
}
func (this Db) TestException4() {
	panic(errors.New("Error example #4"))
}
func (this Db) UpdateBasicInfo(data map[string]string) (result *DirectFormHandlerResult) {
	result = &DirectFormHandlerResult{Success:true}
	if data["email"] == "aaron@sencha.com" {
		result.Success = false
		result.Errors = make(map[string]string, 0)
		result.Errors["email"] = "already exists"
	}
	return
}
func (this Db) TestTime(r *RequestWithTime) string {
	return time.Time(*r.Timestamp).Format(time.RFC3339Nano)
}

func getResponseByTid(responses []*response, tid int) *response {
	resp, _, _ := From(responses).FirstBy(func(x T) (bool, error) {
		return x.(*response).Tid == tid, nil
	})

	return resp.(*response)
}

func TestExtDirect(t *testing.T) {
	SetLogger(&LogrusLogger{nblogger.Create()})
	logrus.SetLevel(logrus.DebugLevel)

	Convey("Default provider serialization", t, func() {
		value, err := Provider.JSON()
		So(err, ShouldBeNil)
		So(value, ShouldEqual, "{\"type\":\"remoting\",\"url\":\"/directapi\",\"namespace\":\"DirectApi\",\"timeout\":30000,\"actions\":{}}")
	})

	Convey("Action registration", t, func() {
		provider := NewProvider()
		provider.Debug(providerDebug)
		provider.Profile(providerProfile)
		provider.RegisterAction(reflect.TypeOf(Db{}))

		Convey("One registered action with name 'Db' expected", func() {
			So(len(provider.Actions), ShouldEqual, 1)
			So(provider.Actions, ShouldContainKey, "Db")
			Convey("and has 9 methods", func() {
				So(len(provider.Actions["Db"]), ShouldEqual, 10)
				Convey("test", func() {
					method, exists, err := From(provider.Actions["Db"]).FirstBy(func(x T) (bool, error) {
						if m, ok := x.(directMethod); ok {
							return m.Name == "test", nil
						} else {
							return false, nil
						}
					})
					So(err, ShouldBeNil)
					So(exists, ShouldBeTrue)
					So(method.(directMethod).Name, ShouldEqual, "test")
					Convey("with no arguments", func() {
						So(*method.(directMethod).Len, ShouldBeZeroValue)
					});
				})
				Convey("getRecords", func() {
					method, exists, err := From(provider.Actions["Db"]).FirstBy(func(x T) (bool, error) {
						if m, ok := x.(directMethod); ok {
							return m.Name == "getRecords", nil
						} else {
							return false, nil
						}
					})
					So(err, ShouldBeNil)
					So(exists, ShouldBeTrue)
					So(method.(directMethod).Name, ShouldEqual, "getRecords")
					Convey("with 1 argument", func() {
						So(*method.(directMethod).Len, ShouldEqual, 1)
					});
				})
				Convey("testEcho1", func() {
					method, exists, err := From(provider.Actions["Db"]).FirstBy(func(x T) (bool, error) {
						if m, ok := x.(directMethod); ok {
							return m.Name == "testEcho1", nil
						} else {
							return false, nil
						}
					})
					So(err, ShouldBeNil)
					So(exists, ShouldBeTrue)
					So(method.(directMethod).Name, ShouldEqual, "testEcho1")
					Convey("with 1 argument", func() {
						So(*method.(directMethod).Len, ShouldEqual, 1)
					});
				})
				Convey("testEcho2", func() {
					method, exists, err := From(provider.Actions["Db"]).FirstBy(func(x T) (bool, error) {
						if m, ok := x.(directMethod); ok {
							return m.Name == "testEcho2", nil
						} else {
							return false, nil
						}
					})
					So(err, ShouldBeNil)
					So(exists, ShouldBeTrue)
					So(method.(directMethod).Name, ShouldEqual, "testEcho2")
					Convey("with 7 arguments", func() {
						So(*method.(directMethod).Len, ShouldEqual, 7)
					});
				})
				Convey("updateBasicInfo", func() {
					method, exists, err := From(provider.Actions["Db"]).FirstBy(func(x T) (bool, error) {
						if m, ok := x.(directMethod); ok {
							return m.Name == "updateBasicInfo", nil
						} else {
							return false, nil
						}
					})
					So(err, ShouldBeNil)
					So(exists, ShouldBeTrue)
					So(method.(directMethod).Name, ShouldEqual, "updateBasicInfo")
					Convey("marked as form handler", func() {
						So(method.(directMethod).FormHandler, ShouldNotBeNil)
						So(*method.(directMethod).FormHandler, ShouldBeTrue)
					});
					Convey("has no len property", func() {
						So(method.(directMethod).Len, ShouldBeNil)
					});
				})
			})
		})

		Convey("Action with methods serialization", func() {
			jsonText, err := provider.JSON()
			So(err, ShouldBeNil)
			So(jsonText, ShouldEqual, `{"type":"remoting","url":"/directapi","namespace":"DirectApi","timeout":30000,"actions":{"Db":[{"name":"getRecords","len":1},{"name":"test","len":0},{"name":"testEcho1","len":1},{"name":"testEcho2","len":7},{"name":"testException1","len":0},{"name":"testException2","len":0},{"name":"testException3","len":0},{"name":"testException4","len":0},{"name":"testTime","len":1},{"name":"updateBasicInfo","formHander":true}]}}`)
			javaScript, err2 := provider.JavaScript()
			So(err2, ShouldBeNil)
			So(javaScript, ShouldEqual, `Ext.ns("DirectApi");DirectApi.REMOTE_API={"type":"remoting","url":"/directapi","namespace":"DirectApi","timeout":30000,"actions":{"Db":[{"name":"getRecords","len":1},{"name":"test","len":0},{"name":"testEcho1","len":1},{"name":"testEcho2","len":7},{"name":"testException1","len":0},{"name":"testException2","len":0},{"name":"testException3","len":0},{"name":"testException4","len":0},{"name":"testTime","len":1},{"name":"updateBasicInfo","formHander":true}]}}`)
		})

		Convey("Duplicated registration", func() {
			provider.RegisterAction(reflect.TypeOf(Db{}))
			So(len(provider.Actions), ShouldEqual, 1)
		})
	})

	Convey("Request with single action call", t, func() {
		provider := NewProvider()
		provider.Debug(providerDebug)
		provider.Profile(providerProfile)
		provider.RegisterAction(reflect.TypeOf(Db{}))
		reqs := mustDecodeTransaction(strings.NewReader(`{"action":"Db","method":"test","data":null,"type":"rpc","tid":1}`))
		Convey("has one parsed request with correct fields", func() {
			So(len(reqs), ShouldEqual, 1)
			So(reqs[0].Action, ShouldEqual, "Db")
			So(reqs[0].Method, ShouldEqual, "test")
			So(string(reqs[0].Data), ShouldEqual, "null")
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
				So(resps[0].Message, ShouldBeNil)
				Convey("which is serialized correctly", func() {
					s, err := json.Marshal(resps)
					So(err, ShouldBeNil)
					// Response is always array even for single request
					So(string(s), ShouldEqual, `[{"type":"rpc","tid":1,"action":"Db","method":"test","result":""}]`)
				})
			})
		})
	})

	Convey("Request with multiple actions call", t, func() {
		provider := NewProvider()
		provider.Debug(providerDebug)
		provider.Profile(providerProfile)
		provider.RegisterAction(reflect.TypeOf(Db{}))
		reqs := mustDecodeTransaction(strings.NewReader(`[{"action":"Db","method":"testEcho1","data":["Hello!"],"type":"rpc","tid":1},{"action":"Db","method":"testEcho2","data":["Hello", 1, 2, 3, 4, null, null],"type":"rpc","tid":2}]`))
		Convey("has 2 parsed requests with correct fields", func() {
			So(len(reqs), ShouldEqual, 2)
			So(reqs[0].Action, ShouldEqual, "Db")
			So(reqs[0].Method, ShouldEqual, "testEcho1")
			So(string(reqs[0].Data), ShouldEqual, `["Hello!"]`)
			So(reqs[0].Tid, ShouldEqual, 1)
			So(reqs[0].Type, ShouldEqual, "rpc")

			So(reqs[1].Action, ShouldEqual, "Db")
			So(reqs[1].Method, ShouldEqual, "testEcho2")
			So(string(reqs[1].Data), ShouldEqual, `["Hello", 1, 2, 3, 4, null, null]`)
			So(reqs[1].Tid, ShouldEqual, 2)
			So(reqs[1].Type, ShouldEqual, "rpc")

			Convey("which is concurrently processed into 2 responses with correct fields", func() {
				t1 := time.Now()
				resps := provider.processRequests(nil, nil, reqs)
				t2 := time.Now()
				So(t2.Sub(t1), ShouldBeLessThan, 50 * time.Millisecond)
				So(len(resps), ShouldEqual, 2)

				testEcho1Resp := getResponseByTid(resps, 1);
				So(testEcho1Resp.Message, ShouldBeNil)
				So(testEcho1Resp.Type, ShouldEqual, "rpc")
				So(testEcho1Resp.Action, ShouldEqual, "Db")
				So(testEcho1Resp.Method, ShouldEqual, "testEcho1")
				So(testEcho1Resp.Result, ShouldEqual, "Hello!")
				So(testEcho1Resp.Tid, ShouldEqual, 1)

				testEcho2Resp := getResponseByTid(resps, 2);
				So(testEcho2Resp.Message, ShouldBeNil)
				So(testEcho2Resp.Type, ShouldEqual, "rpc")
				So(testEcho2Resp.Action, ShouldEqual, "Db")
				So(testEcho2Resp.Method, ShouldEqual, "testEcho2")
				So(testEcho2Resp.Result, ShouldEqual, "Hello12340")
				So(testEcho2Resp.Tid, ShouldEqual, 2)
				Convey("which is serialized correctly", func() {
					s, err := json.Marshal(resps)
					So(err, ShouldBeNil)
					So(string(s), ShouldContainSubstring, `{"type":"rpc","tid":1,"action":"Db","method":"testEcho1","result":"Hello!"}`)
					So(string(s), ShouldContainSubstring, `{"type":"rpc","tid":2,"action":"Db","method":"testEcho2","result":"Hello12340"}`)
				})
			})
		})
	})

	Convey("Exception methods call", t, func() {
		provider := NewProvider()
		provider.Debug(providerDebug)
		provider.Profile(providerProfile)
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
					So(*resp.Message, ShouldContainSubstring, "Error example #")
				}
			})
		})
	})

	Convey("Get records", t, func() {
		provider := NewProvider()
		provider.Debug(providerDebug)
		provider.Profile(providerProfile)
		provider.RegisterAction(reflect.TypeOf(Db{}))
		reqs := mustDecodeTransaction(strings.NewReader(`{"action":"Db","method":"getRecords","data":[{"page":1,"start":0,"limit":25,"sort":[{"property":"text","direction":"ASC"}]}],"type":"rpc","tid":1}`))
		Convey("processed with correct result", func() {
			resps := provider.processRequests(nil, nil, reqs)
			So(len(resps), ShouldEqual, 1)
			So(resps[0].Message, ShouldBeNil)
			So(resps[0].Type, ShouldEqual, "rpc")
			So(resps[0].Result, ShouldEqual, `model= page=1 start=0 limit=25 sort=[{text ASC}]`)
		})
	})

	Convey("Request with time", t, func() {
		provider := NewProvider()
		provider.Debug(providerDebug)
		provider.Profile(providerProfile)
		provider.RegisterAction(reflect.TypeOf(Db{}))
		reqs := mustDecodeTransaction(strings.NewReader(`{"action":"Db","method":"testTime","data":[{"timestamp":"2009-11-10T23:00:00Z"}],"type":"rpc","tid":1}`))
		Convey("processed with correct result", func() {
			resps := provider.processRequests(nil, nil, reqs)
			So(len(resps), ShouldEqual, 1)
			if resps[0].Message != nil {
				So(*resps[0].Message, ShouldBeEmpty)
			}
			So(resps[0].Message, ShouldBeNil)
			So(resps[0].Type, ShouldEqual, "rpc")
			So(resps[0].Result, ShouldEqual, `2009-11-10T23:00:00Z`)
		})
	})

	Convey("Context setting", t, func() {
		provider := NewProvider()
		provider.Debug(providerDebug)
		provider.Profile(providerProfile)
		provider.RegisterAction(reflect.TypeOf(Db{}))
		reqs := mustDecodeTransaction(strings.NewReader(`{"action":"Db","method":"test","data":null,"type":"rpc","tid":1}`))
		resps := provider.processRequests(gcontext.Set(&web.C{
			URLParams:map[string]string{"test":"test1"},
			Env: map[interface{}]interface{}{
				"user": "TestUser",
			},
		}, context.Background()), &http.Request{Host: "test2"}, reqs)
		So(len(resps), ShouldEqual, 1)
		So(resps[0].Message, ShouldBeNil)
		So(resps[0].Type, ShouldEqual, "rpc")
		So(resps[0].Result, ShouldEqual, "test1TestUsertest2")
	})

	Convey("HTTP handlers", t, func() {
		provider := NewProvider()
		provider.Debug(providerDebug)
		provider.Profile(providerProfile)
		provider.RegisterAction(reflect.TypeOf(Db{}))

		mux := web.New()
		mux.Use(gcontext.Middleware)
		mux.Use(func(c *web.C, h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.Env["user"] = "TestUser"
				h.ServeHTTP(w, r)
			})
		})
		mux.Get(provider.URL, API(provider))
		mux.Post(provider.URL, func(c web.C, w http.ResponseWriter, r *http.Request) {
			ActionsHandlerCtx(provider)(gcontext.FromC(c), w, r)
		})
		mux.Post("/directapi2", ActionsHandler(provider))
		srv := httptest.NewServer(mux)
		defer srv.Close()

		Convey("API info request", func() {
			res, err := http.Get(srv.URL + provider.URL)
			Convey("should be processed without error", func() {
				So(err, ShouldBeNil)
				Convey("have correct content type", func() {
					So(res.Header.Get("Content-Type"), ShouldEqual, "text/javascript; charset=utf-8")
					Convey("and return expected body", func() {
						body, err := ioutil.ReadAll(res.Body)
						res.Body.Close()
						So(err, ShouldBeNil)
						So(string(body), ShouldEqual, `Ext.ns("DirectApi");DirectApi.REMOTE_API={"type":"remoting","url":"/directapi","namespace":"DirectApi","timeout":30000,"actions":{"Db":[{"name":"getRecords","len":1},{"name":"test","len":0},{"name":"testEcho1","len":1},{"name":"testEcho2","len":7},{"name":"testException1","len":0},{"name":"testException2","len":0},{"name":"testException3","len":0},{"name":"testException4","len":0},{"name":"testTime","len":1},{"name":"updateBasicInfo","formHander":true}]}}`)
					})
				})
			})
		})

		Convey("API handler request with context", func() {
			res, err := http.Post(srv.URL + provider.URL, "application/json", strings.NewReader(`{"action":"Db","method":"test","data":null,"type":"rpc","tid":33}`))
			Convey("should be processed without error", func() {
				So(err, ShouldBeNil)
				Convey("have correct content type", func() {
					So(res.Header.Get("Content-Type"), ShouldEqual, "application/json; charset=utf-8")
					Convey("and return expected body", func() {
						body, err := ioutil.ReadAll(res.Body)
						res.Body.Close()
						So(err, ShouldBeNil)
						fmt.Println(string(body))
						So(MatchesRegexp(`\[{"type":"rpc","tid":33,"action":"Db","method":"test","result":"TestUser127\.0\.0\.1:\d+"}]`).Matches(string(body)), ShouldBeNil)
					})
				})
			})
		})

		Convey("API handler request without context", func() {
			res, err := http.Post(srv.URL + "/directapi2", "application/json", strings.NewReader(`{"action":"Db","method":"test","data":null,"type":"rpc","tid":33}`))
			Convey("should be processed without error", func() {
				So(err, ShouldBeNil)
				Convey("have correct content type", func() {
					So(res.Header.Get("Content-Type"), ShouldEqual, "application/json; charset=utf-8")
					Convey("and return expected body", func() {
						body, err := ioutil.ReadAll(res.Body)
						res.Body.Close()
						So(err, ShouldBeNil)
						fmt.Println(string(body))
						So(MatchesRegexp(`\[{"type":"rpc","tid":33,"action":"Db","method":"test","result":"127\.0\.0\.1:\d+"}]`).Matches(string(body)), ShouldBeNil)
					})
				})
			})
		})

		Convey("API handler request to exception method", func() {
			res, err := http.Post(srv.URL + provider.URL, "application/json", strings.NewReader(`{"action":"Db","method":"testException1","data":null,"type":"rpc","tid":40}`))
			Convey("should be processed without error", func() {
				So(err, ShouldBeNil)
				Convey("have correct content type", func() {
					So(res.Header.Get("Content-Type"), ShouldEqual, "application/json; charset=utf-8")
					Convey("and return expected body", func() {
						body, err := ioutil.ReadAll(res.Body)
						res.Body.Close()
						bodyString := strings.TrimSuffix(string(body), "\n")
						So(err, ShouldBeNil)
						fmt.Println(bodyString)
						So(bodyString, ShouldEqual, `[{"type":"exception","tid":40,"action":"Db","method":"testException1","message":"Error example #1"}]`)
					})
				})
			})
		})

		Convey("API handler request to form handler", func() {
			res, err := http.Post(srv.URL + "/directapi", "application/x-www-form-urlencoded; charset=UTF-8", strings.NewReader(`extTID=1&extAction=Db&extMethod=updateBasicInfo&extType=rpc&extUpload=false&foo=bar&uid=34&name=Aaron%20Conran&email=aaron%40sencha1.com&company=Sencha%20Inc.`))
			Convey("should be processed without error", func() {
				So(err, ShouldBeNil)
				Convey("have correct content type", func() {
					So(res.Header.Get("Content-Type"), ShouldEqual, "application/json; charset=utf-8")
					Convey("and return expected body", func() {
						body, err := ioutil.ReadAll(res.Body)
						res.Body.Close()
						bodyString := strings.TrimSuffix(string(body), "\n")
						So(err, ShouldBeNil)
						fmt.Println(bodyString)
						So(bodyString, ShouldEqual, `{"type":"rpc","tid":1,"action":"Db","method":"updateBasicInfo","result":{"success":true}}`)
					})
				})
			})
		})

		Convey("API handler request to form handler returning validation error", func() {
			res, err := http.Post(srv.URL + "/directapi", "application/x-www-form-urlencoded; charset=UTF-8", strings.NewReader(`extTID=1&extAction=Db&extMethod=updateBasicInfo&extType=rpc&extUpload=false&foo=bar&uid=34&name=Aaron%20Conran&email=aaron%40sencha.com&company=Sencha%20Inc.`))
			Convey("should be processed without error", func() {
				So(err, ShouldBeNil)
				Convey("have correct content type", func() {
					So(res.Header.Get("Content-Type"), ShouldEqual, "application/json; charset=utf-8")
					Convey("and return expected body", func() {
						body, err := ioutil.ReadAll(res.Body)
						res.Body.Close()
						bodyString := strings.TrimSuffix(string(body), "\n")
						So(err, ShouldBeNil)
						fmt.Println(bodyString)
						So(bodyString, ShouldEqual, `{"type":"rpc","tid":1,"action":"Db","method":"updateBasicInfo","result":{"errors":{"email":"already exists"},"success":false}}`)
					})
				})
			})
		})
	})
}