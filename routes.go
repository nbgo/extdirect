package extdirect

import (
	"net/http"
	"fmt"
	"strings"
	"io"
	"io/ioutil"
	"encoding/json"
	"reflect"
	"time"
	"golang.org/x/net/context"
	"net/url"
	"strconv"
	"github.com/nbgo/fail"
)

// ErrDecodeFromPostRequest has information about decoding error.
type ErrDecodeFromPostRequest struct {
	Details string
	Reason  error
}

var _ fail.CompositeError = ErrDecodeFromPostRequest{}

func (err ErrDecodeFromPostRequest) Error() string {
	return fmt.Sprintf("failed to decode form post: %v: %v", err.Details, err.Reason)
}
// InnerError implements CompositeError.InnerError().
func (err ErrDecodeFromPostRequest) InnerError() error {
	return err.Reason
}

// ErrInvalidContentType occurs when client request contains invalid content type.
type ErrInvalidContentType string

func (err ErrInvalidContentType) Error() string {
	return fmt.Sprintf("invalid content type: %s", string(err))
}

// ErrTypeConversion contains information about type conversion error.
type ErrTypeConversion struct {
	SourceType reflect.Type
	TargetType reflect.Type
	Reason     error
}

func (err ErrTypeConversion) Error() string {
	reason := ""
	if err.Reason != nil {
		reason = fmt.Sprintf(": %v", err.Reason.Error())
	}
	return fmt.Sprintf("cannot convert type %v to type %v%v", err.SourceType, err.TargetType, reason)
}

// ErrDirectActionMethod contains information about error occurred during direct method execution,
type ErrDirectActionMethod struct {
	Action  string
	Method  string
	Err     interface{}
	isPanic bool
}

func (err ErrDirectActionMethod) Error() string {
	return fmt.Sprintf("error executing %v.%v(): %v", err.Action, err.Method, err.Err)
}

type request struct {
	Type     string            `json:"type"`
	Tid      int               `json:"tid"`
	Action   string            `json:"action"`
	Method   string            `json:"method"`
	Upload   bool              `json:"upload"`
	Data     json.RawMessage   `json:"data"`
	FormData map[string]string `json:"-"`
}

type response struct {
	Type    string      `json:"type"`
	Tid     int         `json:"tid"`
	Action  string      `json:"action"`
	Method  string      `json:"method"`
	Message *string     `json:"message,omitempty"`
	Result  interface{} `json:"result,omitempty"`
}

// API is routes for getting Ext.Direct API script.
func API(provider *DirectServiceProvider) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
		if js, err := provider.JavaScript(); err != nil {
			panic(err)
		} else {
			if _, err := w.Write([]byte(js)); err != nil {
				panic(err)
			}
		}
	}
}

// ActionsHandler is route for handling Ext.Direct requests.
func ActionsHandler(provider *DirectServiceProvider) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		actionHandler(provider, nil, w, r)
	}
}

// ActionsHandlerCtx is route with context suppor for handling Ext.Direct requests.
func ActionsHandlerCtx(provider *DirectServiceProvider) func(c context.Context, w http.ResponseWriter, r *http.Request) {
	return func(c context.Context, w http.ResponseWriter, r *http.Request) {
		actionHandler(provider, c, w, r)
	}
}

