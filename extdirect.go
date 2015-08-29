package extdirect
import (
	"reflect"
	"encoding/json"
	"bytes"
	"strings"
	"fmt"
)

type DirectMethodTags struct {}

type DirectServiceProviderType string
const (
	RemotingProvider DirectServiceProviderType = "remoting"
	PollingProvider DirectServiceProviderType = "polling"
)

type DirectServiceProvider struct {
	Id          *string `json:"id,omitempty"`
	Type        DirectServiceProviderType `json:"type"`
	Url         string `json:"url"`
	Namespace   string `json:"namespace"`
	Timeout     int `json:"timeout"`
	Actions     map[string]DirectAction `json:"actions"`
	actionsInfo map[string]directActionInfo
	debug       bool
	profile     bool
}

type DirectAction []DirectMethod

type DirectMethod struct {
	Name        string `json:"name"`
	// Method declaration MUST have one of the following mutually exclusive properties that describe the Method’s calling convention:
	Len         *int `json:"len,omitempty"`
	FormHandler *bool `json:"formHander,omitempty"`
}

type DirectFormHandlerResult struct {
	Errors  map[string]string `json:"errors,omitempty"`
	Success bool `json:"success"`
}

type directActionInfo struct {
	Type          reflect.Type
	Methods       map[string]reflect.Method
	DirectMethods map[string]DirectMethod
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
	debug := this.debug
	if _, ok := this.Actions[actionTypeName]; ok {
		return
	}

	if debug {
		log.Print(fmt.Sprintf("Register action %v", actionTypeName))
	}

	methodsLen := typeInfo.NumMethod()
	directAction := make([]DirectMethod, 0)
	methods := make(map[string]reflect.Method, 0)
	directMethods := make(map[string]DirectMethod, 0)

	if debug {
		log.Print(fmt.Sprintf("\twith %v methods", methodsLen))
	}

	for i := 0; i < methodsLen; i++ {
		methodInfo := typeInfo.Method(i)

		if debug {
			log.Print(fmt.Sprintf("\tregister method %v", methodInfo.Name))
		}

		argsLen := methodInfo.Type.NumIn() - 1
		directMethodName := firstCharToLower(methodInfo.Name)
		directMethod := DirectMethod{Name: directMethodName}

		if debug {
			log.Print(fmt.Sprintf("\t\twith args len = %v", argsLen))
			log.Print("\t\tget method tags")
		}

		// Get method tags.
		if tagsField := getDirectMethodTags(typeInfo, methodInfo.Name, debug); tagsField != nil {
			if debug {
				log.Print("\t\t\ttags found")
			}

			if tagsField.Tag.Get("formhandler") == "true" {
				directMethod.FormHandler = new(bool)
				*directMethod.FormHandler = true
			}
		} else {
			if debug {
				log.Print("\t\t\tno tags found")
			}
		}

		if directMethod.FormHandler == nil {
			directMethod.Len = new(int)
			*directMethod.Len = argsLen
		}

		directAction = append(directAction, directMethod)
		methods[directMethodName] = methodInfo
		directMethods[directMethodName] = directMethod
	}

	this.Actions[actionTypeName] = directAction
	this.actionsInfo[actionTypeName] = directActionInfo{
		Type: typeInfo,
		Methods: methods,
		DirectMethods: directMethods,
	}
}

var Provider *DirectServiceProvider

func init() {
	Provider = NewProvider()
}

func NewProvider() (provider *DirectServiceProvider) {
	provider = &DirectServiceProvider{
		Type: RemotingProvider,
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

func getDirectMethodTags(t reflect.Type, methodName string, debug bool) *reflect.StructField {
	fieldsLen := t.NumField()
	dmt := reflect.TypeOf(DirectMethodTags{})

	if debug {
		log.Print(fmt.Sprintf("\t\t\tsearch tag among %v fields", fieldsLen))
	}

	for i := 0; i < fieldsLen; i++ {
		f := t.Field(i)
		if debug {
			log.Print(fmt.Sprintf("\t\t\t\tfield %v of type %v", f.Name, f.Type))
		}
		if f.Name == (methodName + "Tags") && f.Type == dmt {
			if debug {
				log.Print("\t\t\t\t\tis a tag")
			}

			return &f;
		}
		if debug {
			log.Print(fmt.Sprintf("\t\t\t\t\tis NOT a tag: nameOk=%v, typeOk=%v", f.Name == (methodName + "Tags"), f.Type == dmt))
		}
	}

	return nil
}