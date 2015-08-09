package extdirect
import (
	"reflect"
	"encoding/json"
	"bytes"
	"strings"
	"fmt"
)

type DirectServiceProvider struct {
	Type        string `json:"type"`
	Url         string `json:"url"`
	Namespace   string `json:"namespace"`
	Timeout     int `json:"timeout"`
	Actions     map[string]DirectAction `json:"actions"`
	actionsInfo map[string]directActionInfo
	debug       bool
	profile     bool
}

type DirectAction []interface{}

type DirectMethod struct {
	Name string `json:"name"`
	Len  int `json:"len"`
}

type directActionInfo struct {
	Type    reflect.Type
	Methods map[string]reflect.Method
}

func (this *DirectServiceProvider) Json() (string, error) {
	if jsonText, err := json.Marshal(this); err != nil {
		return "", err
	} else {
		return string(jsonText), nil
	}
}

func (this *DirectServiceProvider) Debug(debug bool) {
	this.debug = debug
}

func (this *DirectServiceProvider) Profile(profile bool) {
	this.profile = profile
}

func (this *DirectServiceProvider) JavaScript() (string, error) {
	if apiJson, err := this.Json(); err != nil {
		return "", err
	} else {
		return fmt.Sprintf("Ext.ns(\"%s\");%s.REMOTE_API=%s", this.Namespace, this.Namespace, apiJson), nil
	}
}

func (this *DirectServiceProvider) RegisterAction(typeInfo reflect.Type) {
	actionTypeName := typeInfo.Name()
	if _, ok := this.Actions[actionTypeName]; ok {
		return
	}

	methodsLen := typeInfo.NumMethod()
	directAction := make([]interface{}, 0)
	methods := make(map[string]reflect.Method, 0)

	for i := 0; i < methodsLen; i++ {
		methodInfo := typeInfo.Method(i)
		argsLen := methodInfo.Type.NumIn() - 1

		var directMethod interface{}
		directMethodName := firstCharToLower(methodInfo.Name)
		directMethod = DirectMethod{
			Name: directMethodName,
			Len: argsLen,
		}

		directAction = append(directAction, directMethod)
		methods[directMethodName] = methodInfo
	}

	this.Actions[actionTypeName] = directAction
	this.actionsInfo[actionTypeName] = directActionInfo{
		Type: typeInfo,
		Methods: methods,
	}
}

var Provider *DirectServiceProvider

func init() {
	Provider = NewProvider()
}

func NewProvider() (provider *DirectServiceProvider) {
	provider = &DirectServiceProvider{
		Type: "remoting",
		Namespace: "DirectApi",
		Url: "/directapi",
		Timeout: 30000,
		Actions: make(map[string]DirectAction),
		actionsInfo: make(map[string]directActionInfo),
	}
	return
}

func firstCharToLower(s string) string {
	if len(s) < 2 {
		return strings.ToLower(s)
	}

	bts := []byte(s)

	lc := bytes.ToLower([]byte{bts[0]})
	rest := bts[1:]

	return string(bytes.Join([][]byte{lc, rest}, nil))
}