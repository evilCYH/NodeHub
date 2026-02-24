package register

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"github.com/evilCYH/NodeHub/internal/utils/desc"
)

type registerInfo struct {
	im  map[string]any
	aim map[string][]desc.Data
}

var registers = map[string]*registerInfo{}

func register(t string, i any) {
	r, exists := registers[t]
	if !exists {
		r = &registerInfo{
			im:  make(map[string]any),
			aim: make(map[string][]desc.Data),
		}
		registers[t] = r
	}
	m := strings.ToLower(reflect.TypeOf(i).Elem().Name())
	r.im[m] = i
	r.aim[m] = desc.Gen(i)
}

func Get[T any](t string, m string, c string) (T, error) {
	ri, exists := registers[t]
	if !exists {
		return *new(T), errors.New("category not found")
	}

	info, exists := ri.im[m]
	if !exists {
		return *new(T), errors.New("item not found")
	}

	ni := reflect.New(reflect.TypeOf(info).Elem()).Interface()

	if c != "" {
		err := json.Unmarshal([]byte(c), ni)
		if err != nil {
			return *new(T), err
		}
	}

	return ni.(T), nil
}

func GetList(t string) []string {
	ri, exists := registers[t]
	if !exists {
		return nil
	}

	keys := make([]string, 0, len(ri.im))
	for k := range ri.im {
		keys = append(keys, k)
	}
	return keys
}

func GetInfoMap(t string) map[string][]desc.Data {
	ri, exists := registers[t]
	if !exists {
		return nil
	}
	return ri.aim
}
