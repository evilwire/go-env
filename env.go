// Package that supports unmarshalling objects from environment variables by
// defining tags appended to fields. The idea here is that there are a lot of
// applications whose config objects are serialised from environment variable
// values.
//
// Consider the following example
//
// 	type CassandraConfig struct {
// 		Hosts 		[]string `env: "CASSANDRA_HOSTS"`
//		Port  		int	 `env: "CASSANDRA_PORT"`
//		Consistency	string	 `env: "CASSANDRA_CONSISTENCY"`
//	 }
//
// 	func main() {
// 		// setting up the config
//		unmarshaller := DefaultEnvMarshaler {
//			Environment: NewOsEnvReader(),
//		}
//		config := CassandraConfig{}
//		unmarshaller.Unmarshal(&config)
//
//		// application logic
//		// ...
//	 }
//
// We believe that the above is pretty straightforward and has a similar
// flavor to the `encoding/json` library.
//
// At this juncture, the unmarshalling is not thread-safe. Explicit synchronisation
// logic is needed to achieve atomicity in code.
//
package goenv

import (
	"github.com/pkg/errors"
	"os"
	"reflect"
)

// Interface for that expresses the ability to look up values from the environment
// via environment variables (LookupEnv) and the ability to query the existence of
// many environment variables at once.
type EnvReader interface {

	// look up the value for a particular env variable
	// returning false if the variable is not registered
	LookupEnv(string) (string, bool)

	// returns whether or not env variables are set in
	// the environment; returning a collection of env
	// variables that are missing
	HasKeys([]string) (bool, []string)
}

// An environment variable reader that implements that EnvReader interface by using the
// os.LookupEnv method.
type OsEnvReader struct {
	lookup func(key string) (string, bool)
}

// Creates a new instance of OsEnvReader
func NewOsEnvReader() *OsEnvReader {
	return &OsEnvReader{
		lookup: os.LookupEnv,
	}
}

// Lookup a certain environment variable by name. Returns the value of the
// environment variable if the variable exists and has an assigned value. Otherwise,
// returns an unspecific value, and the exists flag is set to false.
func (env *OsEnvReader) LookupEnv(key string) (string, bool) {
	return env.lookup(key)
}

// Returns whether or not a set of environment variables have corresponding
// values along with a list of environment variables that do not have values.
func (env *OsEnvReader) HasKeys(keys []string) (bool, []string) {
	missingKeys := []string{}
	for _, key := range keys {
		if _, ok := env.LookupEnv(key); !ok {
			missingKeys = append(missingKeys, key)
		}
	}

	return len(missingKeys) == 0, missingKeys
}

// An interface for any object that defines the UnmarshalEnv method, i.e. a
// method that accepts an EnvReader and can unmarshal from environment variable
// values from the EnvReader
type EnvUnmarshaler interface {
	UnmarshalEnv(EnvReader) error
}

// An interface for any object that implements the Unmarshal method.
type Marshaler interface {
	Unmarshal(interface{}) error
}

// An unmarshaller that uses the DefaultParser and a specific environment reader
// to unmarshal primitive and derived values.
type DefaultEnvMarshaler struct {
	Environment EnvReader
}

// Determines whether or not a specific object type (represented as reflect.Type)
// implements the EnvUnMarshaler interface.
func (marshaler *DefaultEnvMarshaler) implementsUnmarshal(t reflect.Type) bool {
	modelType := reflect.TypeOf((*EnvUnmarshaler)(nil)).Elem()
	return reflect.PtrTo(t).Implements(modelType)
}

// Unmarshals a field in a struct.
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

// Recursively unmarshals a struct.
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

// Unmarshals a given value from environment variables. It accepts a pointer to a given
// object, and either succeeds in unmarshalling the object or returns an error.
//
// Usage:
//
//	 import "github.com/evilwire/go-env"
//
//	 type CassandraConfig struct {
//		Hosts 		[]string `env: "CASSANDRA_HOSTS"`
//		Port  		int	 `env: "CASSANDRA_PORT"`
//		Consistency	string	 `env: "CASSANDRA_CONSISTENCY"`
//	 }
//
//	 func main() {
//		// setting up the config
//		unmarshaller := goenv.DefaultEnvMarshaler {
//			Environment: goenv.NewOsEnvReader(),
//		}
//		config := CassandraConfig{}
//		unmarshaller.Unmarshal(&config)
//
//		// application logic
//		// ...
//	 }
//
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
		return envUnmarsh.UnmarshalEnv(marshaler.Environment)
	}

	if t.Kind() != reflect.Struct {
		return errors.New("Cannot unmarshal non-struct, non-EnvMarshaler objects.")
	}

	val, err := marshaler.unmarshalStruct(t, "")
	if err == nil {
		v.Set(val)
	}
	return err
}
