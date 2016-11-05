package goenv

import (
	"testing"
	"time"
	"fmt"
	"errors"
	"reflect"
)

type MockEnvReader struct {
	EnvValues map[string]string
}

func (reader *MockEnvReader) LookupEnv(key string) (string, bool) {
	if val, ok := reader.EnvValues[key]; ok {
		return val, true
	}

	return "", false
}

func (reader *MockEnvReader) HasKeys(keys []string) (bool, []string) {
	missingEnvVars := []string{}
	for _, key := range keys {
		if _, exists := reader.EnvValues[key]; !exists {
			missingEnvVars = append(missingEnvVars, key)
		}
	}

	return len(missingEnvVars) == 0, missingEnvVars
}


type Equaler interface {
	fmt.Stringer
	Equal(i interface{}) bool
}

type TestCase struct {
	Env map[string]string
	Expected Equaler
}

func ref(s string) *string {
	return &s
}

func test(c TestCase, t *testing.T, obj Equaler) {
	marsh := DefaultEnvMarshaler{
		&MockEnvReader{ c.Env },
	}

	err := marsh.Unmarshal(obj)
	if err != nil {
		t.Errorf("Unmarshal should not raise error. Error: %s", err.Error())
	} else {
		if !c.Expected.Equal(obj) {
			t.Errorf("Marshalled object does not match expected. " +
			"Expected: %+v, Actual: %+v",
				c.Expected, obj,
			)
		}
	}
}


func testFail(env map[string]string, t *testing.T, obj Equaler) {
	marsh := DefaultEnvMarshaler{
		&MockEnvReader{ env },
	}

	err := marsh.Unmarshal(obj)

	if err == nil {
		t.Error("Expecting an error from unmarshalling.")
	}
}

type Obj1 struct {
	A string `env:"OBJ1_A"`
	B uint `env:"OBJ1_B"`
	C bool `env:"OBJ1_C"`
	D []int `env:"OBJ1_D"`
	E time.Duration `env:"OBJ1_E"`
}

func (o *Obj1) Equal(i interface{}) bool {
	if other, ok := i.(*Obj1); !ok {
		return false
	} else {
		firstly := other.A == o.A &&
			other.B == o.B &&
			other.C == o.C &&
			other.E == o.E;

		if !firstly {
			return false
		}

		for index, elt := range other.D {
			if other.D[index] != elt {
				return false
			}
		}
	}
	return true
}

func (o *Obj1) String() string {
	return fmt.Sprintf("%+v", o)
}

func TestUnmarshalObj1(t *testing.T) {
	cases := []TestCase {
		{
			map[string]string {
				"OBJ1_A": "hello",
				"OBJ1_B": "14",
				"OBJ1_C": "true",
				"OBJ1_D": "1, -2, 100, 3",
				"OBJ1_E": "12m",
			},
			&Obj1{
				A: "hello",
				B: 14,
				C: true,
				D: []int{1, -2, 100, 3},
				E: 12 * time.Minute,
			},
		},
		{
			map[string]string {
				"OBJ1_A": "",
				"OBJ1_B": fmt.Sprintf("%d", ^uint(0)),
				"OBJ1_C": "false",
				"OBJ1_D": "1",
				"OBJ1_E": "1h12m",
			},
			&Obj1{
				A: "",
				B: ^uint(0),
				C: false,
				D: []int{1},
				E: 1 * time.Hour + 12 * time.Minute,
			},
		},
		{
			map[string]string {
				"OBJ1_A": "亲蛙",
				"OBJ1_B": "0",
				"OBJ1_C": "true",
				"OBJ1_D": "",
				"OBJ1_E": "0ns",
			},
			&Obj1{
				A: "亲蛙",
				B: 0,
				C: true,
				D: []int{},
				E: 0 * time.Nanosecond,
			},
		},
	}

	for _, c := range cases {
		var obj Obj1
		test(c, t, &obj)
	}
}

func TestUnmarshalObj1Fail(t *testing.T) {
	cases := []map[string]string{
		map[string]string {
			"OBJ1_A": "abc",
			"OBJ1_B": "-14",
			"OBJ1_C": "true",
			"OBJ1_D": "1, -2, 100, 3",
			"OBJ1_E": "12m",
		},
		map[string]string {
			"OBJ1_B": "14",
			"OBJ1_C": "true",
			"OBJ1_D": "1, -2, 100, 3",
			"OBJ1_E": "12m",
		},
	}

	for _, c := range cases {
		var obj Obj1
		testFail(c, t, &obj)
	}
}

