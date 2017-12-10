package goenv

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestUnmarshalString(t *testing.T) {
	marshaler := &DefaultParser{}

	cases := []string{
		"1213",
		"",
		"the apples are large",
		"Die Äpfel sind groß",
		"苹果很大",
	}

	for _, c := range cases {
		var str string
		err := marshaler.Unmarshal(c, &str)
		if err != nil {
			t.Errorf("Should not get error when unmarshaling string.")
		}

		if str != c {
			t.Errorf("Expect marshal of %s but received \"%s\" instead", c, str)
		}

	}
}

func TestUnmarshalStringFail(t *testing.T) {
	strIface := reflect.Zero(reflect.TypeOf("")).Interface()
	marshaler := &DefaultParser{}
	err := marshaler.Unmarshal("hello", strIface)
	if err == nil {
		t.Error("Expecting an error")
	}

	if s, ok := strIface.(string); !ok {
		t.Errorf("We expect %s to be a string", s)
	} else if s == "hello" {
		t.Error("`s` should NOT be set.")
	}
}

func TestUnmarshalStructFail(t *testing.T) {
	strIface := reflect.Zero(reflect.TypeOf(struct{}{})).Interface()
	marshaler := &DefaultParser{}
	err := marshaler.Unmarshal("unsettable", strIface)
	if err == nil {
		t.Error("Expecting an error")
	}
}

func TestUnmarshalBool(t *testing.T) {
	marshaler := DefaultParser{}

	cases := []struct {
		StrVal   string
		Expected bool
	}{
		{"true", true},
		{"false", false},
		{"True", true},
		{"tRUE", true},
		{"TRUE", true},
		{"False", false},
		{"fALsE", false},
		{"FALSE", false},
	}

	for _, c := range cases {
		var b bool
		err := marshaler.Unmarshal(c.StrVal, &b)

		if err != nil {
			t.Errorf("Should not get error when unmarshaling bool.")
		}
	}
}

func TestUnmarshalBoolFail(t *testing.T) {
	marshaler := &DefaultParser{}
	cases := []string{
		"not_true",
		"yes",
		"no",
		"bugger",
		"",
	}

	for _, c := range cases {
		var b bool
		err := marshaler.Unmarshal(c, &b)
		if err == nil {
			t.Errorf("Should not be able to marshal \"%s\" into a bool.", c)
		}
	}
}

func TestUnmarshalUint8(t *testing.T) {
	marshaler := &DefaultParser{}

	cases := []struct {
		StrVal   string
		Expected uint8
	}{
		{"0", 0},
		{"1", 1},
		{"4", 4},
		{"13", 13},
		{"255", 255},
	}

	for _, c := range cases {
		var v uint8
		err := marshaler.Unmarshal(c.StrVal, &v)
		if err != nil {
			t.Errorf("Should not get error when unmarshaling uint8.")
		}

		if v != c.Expected {
			t.Errorf("Expect marshal of %d but received %d instead", c.Expected, v)
		}

	}
}

func TestUnmarshalUint16(t *testing.T) {
	marshaler := &DefaultParser{}
	cases := []struct {
		StrVal   string
		Expected uint16
	}{
		{"0", 0},
		{"1", 1},
		{"4", 4},
		{"13", 13},
		{"256", 256},
		{"65534", 65534},
		{"65535", 65535},
	}

	for _, c := range cases {
		var v uint16
		err := marshaler.Unmarshal(c.StrVal, &v)

		if err != nil {
			t.Errorf("Should not get error when unmarshaling uint16.")
		}

		if v != c.Expected {
			t.Errorf("Expect marshal of %d but received %d instead", c.Expected, v)
		}

	}
}

func TestUnmarshalUint32(t *testing.T) {
	marshaler := &DefaultParser{}
	cases := []struct {
		StrVal   string
		Expected uint32
	}{
		{"0", 0},
		{"1", 1},
		{"4", 4},
		{"13", 13},
		{"256", 256},
		{"65536", 65536},
		{"4294967294", 4294967294},
		{"4294967295", 4294967295},
	}

	for _, c := range cases {
		var v uint32
		err := marshaler.Unmarshal(c.StrVal, &v)

		if err != nil {
			t.Errorf("Should not get error when unmarshaling uint32.")
		}

		if v != c.Expected {
			t.Errorf("Expect marshal of %d but received %d instead", c.Expected, v)
		}
	}
}

