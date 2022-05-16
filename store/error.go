package store

import (
	"fmt"
	"reflect"
)

type AlreadyExistsError struct {
	val interface{}
}

func (e *AlreadyExistsError) Error() string {
	return fmt.Sprintf("%s already exists", reflect.TypeOf(e.val).String())
}

func (e *AlreadyExistsError) Value() interface{} {
	return e.val
}

func alreadyExists(v interface{}) *AlreadyExistsError {
	return &AlreadyExistsError{
		val: v,
	}
}