func actionHandler(provider *DirectServiceProvider, c context.Context, w http.ResponseWriter, r *http.Request) {
	var reqs []*request
	var err error
	contentType := r.Header.Get("Content-Type")
	isFormHandler := false

	switch {
	case strings.HasPrefix(contentType, "application/json"):
		reqs = mustDecodeTransaction(r.Body)
	case strings.HasPrefix(contentType, "application/x-www-form-urlencoded"):
		r.ParseForm()
		reqs = mustDecodeFormPost(r.Form)
		isFormHandler = true
	default:
		panic(ErrInvalidContentType(contentType))
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if !isFormHandler {
		err = json.NewEncoder(w).Encode(provider.processRequests(c, r, reqs))
	} else {
		resps := provider.processRequests(c, r, reqs)
		err = json.NewEncoder(w).Encode(resps[0])
	}
	if err != nil {
		panic(err)
	}
}

func (provider *DirectServiceProvider) processRequests(c context.Context, r *http.Request, reqs []*request) []*response {
	resps := make([]*response, len(reqs))
	respsChannel := make(chan *response, len(reqs))
	for _, req := range reqs {
		go func(req *request) {
			resp := &response{
				Tid: req.Tid,
				Action: req.Action,
				Method: req.Method,
				Type: req.Type,
			}
			var tStart time.Time
			profilingStarted := false

			logProfiling := func() {
				if profilingStarted {
					duration := time.Now().Sub(tStart)
					log.Print(logLevelInfo, fmt.Sprintf("%s.%s() %v ", req.Action, req.Method, duration), map[string]interface{}{"duration":duration, "action": req.Action, "method": req.Method})
					profilingStarted = false
				}
			}

			defer func() {
				logProfiling()
				if err := recover(); err != nil {
					log.Print(fail.New(ErrDirectActionMethod{req.Action, req.Method, err, true}))
					resp.Type = "exception"
					respMessage := fmt.Sprintf("%v", err)
					resp.Message = &respMessage
				}
				respsChannel <- resp
			}()

			// Create instance of action type
			actionInfo := provider.actionsInfo[req.Action]
			if provider.debug {
				log.Print(fmt.Sprintf("Create instance of action %s (type %v)", req.Action, actionInfo.Type))
			}
			actionVal := reflect.New(actionInfo.Type).Elem()

			// Set context and request
			if c != nil || r != nil {
				if provider.debug {
					log.Print("Set action context/request.")
				}
				contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
				requestType := reflect.TypeOf(&http.Request{})
				fieldsLen := actionInfo.Type.NumField()
				for i := 0; i < fieldsLen; i++ {
					t := actionInfo.Type.Field(i).Type

					if t.Implements(contextType) {
						if c != nil {
							if provider.debug {
								log.Print("Set action context.")
							}
							actionVal.Field(i).Set(reflect.ValueOf(c))
						} else {
							if provider.debug {
								log.Print(logLevelWarn, "Context cannot be set to action instance because context is nil.")
							}
						}
					}

					if t == requestType {
						if r != nil {
							if provider.debug {
								log.Print("Set action request.")
							}
							actionVal.Field(i).Set(reflect.ValueOf(r))
						}
					}
				}
			}

			if provider.debug {
				log.Print(fmt.Sprintf("Prepare arguments for method %s.%s", req.Action, req.Method))
			}
			methodInfo := actionInfo.Methods[req.Method]
			directMethod := actionInfo.DirectMethods[req.Method]
			isFormHandler := false
			if directMethod.FormHandler != nil {
				isFormHandler = *directMethod.FormHandler
			}
			if provider.debug {
				log.Print(fmt.Sprintf("Direct method to use: %s, formhandler=%v", directMethod.Name, isFormHandler))
			}
			methodArgsLen := methodInfo.Type.NumIn() - 1
			var args []reflect.Value
			if (req.Data != nil && !isFormHandler) || (req.FormData != nil && isFormHandler) {
				if isFormHandler {
					if provider.debug {
						log.Print("Prepare arguments for form handler call.")
					}
					args = make([]reflect.Value, 1)
					args[0] = reflect.ValueOf(req.FormData)
					// TODO: Support structure type argument for form handler.
				} else {
					args = make([]reflect.Value, methodArgsLen)
					var argsArray []json.RawMessage
					if err := json.Unmarshal(req.Data, &argsArray); err != nil {
						panic(fail.NewErrWithReason("could not parse request data", err))
					}
					for i, arg := range argsArray {
						methodArgType := methodInfo.Type.In(i + 1)
						if provider.debug {
							log.Print(fmt.Sprintf("Parse `%v` into %v", string(arg), methodArgType))
						}
						argValue := reflect.New(methodArgType).Elem()
						argRef := argValue.Addr().Interface()
						json.Unmarshal(arg, argRef)
						args[i] = reflect.ValueOf(argValue.Interface())
					}
				}
			}

			if provider.profile {
				profilingStarted = true
				tStart = time.Now()
			}

			if provider.debug {
				log.Print(fmt.Sprintf("Call method %s.%s", req.Action, req.Method))
			}
			// Call action method.
			resultsValues := actionVal.MethodByName(methodInfo.Name).Call(args)

			logProfiling()
			for i, resultValue := range resultsValues {
				if methodInfo.Type.Out(i).Name() == "error" {
					if err, isErr := resultValue.Interface().(error); isErr {
						log.Print(&ErrDirectActionMethod{req.Action, req.Method, err, false})
						resp.Type = "exception"
						respMessage := fmt.Sprintf("%v", err)
						resp.Message = &respMessage
						resp.Result = nil
						break;
					}
				} else {
					result := resultValue.Interface()
					resp.Result = result
				}
			}
		}(req)
	}

	for i := 0; i < len(reqs); i++ {
		var resp = <-respsChannel
		resps[i] = resp
	}

	return resps
}

func mustDecodeFormPost(f url.Values) []*request {
	req := &request{
		Type:   f["extType"][0],
		Action: f["extAction"][0],
		Method: f["extMethod"][0],
	}
	tid, tidErr := strconv.Atoi(f["extTID"][0]);
	if tidErr != nil {
		panic(fail.New(ErrDecodeFromPostRequest{"could not parse TID", tidErr}))
	}
	req.Tid = tid
	upload, hasUpload := f["extUpload"]
	req.Upload = hasUpload && strings.ToLower(upload[0]) == "true"

	data := make(map[string]string, 0)
	for k, v := range f {
		if k == "extType" || k == "extTID" || k == "extAction" || k == "extMethod" || k == "extUpload" {
			continue
		}
		data[k] = v[0]
	}
	req.FormData = data

	return []*request{req}
}

func mustDecodeTransaction(r io.Reader) []*request {
	if jsonData, err := ioutil.ReadAll(r); err != nil {
		panic(err)
	} else {
		var reqs []*request
		if err := json.Unmarshal(jsonData, &reqs); err != nil {
			// Attempt to unmarshal as a single request.
			var req request
			if err := json.Unmarshal(jsonData, &req); err != nil {
				panic(err)
			} else {
				reqs = make([]*request, 1)
				reqs[0] = &req
			}
		}
		return reqs
	}
}