// Package fmterr is a set of tools for error.
package fmterr

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/tsingsun/woocoo/web/handler"
	"reflect"
)

var UnknownError = errors.New("unknown error")

// codeMap internal error code map
var codeMap = make(map[uint64]string)

// Error is a wrapper for gin.Error
type Error gin.Error

// Unwrap returns the wrapped error, to allow interoperability with errors.Is(), errors.As() and errors.Unwrap()
func (e *Error) Unwrap() error {
	return e.Err
}

// Code return the error code
func (e *Error) Code() gin.ErrorType {
	return e.Type
}

// MetaData return the error metadata
func (e *Error) self() *Error {
	return e
}

// Error implement Error interface
func (e *Error) Error() string {
	if e.Err == nil {
		return ""
	}
	// 直接强制类型转换
	return e.Err.Error()
}

// JSON creates a properly formatted JSON
func (e *Error) JSON() any {
	jsonData := map[string]any{}
	if e.Meta != nil {
		value := reflect.ValueOf(e.Meta)
		switch value.Kind() {
		case reflect.Struct:
			return e.Meta
		default:
			jsonData["meta"] = e.Meta
		}
	}
	if _, ok := jsonData["error"]; !ok {
		jsonData["error"] = e.Error()
	}
	return jsonData
}

// New create a new error
func New(code uint64, err error) *Error {
	return &Error{
		Type: gin.ErrorType(code),
		Err:  err,
	}
}

// Newf create a new error
func Newf(code uint64, format string, a ...any) *Error {
	return &Error{
		Type: gin.ErrorType(code),
		Err:  fmt.Errorf(format, a...),
	}
}

// Code create a new error by code only
func Code(code uint64) *Error {
	return &Error{
		Err:  UnknownError,
		Type: gin.ErrorType(code),
	}
}

// Codef create a new error by code and args.
// args is a list of key value pairs, key is string, value is any.
// length of kvs equals 1 is meaning the error message.
func Codef(code uint64, kvs ...any) *Error {
	e := Code(code)
	if l := len(kvs); l > 0 {
		if l == 1 {
			e.Err = errors.New(kvs[0].(string))
			return e
		}
		meta := make(map[string]any)
		for i := 0; i < len(kvs); i += 2 {
			if i+1 < len(kvs) {
				key, ok := kvs[i].(string)
				if !ok {
					continue
				}
				meta[key] = kvs[i+1]
			}
		}
		e.Meta = meta
	}
	return e
}

// Codel create a new error by code and args. args set to Meta will use int key.
func Codel(code uint64, a ...any) *Error {
	e := Code(code)
	e.Meta = a
	return e
}

// ParseCodeMap parse error code map from configuration.
// The configuration format is {int key} : {string}, for example:
//
//	1000: "error text"
func ParseCodeMap(cfg *conf.Configuration) error {
	return cfg.Unmarshal(&codeMap)
}

// InitErrorHandler pass to the error handler component
func InitErrorHandler(cfg *conf.Configuration) error {
	if err := ParseCodeMap(cfg); err != nil {
		return err
	}
	handler.SetErrorMap(codeMap, nil)
	return nil
}