type Obj2 struct {
	A *string `env:"OBJ2_A"`
	B uint
}

func (o *Obj2) Equal(i interface{}) bool {
	if other, ok := i.(*Obj2); !ok {
		return false
	} else {
		return *(other.A) == *(o.A)
	}
}

func (o *Obj2) String() string {
	return fmt.Sprintf("{A: %s}", *(o.A))
}

func TestUnmarshalObj2(t *testing.T) {
	cases := []TestCase {
		{
			map[string]string {
				"OBJ2_A": "hello",
			},
			&Obj2{
				A: ref("hello"),
			},
		},
	}
	for _, c := range cases {
		var obj Obj2
		test(c, t, &obj)
	}
}

type NestedObj1 struct {
	A Obj1 `env:"NESTED_"`
	F uint `env:"NESTED_OBJ1_F"`
}

func (o *NestedObj1) Equal(i interface{}) bool {
	if other, ok := i.(*NestedObj1); !ok {
		return false
	} else {
		return other.A.Equal(&(o.A)) && other.F == o.F
	}
}

func (o *NestedObj1) String() string {
	aStr := fmt.Sprintf("%+v", o.A)
	return fmt.Sprintf("{A: %s, F: %d}", aStr, o.F)
}

func TestUnmarshalNestedObj1(t *testing.T) {
	cases := []TestCase {
		{
			map[string]string{
				"NESTED_OBJ1_A": "hello",
				"NESTED_OBJ1_B": "14",
				"NESTED_OBJ1_C": "true",
				"NESTED_OBJ1_D": "1, -2, 100, 3",
				"NESTED_OBJ1_E": "12m",
				"NESTED_OBJ1_F": "65536",
			},
			&NestedObj1{
				A: Obj1{
					A: "hello",
					B: 14,
					C: true,
					D: []int{1, -2, 100, 3},
					E: 12 * time.Minute,
				},
				F: 65536,
			},
		},
	}

	for _, c := range cases {
		var obj NestedObj1
		test(c, t, &obj)
	}
}

func TestUnmarshalNestedObj1Fail(t *testing.T) {
	cases := []map[string]string{
		map[string]string {
			"NESTED_OBJ1_A": "hello",
			"NESTED_OBJ1_B": "-14",
			"NESTED_OBJ1_C": "true",
			"NESTED_OBJ1_D": "1, -2, 100, 3",
			"NESTED_OBJ1_E": "12m",
			"NESTED_OBJ1_F": "65536",
		},
		map[string]string {
			"OBJ1_A": "abc",
			"OBJ1_B": "-14",
			"OBJ1_C": "true",
			"OBJ1_D": "1, -2, 100, 3",
			"OBJ1_E": "12m",
			"NESTED_OBJ1_F": "65536",
		},
		map[string]string {
			"NESTED_OBJ1_A": "hello",
			"NESTED_OBJ1_C": "true",
			"NESTED_OBJ1_D": "1, -2, 100, 3",
			"NESTED_OBJ1_E": "12m",
			"NESTED_OBJ1_F": "65536",
		},
		map[string]string {
			"NESTED_OBJ1_A": "hello",
			"NESTED_OBJ1_B": "14",
			"NESTED_OBJ1_C": "true",
			"NESTED_OBJ1_D": "1, -2, 100, 3",
			"NESTED_OBJ1_E": "12m",
		},
	}

	for _, c := range cases {
		var obj NestedObj1
		testFail(c, t, &obj)
	}
}

type NestedObj2 struct {
	A *Obj1 `env:"NESTED_OBJ2_"`
	B []uint `env:"NESTED_OBJ2_B"`
	C *[]uint `env:"NESTED_OBJ2_C"`
}

func (o *NestedObj2) Equal(i interface{}) bool {
	if other, ok := i.(*NestedObj2); !ok {
		return false
	} else {

		if !other.A.Equal(o.A) {
			return false
		}

		for i, b := range other.B {
			if o.B[i] != b {
				return false
			}
		}

		for i, c := range *(other.C) {
			if (*(o.C))[i] != c {
				return false
			}
		}

		return true
	}
}

