package util

import (
	"os"
	"reflect"
	"github.com/pkg/errors"
)

type EnvReader interface {

	// look up the value for a particular env variable
	// returning false if the variable is not registered
	LookupEnv(string) (string, bool)

	// returns whether or not env variables are set in
	// the environment; returning a collection of env
	// variables that are missing
	HasKeys([]string) (bool, []string)
}


type OsEnvReader struct{
	lookup func(key string) (string, bool)
}

func NewOsEnvReader() *OsEnvReader {
	return &OsEnvReader {
		lookup: os.LookupEnv,
	}
}


func (env *OsEnvReader) LookupEnv(key string) (string, bool) {
	return env.lookup(key)
}

func (env *OsEnvReader) HasKeys(keys []string) (bool, []string) {
	missingKeys := []string{}
	for _, key := range keys {
		if _, ok := env.LookupEnv(key); !ok {
			missingKeys = append(missingKeys, key)
		}
	}

	return len(missingKeys) == 0, missingKeys
}

type EnvUnmarshaler interface {
	UnmarshalEnv(EnvReader) error
}

type Marshaler interface {
	Unmarshal(interface{}) error
}

type DefaultEnvMarshaler struct {
	Environment EnvReader
}

func (marshaler *DefaultEnvMarshaler) implementsUnmarshal(t reflect.Type) bool {
	modelType := reflect.TypeOf((*EnvUnmarshaler)(nil)).Elem()
	return reflect.PtrTo(t).Implements(modelType)
}

func (marshaler *DefaultEnvMarshaler) unmarshalField(
	fieldStruct reflect.StructField,
	structFieldVal reflect.Value,
	fieldEnvTag string,
	parser *DefaultParser,
) error {
	structFieldType := structFieldVal.Type()
	fieldName := fieldStruct.Name

	if structFieldType.Kind() == reflect.Ptr {
		indirectType := structFieldType.Elem()
		if indirectType.Kind() == reflect.Struct {
			indirectVal, unmarshalErr := marshaler.unmarshalStruct(
				indirectType, fieldEnvTag)
			if unmarshalErr != nil {
				return errors.Wrapf(
					unmarshalErr,
					"Cannot unmarshal %s to field %s in type",
					fieldEnvTag,
					fieldName,
				)
			}
			structFieldVal.Set(indirectVal.Addr())
			return nil
		}

		envVal, hasVal := marshaler.Environment.LookupEnv(fieldEnvTag)
		if !hasVal {
			return errors.Errorf(
				"Cannot retrieve any value from environment var %s",
				fieldEnvTag,
			)
		}
		indirectVal, parseErr := parser.ParseType(envVal, indirectType)
		if parseErr != nil {
			return errors.Wrapf(parseErr,
				"Cannot unmarshal %s to field %s in type (Env: %s)",
				fieldEnvTag,
				fieldName,
				envVal,
			)
		}
		structFieldVal.Set(indirectVal.Addr())
		return nil

	}

	if structFieldType.Kind() == reflect.Struct {
		fieldVal, err := marshaler.unmarshalStruct(
			structFieldType, fieldEnvTag)
		if err != nil {
			return errors.Wrapf(
				err,
				"Cannot unmarshal %s to field %s in type",
				fieldEnvTag,
				fieldName,
			)
		}
		structFieldVal.Set(fieldVal)
		return nil
	}

	envVal, hasVal := marshaler.Environment.LookupEnv(fieldEnvTag)
	if !hasVal {
		return errors.Errorf(
			"Cannot retrieve any value from environment var %s",
			fieldEnvTag,
		)
	}
	fieldVal, parseErr := parser.ParseType(envVal, structFieldType)
	if parseErr != nil {
		return errors.Wrapf(parseErr,
			"Cannot unmarshal %s to field %s in type (Env: %s)",
			fieldEnvTag,
			fieldName,
			envVal,
		)
	}
	structFieldVal.Set(fieldVal)
	return nil
}

func (marshaler *DefaultEnvMarshaler) unmarshalStruct(t reflect.Type, envPrefix string) (reflect.Value, error) {
	val := reflect.New(t).Elem()
	parser := &DefaultParser{}

	tKind := t.Kind()
	if tKind == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			fieldStruct := t.Field(i)
			fieldEnvTag := fieldStruct.Tag.Get("env")

			if fieldEnvTag != "" {
				fieldEnvTag = envPrefix + fieldEnvTag
				structFieldVal := val.Field(i)
				err := marshaler.unmarshalField(
					fieldStruct,
					structFieldVal,
					fieldEnvTag,
					parser)
				if err != nil {
					return val, err
				}
			}
		}

		return val, nil
	}

	return val, errors.Errorf("Cannot unmarshal non-struct type %s", tKind)
}

func (marshaler *DefaultEnvMarshaler) Unmarshal(i interface{}) error {
	t := reflect.TypeOf(i)
	v := reflect.ValueOf(i)
	if t.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
		t = v.Type()
	}

	// if the object implements EnvUnmarshaler, then use UnmarshalEnv method
	// of the type
	if marshaler.implementsUnmarshal(t) {
		envUnmarsh, _ := i.(EnvUnmarshaler)
		err := envUnmarsh.UnmarshalEnv(marshaler.Environment)
		if err != nil {
			return err
		}

		return nil
	}

	if t.Kind() == reflect.Struct {
		val, err := marshaler.unmarshalStruct(t, "")
		if err != nil {
			return err
		}
		v.Set(val)
		return nil
	}

	return errors.New("Cannot unmarshal non-struct, non-EnvMarshaler objects.")
}