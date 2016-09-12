package util

import (
	"reflect"
	"time"
	"strings"
	"strconv"
	"github.com/pkg/errors"
)


type DefaultParser struct { }

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