func (o *NestedObj2) String() string {
	aStr := fmt.Sprintf("%+v", *(o.A))
	return fmt.Sprintf("{A: %s, B: %v, C: %v}",
		aStr, o.B, *(o.C),
	)
}

func TestUnmarshalNestedObj2(t *testing.T) {
	cases := []TestCase {
		{
			map[string]string{
				"NESTED_OBJ2_OBJ1_A": "hello",
				"NESTED_OBJ2_OBJ1_B": "14",
				"NESTED_OBJ2_OBJ1_C": "true",
				"NESTED_OBJ2_OBJ1_D": "1, -2, 100, 3",
				"NESTED_OBJ2_OBJ1_E": "12m",
				"NESTED_OBJ2_B": "0, 1, 2, 4",
				"NESTED_OBJ2_C": "0, 1, 2, 4",
			},
			&NestedObj2{
				A: &Obj1{
					A: "hello",
					B: 14,
					C: true,
					D: []int{1, -2, 100, 3},
					E: 12 * time.Minute,
				},
				B: []uint{0, 1, 2, 4},
				C: &[]uint{0, 1, 2, 4},
			},
		},
	}

	for _, c := range cases {
		var obj NestedObj2
		test(c, t, &obj)
	}
}

func TestUnmarshalNestedObj2Fail(t *testing.T) {
	cases := []map[string]string{
		map[string]string {
			"NESTED_OBJ2_OBJ1_A": "hello",
			"NESTED_OBJ2_OBJ1_B": "-14",
			"NESTED_OBJ2_OBJ1_C": "true",
			"NESTED_OBJ2_OBJ1_D": "1, -2, 100, 3",
			"NESTED_OBJ2_OBJ1_E": "12m",
			"NESTED_OBJ2_B": "0,1,2,4",
			"NESTED_OBJ2_C": "0,1,2,4",
		},
		map[string]string {
			"OBJ1_A": "abc",
			"OBJ1_B": "-14",
			"OBJ1_C": "true",
			"OBJ1_D": "1, -2, 100, 3",
			"OBJ1_E": "12m",
			"NESTED_OBJ2_B": "0,1,2,4",
			"NESTED_OBJ2_C": "0,1,2,4",
		},
		map[string]string {
			"NESTED_OBJ2_OBJ1_A": "hello",
			"NESTED_OBJ2_OBJ1_C": "true",
			"NESTED_OBJ2_OBJ1_D": "1, -2, 100, 3",
			"NESTED_OBJ2_OBJ1_E": "12m",
			"NESTED_OBJ2_B": "0,1,2,4",
			"NESTED_OBJ2_C": "0,1,2,4",
		},
		map[string]string {
			"NESTED_OBJ2_OBJ1_A": "hello",
			"NESTED_OBJ2_OBJ1_B": "14",
			"NESTED_OBJ2_OBJ1_C": "true",
			"NESTED_OBJ2_OBJ1_D": "1, -2, 100, 3",
			"NESTED_OBJ2_OBJ1_E": "12m",
			"NESTED_OBJ2_B": "0,1,2,-4",
			"NESTED_OBJ2_C": "0,1,2,4",

		},
		map[string]string {
			"NESTED_OBJ2_OBJ1_A": "hello",
			"NESTED_OBJ2_OBJ1_B": "14",
			"NESTED_OBJ2_OBJ1_C": "true",
			"NESTED_OBJ2_OBJ1_D": "1, -2, 100, 3",
			"NESTED_OBJ2_OBJ1_E": "12m",
			"NESTED_OBJ2_B": "0,1,2,4",
			"NESTED_OBJ2_C": "0,1,2,",
		},
		map[string]string {
			"NESTED_OBJ2_OBJ1_A": "hello",
			"NESTED_OBJ2_OBJ1_B": "14",
			"NESTED_OBJ2_OBJ1_C": "true",
			"NESTED_OBJ2_OBJ1_D": "1, -2, 100, 3",
			"NESTED_OBJ2_OBJ1_E": "12m",
			"NESTED_OBJ2_C": "0,1,2",
		},
		map[string]string {
			"NESTED_OBJ2_OBJ1_A": "hello",
			"NESTED_OBJ2_OBJ1_B": "14",
			"NESTED_OBJ2_OBJ1_C": "true",
			"NESTED_OBJ2_OBJ1_D": "1, -2, 100, 3",
			"NESTED_OBJ2_OBJ1_E": "12m",
			"NESTED_OBJ2_B": "0,1,2,4",
		},
	}

	for _, c := range cases {
		var obj NestedObj2
		testFail(c, t, &obj)
	}
}

