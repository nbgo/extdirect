package extdirect

import (
	"reflect"
	"encoding/json"
	"bytes"
	"strings"
	"fmt"
)

// DirectMethodTags serves to host tags for some direct method.
// Example: UpdateBasicInfoTags DirectMethodTags `formhandler:"true"`
// means tag `formhandler:"true"` targets UpdateBasicInfo direct method.
type DirectMethodTags struct{}

type directServiceProviderType string

const (
// RemotingProvider is remoting provider type.
	RemotingProvider directServiceProviderType = "remoting"

// PollingProvider is polling provider type.
	PollingProvider directServiceProviderType = "polling"
)

// DirectServiceProvider represents Ext Direct service settings.
type DirectServiceProvider struct {
	ID          *string `json:"id,omitempty"`
	Type        directServiceProviderType `json:"type"`
	URL         string `json:"url"`
	Namespace   string `json:"namespace"`
	Timeout     int `json:"timeout"`
	Actions     map[string]directAction `json:"actions"`
	actionsInfo map[string]directActionInfo
	debug       bool
	profile     bool
}

type directAction []directMethod

type directMethod struct {
	Name        string `json:"name"`
	// Method declaration MUST have one of the following mutually exclusive properties that describe the Method’s calling convention:
	Len         *int `json:"len,omitempty"`
	FormHandler *bool `json:"formHander,omitempty"`
}

// DirectFormHandlerResult is a result of form handler execution.
type DirectFormHandlerResult struct {
	Errors  map[string]string `json:"errors,omitempty"`
	Success bool `json:"success"`
}

type directActionInfo struct {
	Type          reflect.Type
	Methods       map[string]reflect.Method
	DirectMethods map[string]directMethod
}

// JSON returns provider as JSON string.
func (provider DirectServiceProvider) JSON() (string, error) {
	jsonText, err := json.Marshal(provider);
	if err != nil {
		return "", err
	}
	return string(jsonText), nil
}

// Debug enables/disables debugging for provider.
func (provider *DirectServiceProvider) Debug(debug bool) {
	provider.debug = debug
}

// Profile enables/disables profiling for provider.
func (provider *DirectServiceProvider) Profile(profile bool) {
	provider.profile = profile
}

// JavaScript returns javascript declaration of the provider.
func (provider DirectServiceProvider) JavaScript() (string, error) {
	apiJSON, err := provider.JSON();
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Ext.ns(\"%s\");%s.REMOTE_API=%s", provider.Namespace, provider.Namespace, apiJSON), nil
}

// RegisterAction registers action.
func (provider *DirectServiceProvider) RegisterAction(typeInfo reflect.Type) {
	actionTypeName := typeInfo.Name()
	debug := provider.debug
	if _, ok := provider.Actions[actionTypeName]; ok {
		return
	}

	if debug {
		log.Print(fmt.Sprintf("Register action %v", actionTypeName))
	}

	methodsLen := typeInfo.NumMethod()
	var directAction []directMethod
	methods := make(map[string]reflect.Method, 0)
	directMethods := make(map[string]directMethod, 0)

	if debug {
		log.Print(fmt.Sprintf("\twith %v method(s)", methodsLen))
	}

	for i := 0; i < methodsLen; i++ {
		methodInfo := typeInfo.Method(i)

		if debug {
			log.Print(fmt.Sprintf("\tregister method %v", methodInfo.Name))
		}

		argsLen := methodInfo.Type.NumIn() - 1
		directMethodName := firstCharToLower(methodInfo.Name)
		directMethod := directMethod{Name: directMethodName}

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

	provider.Actions[actionTypeName] = directAction
	provider.actionsInfo[actionTypeName] = directActionInfo{
		Type: typeInfo,
		Methods: methods,
		DirectMethods: directMethods,
	}
}

// Provider is default provider.
var Provider *DirectServiceProvider

func init() {
	Provider = NewProvider()
}

// NewProvider creates new provider with default configuration.
func NewProvider() (provider *DirectServiceProvider) {
	provider = &DirectServiceProvider{
		Type: RemotingProvider,
		Namespace: "DirectApi",
		URL: "/directapi",
		Timeout: 30000,
		Actions: make(map[string]directAction),
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