package ext

/*
#include <stdlib.h>
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"unsafe"
)

type Response struct {
	Ok    bool `json:"ok"`
	Value any  `json:"value,omitempty"`
	Error string `json:"error,omitempty"`
}

func toCString(v any) *C.char {
	bytes, _ := json.Marshal(v)
	return C.CString(string(bytes))
}

func readInput(ptr *C.char) (map[string]any, error) {
	if ptr == nil {
		return nil, fmt.Errorf("null input")
	}
	var data map[string]any
	err := json.Unmarshal([]byte(C.GoString(ptr)), &data)
	return data, err
}

var registry = map[string]func(map[string]any) (any, error){}

func Register(name string, fn func(map[string]any) (any, error)) {
	registry[name] = fn
}

//export titan_invoke
func titan_invoke(input *C.char) *C.char {
	payload, err := readInput(input)
	if err != nil {
		return toCString(Response{Ok: false, Error: err.Error()})
	}

	fnName := payload["fn"].(string)
	data, _ := payload["data"].(map[string]any)

	fn := registry[fnName]
	result, err := fn(data)

	if err != nil {
		return toCString(Response{Ok: false, Error: err.Error()})
	}

	return toCString(Response{Ok: true, Value: result})
}

//export titan_free
func titan_free(ptr *C.char) {
	if ptr != nil {
		C.free(unsafe.Pointer(ptr))
	}
}

func main() {}