// Package goenv supports unmarshalling objects from environment variables by
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

// EnvReader is an interface for expressing the ability to look up values from the environment
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

// OsEnvReader is an environment variable reader that implements that EnvReader interface by using the
// os.LookupEnv method.
type OsEnvReader struct {
	lookup func(key string) (string, bool)
}

// NewOsEnvReader creates a new instance of OsEnvReader
func NewOsEnvReader() *OsEnvReader {
	return &OsEnvReader{
		lookup: os.LookupEnv,
	}
}

// LookupEnv - Lookup a certain environment variable by name. Returns the value of the
// environment variable if the variable exists and has an assigned value. Otherwise,
// returns an unspecific value, and the exists flag is set to false.
func (env *OsEnvReader) LookupEnv(key string) (string, bool) {
	return env.lookup(key)
}

// HasKeys - Returns whether or not a set of environment variables have corresponding
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

// EnvUnmarshaler is an interface for any object that defines the UnmarshalEnv method, i.e. a
// method that accepts an EnvReader and can unmarshal from environment variable
// values from the EnvReader
type EnvUnmarshaler interface {
	UnmarshalEnv(EnvReader) error
}

// Marshaler - An interface for any object that implements the Unmarshal method.
type Marshaler interface {
	Unmarshal(interface{}) error
}

// DefaultEnvMarshaler - An unmarshaller that uses the DefaultParser and a specific environment reader
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

func (marshaler *DefaultEnvMarshaler) unmarshalType(
	fieldType reflect.Type, fieldEnvTag string, parser *DefaultParser,
) (*reflect.Value, error) {
	envVal, hasVal := marshaler.Environment.LookupEnv(fieldEnvTag)
	if !hasVal {
		return nil, errors.Errorf(
			"cannot retrieve any value from environment var %s",
			fieldEnvTag,
		)
	}

	fieldVal, parseErr := parser.ParseType(envVal, fieldType)
	if parseErr != nil {
		return nil, errors.Wrapf(parseErr,
			"cannot unmarshal %s to type %s (Env: %s)",
			envVal,
			fieldType.Name(),
			fieldEnvTag,
		)
	}

	return &fieldVal, nil
}

func (marshaler *DefaultEnvMarshaler) unmarshalNonPtr(
	fieldType reflect.Type,
	fieldEnvTag string,
	parser *DefaultParser,
) (*reflect.Value, error) {
	if fieldType.Name() == "Time" {
		return marshaler.unmarshalType(fieldType, fieldEnvTag, parser)
	}

	if fieldType.Kind() == reflect.Struct {
		fieldVal, err := marshaler.unmarshalStruct(fieldType, fieldEnvTag)
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"cannot unmarshal %s to type %s",
				fieldEnvTag,
				fieldType.Name(),
			)
		}
		return &fieldVal, nil
	}

	return marshaler.unmarshalType(fieldType, fieldEnvTag, parser)
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
		indirectVal, unmarshErr := marshaler.unmarshalNonPtr(indirectType, fieldEnvTag, parser)
		if unmarshErr != nil {
			return errors.Wrapf(unmarshErr, "error unmarshaling field %s", fieldName)
		}
		structFieldVal.Set(indirectVal.Addr())
		return nil

	}

	fieldVal, unmarshErr := marshaler.unmarshalNonPtr(structFieldType, fieldEnvTag, parser)
	if unmarshErr != nil {
		return errors.Wrapf(unmarshErr, "error unmarshaling field %s", fieldName)
	}

	structFieldVal.Set(*fieldVal)
	return nil
}

// Recursively unmarshals a struct.
func (marshaler *DefaultEnvMarshaler) unmarshalStruct(t reflect.Type, envPrefix string) (reflect.Value, error) {
	val := reflect.New(t).Elem()
	parser := &DefaultParser{}

	tKind := t.Kind()
	if tKind != reflect.Struct {
		return val, errors.Errorf("cannot unmarshal non-struct type %s", tKind)
	}

	for i := 0; i < t.NumField(); i++ {
		fieldStruct := t.Field(i)
		fieldEnvTag := fieldStruct.Tag.Get("env")

		if fieldEnvTag == "" {
			continue
		}

		fieldEnvTag = envPrefix + fieldEnvTag
		structFieldVal := val.Field(i)
		err := marshaler.unmarshalField(fieldStruct, structFieldVal, fieldEnvTag, parser)
		if err != nil {
			return val, err
		}
	}

	return val, nil
}

// Unmarshal - Unmarshals a given value from environment variables. It accepts a pointer to a given
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
		return errors.New("cannot unmarshal non-struct, non-EnvMarshaler objects")
	}

	val, err := marshaler.unmarshalStruct(t, "")
	if err == nil {
		v.Set(val)
	}
	return err
}
