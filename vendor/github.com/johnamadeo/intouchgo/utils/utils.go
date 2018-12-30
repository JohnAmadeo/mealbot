package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type Message struct {
	Message string
}

func MessageToBytes(message string) []byte {
	bytes, _ := json.Marshal(Message{message})
	return bytes
}

func PrintErr(err error) {
	fmt.Println(err.Error())
}

// Use JSON tag information to create a form values map.
func JSONToForm(v interface{}) map[string]string {
	value := reflect.ValueOf(v)
	t := value.Type()
	params := make(map[string]string)
	for i := 0; i < value.NumField(); i++ {
		f := t.Field(i)
		name := f.Tag.Get("json")
		fv := value.Field(i).Interface()
		if fv == nil {
			continue
		}
		switch x := fv.(type) {
		case *string:
			if x != nil {
				params[name] = *x
			}
		case string:
			if x != "" {
				params[name] = x
			}
		case int:
			if x != 0 {
				params[name] = strconv.Itoa(x)
			}
		case *bool:
			if x != nil {
				params[name] = fmt.Sprintf("%v", *x)
			}
		case bool:
			params[name] = fmt.Sprintf("%v", x)
		case int64:
			if x != 0 {
				params[name] = strconv.FormatInt(x, 10)
			}
		case float64:
			params[name] = fmt.Sprintf("%.2f", x)
		case []string:
			if len(x) > 0 {
				params[name] = strings.Join(x, " ")
			}
		case map[string]string:
			for mapkey, mapvalue := range x {
				params[name+"["+mapkey+"]"] = mapvalue
			}
		default:
			// ignore
			panic(fmt.Errorf("Unknown field type: " + value.Field(i).Type().String()))
		}
	}
	return params
}