func TestUnmarshalUint64(t *testing.T) {
	marshaler := &DefaultParser{}
	cases := []struct {
		StrVal   string
		Expected uint64
	}{
		{"0", 0},
		{"1", 1},
		{"4", 4},
		{"13", 13},
		{"18446744073709551614", 18446744073709551614},
		{"18446744073709551615", 18446744073709551615},
	}

	for _, c := range cases {
		var v uint64
		err := marshaler.Unmarshal(c.StrVal, &v)

		if err != nil {
			t.Errorf("Should not get error when unmarshaling uint64.")
		}

		if v != c.Expected {
			t.Errorf("Expect marshal of %d but received %d instead", c.Expected, v)
		}
	}
}

func TestUnmarshalUint(t *testing.T) {
	marshaler := &DefaultParser{}
	cases := []struct {
		StrVal   string
		Expected uint
	}{
		{"0", 0},
		{"1", 1},
		{"4", 4},
		{"13", 13},

		// ensure the test is independent of uint max size
		{fmt.Sprintf("%d", ^uint(0)-1), ^uint(0) - 1},
		{fmt.Sprintf("%d", ^uint(0)), ^uint(0)},
	}

	for _, c := range cases {
		var v uint
		err := marshaler.Unmarshal(c.StrVal, &v)

		if err != nil {
			t.Errorf("Should not get error when unmarshaling uint.")
		}

		if v != c.Expected {
			t.Errorf("Expect marshal of %d but received %d instead", c.Expected, v)
		}
	}
}

func TestUnmarshalUint8Fail(t *testing.T) {
	cases := []string{
		"256",
		"-12",
		"abc",
		"",
		"123.12",
	}
	marshaler := DefaultParser{}

	for _, c := range cases {
		var v uint8
		err := marshaler.Unmarshal(c, &v)
		if err == nil {
			t.Errorf("Should not be able to marshal \"%s\" into an uint8.", c)
		}
	}
}

func TestUnmarshalUint16Fail(t *testing.T) {
	cases := []string{
		"65536",
		"-12",
		"abc",
		"",
		"123.12",
	}
	marshaler := DefaultParser{}

	for _, c := range cases {
		var v uint16
		err := marshaler.Unmarshal(c, &v)
		if err == nil {
			t.Errorf("Should not be able to marshal \"%s\" into an uint16.", c)
		}
	}
}

func TestUnmarshalUint32Fail(t *testing.T) {
	cases := []string{
		"4294967296",
		"-12",
		"abc",
		"",
		"123.12",
	}
	marshaler := DefaultParser{}

	for _, c := range cases {
		var v uint32
		err := marshaler.Unmarshal(c, &v)
		if err == nil {
			t.Errorf("Should not be able to marshal \"%s\" into an uint32.", c)
		}
	}
}

func TestUnmarshalUint64Fail(t *testing.T) {
	cases := []string{
		"18446744073709551616",
		"-12",
		"abc",
		"",
		"123.12",
	}
	marshaler := DefaultParser{}

	for _, c := range cases {
		var v uint64
		err := marshaler.Unmarshal(c, &v)
		if err == nil {
			t.Errorf("Should not be able to marshal \"%s\" into an uint64.", c)
		}
	}
}

func TestUnmarshalUintFail(t *testing.T) {
	cases := []string{
		"-12",
		"abc",
		"",
		"123.12",
	}
	marshaler := DefaultParser{}

	for _, c := range cases {
		var v uint
		err := marshaler.Unmarshal(c, &v)
		if err == nil {
			t.Errorf("Should not be able to marshal \"%s\" into an uint.", c)
		}
	}
}

