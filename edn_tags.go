package edn

import (
	"encoding/base64"
	"errors"
	"math/big"
	"reflect"
	"sync"
	"time"
)

var (
	ErrNotFunc         = errors.New("Value is not a function")
	ErrMismatchArities = errors.New("Function does not have single argument in, two argument out")
	ErrNotConcrete     = errors.New("Value is not a concrete non-function type")
	ErrTagOverwritten  = errors.New("Previous tag implementation was overwritten")
)

var globalTags TagMap

// A TagMap contains mappings from tag literals to functions and structs that is
// used when decoding.
type TagMap struct {
	sync.RWMutex
	m map[string]reflect.Value
}

var errorType = reflect.TypeOf((*error)(nil)).Elem()

// AddTagFn adds fn as a converter function for tagname tags to this TagMap. fn
// must have the signature func(T) (U, error), where T is the expected input
// type and U is the output type.
func (tm *TagMap) AddTagFn(tagname string, fn interface{}) error {
	// TODO: check name
	rfn := reflect.ValueOf(fn)
	rtyp := rfn.Type()
	if rtyp.Kind() != reflect.Func {
		return ErrNotFunc
	}
	if rtyp.NumIn() != 1 || rtyp.NumOut() != 2 || !rtyp.Out(1).Implements(errorType) {
		// ok to have variadic arity?
		return ErrMismatchArities
	}
	return tm.addVal(tagname, rfn)
}

func (tm *TagMap) addVal(name string, val reflect.Value) error {
	tm.Lock()
	if tm.m == nil {
		tm.m = map[string]reflect.Value{}
	}
	_, ok := tm.m[name]
	tm.m[name] = val
	tm.Unlock()
	if ok {
		return ErrTagOverwritten
	} else {
		return nil
	}
}

// AddTagFn adds fn as a converter function for tagname tags to the global
// TagMap. fn must have the signature func(T) (U, error), where T is the
// expected input type and U is the output type.
func AddTagFn(tagname string, fn interface{}) error {
	return globalTags.AddTagFn(tagname, fn)
}

// AddTagStructs adds the struct as a matching struct for tagname tags to this
// TagMap. val can not be a channel, function, interface or an unsafe pointer.
func (tm *TagMap) AddTagStruct(tagname string, val interface{}) error {
	rstruct := reflect.ValueOf(val)
	switch rstruct.Type().Kind() {
	case reflect.Invalid, reflect.Chan, reflect.Func, reflect.Interface, reflect.UnsafePointer:
		return ErrNotConcrete
	}
	return tm.addVal(tagname, rstruct)
}

// AddTagStructs adds the struct as a matching struct for tagname tags to the
// global TagMap. val can not be a channel, function, interface or an unsafe
// pointer.
func AddTagStruct(tagname string, val interface{}) error {
	return globalTags.AddTagStruct(tagname, val)
}

func init() {
	err := AddTagFn("inst", func(s string) (time.Time, error) {
		return time.Parse(time.RFC3339Nano, s)
	})
	if err != nil {
		panic(err)
	}
	err = AddTagFn("base64", base64.StdEncoding.DecodeString)
	if err != nil {
		panic(err)
	}
}

// A MathContext specifies the precision and rounding mode for
// `math/big.Float`s when decoding.
type MathContext struct {
	Precision uint
	Mode      big.RoundingMode
}

// The GlobalMathContext is the global MathContext. It is used if no other
// context is provided.
var GlobalMathContext = MathContext{
	Mode:      big.ToNearestEven,
	Precision: 192,
}
