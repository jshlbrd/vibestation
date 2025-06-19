package message

import (
	"encoding/json"
	"reflect"
)

func jsonEqual(a, b string) bool {
	var o1, o2 interface{}
	if err := json.Unmarshal([]byte(a), &o1); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(b), &o2); err != nil {
		return false
	}
	return reflect.DeepEqual(o1, o2)
}
