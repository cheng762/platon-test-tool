package testcases

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

func listCase(m caseTest) []string {
	var names []string
	object := reflect.TypeOf(m)
	for i := 0; i < object.NumMethod(); i++ {
		method := object.Method(i)
		if strings.HasPrefix(method.Name, PrefixCase) {
			names = append(names, strings.TrimPrefix(method.Name, "Case"))
		}
	}
	return names
}

func execCase(caseName string, m caseTest) error {
	methodNames := PrefixCase + caseName
	val := reflect.ValueOf(m).MethodByName(methodNames).Call([]reflect.Value{})
	if val[0].IsNil() {
		return nil
	}
	err := val[0].Interface().(error)
	return err
}

func SendError(caseName string, err error) error {
	log.Printf("[fail]test case %v fail: %v ", caseName, err)
	return fmt.Errorf("test case %v fail: %v ", caseName, err)
}