func TestUnmarshalInt8(t *testing.T) {
	marshaler := &DefaultParser{}

	cases := []struct {
		StrVal   string
		Expected int8
	}{
		{"0", 0},
		{"1", 1},
		{"4", 4},
		{"13", 13},
		{"127", 127},
		{"-1", -1},
		{"-4", -4},
		{"-128", -128},
	}

	for _, c := range cases {
		var v int8
		err := marshaler.Unmarshal(c.StrVal, &v)
		if err != nil {
			t.Errorf("Should not get error when unmarshaling int8.")
		}

		if v != c.Expected {
			t.Errorf("Expect marshal of %d but received %d instead", c.Expected, v)
		}

	}
}

func TestUnmarshalInt16(t *testing.T) {
	marshaler := &DefaultParser{}

	cases := []struct {
		StrVal   string
		Expected int16
	}{
		{"0", 0},
		{"1", 1},
		{"4", 4},
		{"13", 13},
		{"128", 128},
		{"256", 256},
		{"32767", 32767},
		{"-1", -1},
		{"-4", -4},
		{"-256", -256},
		{"-32768", -32768},
	}

	for _, c := range cases {
		var v int16
		err := marshaler.Unmarshal(c.StrVal, &v)
		if err != nil {
			t.Errorf("Should not get error when unmarshaling int16.")
		}

		if v != c.Expected {
			t.Errorf("Expect marshal of %d but received %d instead", c.Expected, v)
		}

	}
}

func TestUnmarshalInt32(t *testing.T) {
	marshaler := &DefaultParser{}

	cases := []struct {
		StrVal   string
		Expected int32
	}{
		{"0", 0},
		{"1", 1},
		{"4", 4},
		{"13", 13},
		{"32768", 32768},
		{"65536", 65536},
		{"2147483647", 2147483647},
		{"-1", -1},
		{"-4", -4},
		{"-256", -256},
		{"-32768", -32768},
		{"-2147483648", -2147483648},
	}

	for _, c := range cases {
		var v int32
		err := marshaler.Unmarshal(c.StrVal, &v)
		if err != nil {
			t.Errorf("Should not get error when unmarshaling int32.")
		}

		if v != c.Expected {
			t.Errorf("Expect marshal of %d but received %d instead", c.Expected, v)
		}

	}
}

func TestUnmarshalInt64(t *testing.T) {
	marshaler := &DefaultParser{}

	cases := []struct {
		StrVal   string
		Expected int64
	}{
		{"0", 0},
		{"1", 1},
		{"4", 4},
		{"13", 13},
		{"9223372036854775807", 9223372036854775807},
		{"-1", -1},
		{"-4", -4},
		{"-256", -256},
		{"-9223372036854775808", -9223372036854775808},
	}

	for _, c := range cases {
		var v int64
		err := marshaler.Unmarshal(c.StrVal, &v)
		if err != nil {
			t.Errorf("Should not get error when unmarshaling int64.")
		}

		if v != c.Expected {
			t.Errorf("Expect marshal of %d but received %d instead", c.Expected, v)
		}

	}
}

func TestUnmarshalInt(t *testing.T) {
	marshaler := &DefaultParser{}

	cases := []struct {
		StrVal   string
		Expected int
	}{
		{"0", 0},
		{"1", 1},
		{"4", 4},
		{"13", 13},
		{"32768", 32768},
		{"65536", 65536},
		{"2147483647", 2147483647},
		{"-1", -1},
		{"-4", -4},
		{"-256", -256},
		{"-32768", -32768},
		{"-2147483648", -2147483648},
	}

	for _, c := range cases {
		var v int
		err := marshaler.Unmarshal(c.StrVal, &v)
		if err != nil {
			t.Errorf("Should not get error when unmarshaling int.")
		}

		if v != c.Expected {
			t.Errorf("Expect marshal of %d but received %d instead", c.Expected, v)
		}
	}
}

func TestUnmarshalInt8Fail(t *testing.T) {
	cases := []string{
		"",
		"-129",
		"128",
		"256",
		"abc",
		"123.12",
		"123.0",
	}
	marshaler := DefaultParser{}

	for _, c := range cases {
		var v int8
		err := marshaler.Unmarshal(c, &v)
		if err == nil {
			t.Errorf("Should not be able to marshal \"%s\" into an int8.", c)
		}
	}
}