type EnvMarshalerObj1 struct {
	A uint `env:"ENV_MARSHALER_OBJ1_A"`
	B string `env:"ENV_MARSHALER_OBJ1_B"`
}

func (o *EnvMarshalerObj1) Equal(i interface{}) bool {
	if other, ok := i.(*EnvMarshalerObj1); !ok {
		return false
	} else {
		return other.A == o.A && other.B == o.B
	}
}

func (o *EnvMarshalerObj1) String() string {
	return fmt.Sprintf("%+v", o)
}

func (o *EnvMarshalerObj1) UnmarshalEnv(env EnvReader) error {
	bStr, valExists := env.LookupEnv("ENV_MARSHALER_OBJ1_B")
	if !valExists {
		return errors.New("Cannot marshal UnmarshalableEnvObj1: missing UNMARSHALABLE_ENV_OBJ1_B")
	}
	o.A = 3
	o.B = bStr

	return nil
}

func TestUnmarshalEnvMarshalerObj1(t *testing.T) {
	cases := []TestCase {
		{
			map[string]string{
				"ENV_MARSHALER_OBJ1_B": "a",
			},
			&EnvMarshalerObj1{
				3, "a",
			},
		},
		{
			map[string]string{
				"ENV_MARSHALER_OBJ1_B": "",
			},
			&EnvMarshalerObj1{
				3, "",
			},
		},
		{
			map[string]string{
				"ENV_MARSHALER_OBJ1_A": "1",
				"ENV_MARSHALER_OBJ1_B": "",
			},
			&EnvMarshalerObj1{
				3, "",
			},
		},
	}

	for _, c := range cases {
		var obj EnvMarshalerObj1
		test(c, t, &obj)
	}
}

func TestUnmarshalEnvMarshalerObj1Fail(t *testing.T) {
	cases := []map[string]string {
		map[string]string {
		},
		map[string]string {
			"ENV_MARSHALER_OBJ1_A": "12",
		},
	}
	for _, c := range cases {
		var obj EnvMarshalerObj1
		testFail(c, t, &obj)
	}
}

type EnvMarshalerObj2 uint

func (o *EnvMarshalerObj2) Equal(i interface{}) bool {
	if other, ok := i.(*EnvMarshalerObj2); !ok {
		return false
	} else {
		return uint(*o) == uint(*other)
	}
}

func (o *EnvMarshalerObj2) String() string {
	return fmt.Sprintf("%d", uint(*o))
}

func (o *EnvMarshalerObj2) UnmarshalEnv(env EnvReader) error {
	*o = EnvMarshalerObj2(1)
	return nil
}

func TestUnmarshalEnvMarshalerObj2(t *testing.T) {
	envMarsh := EnvMarshalerObj2(1)
	testCase := TestCase{
		map[string]string{},
		&envMarsh,
	}

	var obj EnvMarshalerObj2
	test(testCase, t, &obj)
}

type NonEnvMarshaler uint

func (o *NonEnvMarshaler) Equal(i interface{}) bool {
	if other, ok := i.(*EnvMarshalerObj2); !ok {
		return false
	} else {
		return uint(*o) == uint(*other)
	}
}

func (o *NonEnvMarshaler) String() string {
	return fmt.Sprintf("%d", uint(*o))
}

func TestNonStructNonEnvMarshalerFail(t *testing.T) {
	var obj NonEnvMarshaler
	testFail(map[string]string{}, t, &obj)
}

func TestUnmarshalStructFailDirectly(t *testing.T) {
	marshaler := DefaultEnvMarshaler{}

	badType := reflect.TypeOf("")
	_, err := marshaler.unmarshalStruct(badType, "")
	if err == nil {
		t.Error("We do not expect to succeed unmarshaling a string in unmarshalStruct")
	}
}