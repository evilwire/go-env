// Defines an object that parses values from strings.
package goenv

import (
	"reflect"
	"time"
	"strings"
	"strconv"
	"github.com/pkg/errors"
)


// A default way to parse a string into a specific primitive or pointer.
type DefaultParser struct { }

// Parse a string value for a specific type given by reflect.Type.
// For example, ParseType might accept str="2" and reflect.Type=reflect.Uint
// and parses the uint value of 2 returned as reflect.Value.
//
// In this particular case, we parse all numeric types, pointers, strings,
// booleans, arrays and slices. The method handles Durations differently, though
// under the hood, the type is treated the same way as int64. In particular, we
// parse durations of the form `1m3s` and more generally, expects the string to be
// parse-able via ParseDuration.
//
// If the object isn't one of the supported types, it throws an error.
func (marshaler *DefaultParser) ParseType(str string, t reflect.Type) (reflect.Value, error) {
	val := reflect.New(t).Elem()
	tName := t.Name()
	tKind := t.Kind()

	if tName == "Duration" {
		// do duration stuff here
		duration, err := time.ParseDuration(str)
		if err != nil {
			return val, errors.Wrapf(err, "Could not parse duration \"%s\"", str)
		}

		durVal := reflect.ValueOf(duration)
		val.Set(durVal)

		return val, nil
	}

	switch tKind {

	case reflect.Ptr:
		indirectVal, err := marshaler.ParseType(str, t.Elem())
		if err != nil {
			return val, err
		}
		val.Set(indirectVal.Addr())

	case reflect.String:
		val.SetString(strings.TrimSpace(str))

	case reflect.Bool:
		b, err := strconv.ParseBool(strings.ToLower(str))
		if err != nil {
			return val, errors.Wrapf(err, "Cannot convert %s to a boolean value.", str)
		}
		val.SetBool(b)

	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		uintVal, convErr := strconv.ParseUint(str, 10, 64)
		if convErr != nil {
			return val, errors.Wrapf(
				convErr,
				"Cannot convert %s to %s", str, tName)
		}

		if val.OverflowUint(uintVal) {
			return val, errors.Errorf("The value %d overflows type %s", uintVal, tName)
		}

		val.SetUint(uintVal)

	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		intVal, convErr := strconv.ParseInt(str, 10, 64)
		if convErr != nil {
			return val, errors.Wrapf(
				convErr,
				"Cannot convert %s to %s", str, tName)
		}

		if val.OverflowInt(intVal) {
			return val, errors.Errorf("The value %d overflows type %s", intVal, tName)
		}

		val.SetInt(intVal)

	case reflect.Float32, reflect.Float64:
		floatVal, convErr := strconv.ParseFloat(str, 64)
		if convErr != nil {
			return val, errors.Wrapf(
				convErr,
				"Cannot convert %s to %s", str, tName)
		}

		if val.OverflowFloat(floatVal) {
			return val, errors.Errorf("The value %d overflows type %s", floatVal, tName)
		}
		val.SetFloat(floatVal)

	case reflect.Array, reflect.Slice:
		var elts []string

		// it seems that "" makes more sense as a way to express an empty
		// list than an element with nothing in it
		if str == "" {
			elts = []string{}
		} else {
			elts = strings.Split(str, ",")
		}
		arrVal := reflect.MakeSlice(t, len(elts), len(elts))
		eltType := t.Elem()

		for i, elt := range elts {
			trimmedElt := strings.TrimSpace(elt)
			eltVal, marshalErr := marshaler.ParseType(trimmedElt, eltType)
			if marshalErr != nil {
				return val, errors.Wrapf(
					marshalErr,
					"Could not marshal element %d", i)
			}
			arrVal.Index(i).Set(eltVal)
		}
		val.Set(arrVal)

	default:
		return val, errors.Errorf("Cannot unmarshal objects of type %s", tName)
	}

	return val, nil
}

// Unmarshals a string into any one of the string-parseable types, which include
// (pointers of) numeric types, strings, booleans, arrays and slices. The method also
// handles Duration separately.
//
// The method throws an error if the underlying interface is unsettable (see
// https://golang.org/pkg/reflect/#Value.CanSet), or if parsing for the value resulted in
// error
func (marshaler *DefaultParser) Unmarshal(val string, i interface{}) error {
	t := reflect.TypeOf(i)
	v := reflect.ValueOf(i)
	if t.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
		t = v.Type()
	}
	tName := t.Name()

	// if the object is not settable (see reflect pkg)
	// then we should raise an error immediately
	if !v.CanSet() {
		if tName != "" {
			return errors.Errorf("Cannot serialize into an unsettable %s type.", tName)
		}
		return errors.New("Cannot serialize into an unsettable type.")
	}

	unmarshaledVal, err := marshaler.ParseType(val, t)
	if err != nil {
		return err
	}
	v.Set(unmarshaledVal)

	return nil
}