func TestUnmarshalInt16Fail(t *testing.T) {
	cases := []string{
		"",
		"-32769",
		"32768",
		"65536",
		"abc",
		"123.12",
		"123.0",
	}
	marshaler := DefaultParser{}

	for _, c := range cases {
		var v int16
		err := marshaler.Unmarshal(c, &v)
		if err == nil {
			t.Errorf("Should not be able to marshal \"%s\" into an int8.", c)
		}
	}
}

func TestUnmarshalInt32Fail(t *testing.T) {
	cases := []string{
		"",
		"abc",
		"123.12",
		"123.0",
	}
	marshaler := DefaultParser{}

	for _, c := range cases {
		var v int32
		err := marshaler.Unmarshal(c, &v)
		if err == nil {
			t.Errorf("Should not be able to marshal \"%s\" into an int8.", c)
		}
	}
}

func TestUnmarshalFloat(t *testing.T) {
	marshaler := &DefaultParser{}

	cases := []struct {
		StrVal   string
		Expected float64
	}{
		{"0", 0.0},
		{"0.0", 0.0},
		{"4", 4.0},
		{"0.212", 0.212},
		{"9223372036854.775807", 9223372036854.775807},
		{"-1.0", -1.0},
		{"-4.0", -4.0},
		{"-2.56", -2.56},
		{"-922.3372036854775808", -922.3372036854775808},
	}

	for _, c := range cases {
		var v float64
		err := marshaler.Unmarshal(c.StrVal, &v)
		if err != nil {
			t.Errorf("Should not get error when unmarshaling float64.")
		}

		if v != c.Expected {
			t.Errorf("Expect marshal of %.2f but received %.2f instead", c.Expected, v)
		}

	}
}

func TestUnmarshalFloat32Fail(t *testing.T) {
	cases := []string{
		"",
		"1e100",
		"1,200.00",
		"1.200,00",

		// v--- this should totally be legit, silly yanks!
		"1,20",
		"abc",
	}
	marshaler := DefaultParser{}

	for _, c := range cases {
		var v float32
		err := marshaler.Unmarshal(c, &v)
		if err == nil {
			t.Errorf("Should not be able to marshal \"%s\" into an float32.", c)
		}
	}
}

func TestUnmarshalStringSlice(t *testing.T) {
	marshaler := &DefaultParser{}

	cases := []struct {
		StrVal   string
		Expected []string
	}{
		{"a", []string{"a"}},
		{"", []string{}},
		{"a,b", []string{"a", "b"}},
		{"a,b,c,d,abc", []string{"a", "b", "c", "d", "abc"}},
		{",", []string{"", ""}},
		{"a ,b", []string{"a", "b"}},
		{"a, b,", []string{"a", "b", ""}},
		{",a, b,  ", []string{"", "a", "b", ""}},
		{"a,,,b", []string{"a", "", "", "b"}},
	}

	for _, c := range cases {
		var a []string
		err := marshaler.Unmarshal(c.StrVal, &a)

		if err != nil {
			t.Errorf("Unmarshal should not raise error when handling \"%s\"", c.StrVal)
		} else {
			if len(c.Expected) != len(a) {
				t.Errorf(
					"The expected length differs to actual length. "+
						"Expected: %d, actual: %d (marshalling \"%s\")",
					len(c.Expected),
					len(a),
					c.StrVal,
				)
			}

			for i, elt := range c.Expected {
				if a[i] != elt {
					t.Errorf("Expected element %d: %s, actual: %s",
						i,
						c.Expected[i],
						a[i],
					)
				}
			}
		}
	}
}

func TestUnmarshalIntSlice(t *testing.T) {
	marshaler := &DefaultParser{}

	cases := []struct {
		StrVal   string
		Expected []int
	}{
		{"1", []int{1}},
		{"1,2,3", []int{1, 2, 3}},
		{"-4,0,0,1", []int{-4, 0, 0, 1}},
		{"", []int{}},
	}

	for _, c := range cases {
		var a []int
		err := marshaler.Unmarshal(c.StrVal, &a)

		if err != nil {
			t.Errorf("Unmarshal should not raise error when handling \"%s\"", c.StrVal)
		} else {
			if len(c.Expected) != len(a) {
				t.Errorf(
					"Expected length differs to actual length. "+
						"Expected: %d, actual: %d (marshalling \"%s\")",
					len(c.Expected),
					len(a),
					c.StrVal,
				)
			}

			for i, elt := range c.Expected {
				if a[i] != elt {
					t.Errorf("Expected element %d: %d, actual: %d",
						i,
						c.Expected[i],
						a[i])
				}
			}
		}
	}
}

func TestUnmarshalUIntSliceFail(t *testing.T) {
	marshaler := &DefaultParser{}

	cases := []string{
		"a",
		",",
		"1,2,",
		"-1,0,1",
		"0,0.1",
		"3,1,1,b",
		"-1,,-2",
	}

	for _, c := range cases {
		var v []uint
		err := marshaler.Unmarshal(c, &v)
		if err == nil {
			t.Errorf("Should not be able to marshal \"%s\" into []uint.", c)
		}
	}
}

func TestUnmarshalDuration(t *testing.T) {
	marshaler := &DefaultParser{}
	cases := []struct {
		StrVal   string
		Expected time.Duration
	}{
		{"1ns", 1 * time.Nanosecond},
		{"1us", 1 * time.Microsecond},
		{"1ms", 1 * time.Millisecond},
		{"1s", 1 * time.Second},
		{"1m", 1 * time.Minute},
		{"1h", 1 * time.Hour},
		{"1h2m", 1*time.Hour + 2*time.Minute},
		{"-1m", -1 * time.Minute},
		{"-1h30m", -1*time.Hour - 30*time.Minute},
		{"1h2m200us", 1*time.Hour + 2*time.Minute + 200*time.Microsecond},
	}

	for _, c := range cases {
		var d time.Duration
		err := marshaler.Unmarshal(c.StrVal, &d)

		if err != nil {
			t.Errorf("Unmarshal should not raise error when handling \"%s\"", c.StrVal)
		} else {
			if d != c.Expected {
				t.Errorf("Expected %s, received %s instead",
					c.Expected.String(),
					d.String(),
				)
			}
		}
	}
}

func TestUnmarshalDurationFail(t *testing.T) {

	marshaler := DefaultParser{}
	cases := []string{
		"2 hours",
		"h3ms",
		"s",
		"30min",
		"1h-30m10s",
		"",
	}

	for _, c := range cases {
		var d time.Duration

		err := marshaler.Unmarshal(c, &d)
		if err == nil {
			t.Errorf("Should not be able to marshal \"%s\" into time.Duration.", c)
		}
	}
}

func TestUnmarshalUnknownObjFail(t *testing.T) {
	marshaler := DefaultParser{}
	obj := struct{ A uint }{}

	err := marshaler.Unmarshal("1", &obj)
	if err == nil {
		t.Error("We expect the parser to fail for struct types.")
	}
}

func TestParseType(t *testing.T) {
	var a *uint
	uintPtr := reflect.TypeOf(a)

	marshaler := DefaultParser{}
	testVal := uint(3141589)
	val, err := marshaler.ParseType(fmt.Sprintf("%d", testVal), uintPtr)
	if err != nil {
		t.Error("We expect parse to succeed for uint pointer.")
	}

	if val.Type().Kind() != reflect.Ptr || val.Type().Elem().Kind() != reflect.Uint {
		t.Error("We expected the type of the value to be an uint pointer.")
	}

	actualVal := val.Elem().Uint()
	if uint(actualVal) != testVal {
		t.Errorf("Expected: %d, Actual: %d", testVal, actualVal)
	}
}

func TestParseTypeFail(t *testing.T) {
	var a *uint
	uintPtr := reflect.TypeOf(a)

	marshaler := DefaultParser{}
	_, err := marshaler.ParseType("-1", uintPtr)
	if err == nil {
		t.Error("We expect parse to fail for incorrect pointer.")
	}
}
