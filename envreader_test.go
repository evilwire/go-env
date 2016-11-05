package util

import (
	"testing"
	"reflect"
	"os"
)


type LookupEnvMock struct {
	EnvVars map[string]string
	Arg string
}


func (mock *LookupEnvMock) LookupEnv(key string) (string, bool) {
	mock.Arg = key
	val, ok := mock.EnvVars[key]
	return val, ok
}


func TestNewOsEnvReader(t *testing.T) {
	envReader := NewOsEnvReader()

	lookupEnv := reflect.ValueOf(envReader.lookup)
	osLookupEnv := reflect.ValueOf(os.LookupEnv)

	if lookupEnv.Pointer() != osLookupEnv.Pointer() {
		t.Error("Expects the lookup function to be os.LookupEnv")
	}
}


func TestOsEnvReader_LookupEnv(t *testing.T) {
	osEnv := map[string]string {
		"A": "hello",
		"B": "",
	}

	testCases := []struct {
		Key string
		HasKey bool
		Expected string
	}{
		{
			"A",
			true,
			"hello",
		},
		{
			"B",
			true,
			"",
		},
		{
			"C",
			false,
			"",
		},
	}


	for i, c := range testCases {
		mockOs := LookupEnvMock{
			EnvVars: osEnv,
		}

		envReader := OsEnvReader {
			lookup: mockOs.LookupEnv,
		}

		val, exists := envReader.LookupEnv(c.Key)

		if exists != c.HasKey {
			t.Errorf("TC %d: Does env var %s have value? Expected %t, actual %t",
				i,
				c.Key,
				c.HasKey,
				exists,
			)
		}

		if c.HasKey && val != c.Expected {
			t.Errorf("TC %d: Expect value to be %s, actual %s",
				i,
				c.Expected,
				val,
			)
		}
	}
}


func contains(v string, b []string) bool {
	for _, bV := range b {
		if bV == v {
			return true
		}
	}
	return false
}


func isSubsetOf(a, b []string) bool {
	for _, v := range a {
		if !contains(v, b) {
			return false
		}
	}

	return true
}


func sameKeys(a, b []string) bool {
	return isSubsetOf(a, b) && isSubsetOf(b, a)
}


func TestOsEnvReader_HasKeys(t *testing.T) {
	testCases := []struct {
		Env map[string]string
		TestKeys []string
		ExpectHasKeys bool
		ExpectMissingKeys []string
	}{
		{
			Env: map[string]string {
				"A": "hello",
				"B": "goodbye",
				"C": "",
			},
			TestKeys: []string {
				"A",
			},
			ExpectHasKeys: true,
			ExpectMissingKeys: []string {},
		},
		{
			Env: map[string]string {
				"A": "hello",
				"B": "goodbye",
				"C": "",
			},
			TestKeys: []string {
			},
			ExpectHasKeys: true,
			ExpectMissingKeys: []string {},
		},
		{
			Env: map[string]string {
				"A": "hello",
				"B": "goodbye",
				"C": "",
			},
			TestKeys: []string {
				"A", "B", "C",
			},
			ExpectHasKeys: true,
			ExpectMissingKeys: []string {},
		},
		{
			Env: map[string]string {
				"A": "hello",
				"B": "goodbye",
				"C": "",
			},
			TestKeys: []string {
				"D",
			},
			ExpectHasKeys: false,
			ExpectMissingKeys: []string {
				"D",
			},
		},
		{
			Env: map[string]string {
				"A": "hello",
				"B": "goodbye",
				"C": "",
			},
			TestKeys: []string {
				"D", "E",
			},
			ExpectHasKeys: false,
			ExpectMissingKeys: []string {
				"D", "E",
			},
		},
		{
			Env: map[string]string {
				"A": "hello",
				"B": "goodbye",
				"C": "",
			},
			TestKeys: []string {
				"A", "D", "E",
			},
			ExpectHasKeys: false,
			ExpectMissingKeys: []string {
				"D", "E",
			},
		},
	}

	for i, c := range testCases {
		mockOs := LookupEnvMock{
			EnvVars: c.Env,
		}

		envreader := OsEnvReader {
			lookup: mockOs.LookupEnv,
		}

		hasKeys, missingKeys := envreader.HasKeys(c.TestKeys)

		if hasKeys != c.ExpectHasKeys {
			t.Errorf("TC %d: Has Keys? Expected %t, actual %t",
				i,
				c.ExpectHasKeys,
				hasKeys,
			)
		}

		if !sameKeys(missingKeys, c.ExpectMissingKeys) {
			t.Errorf("TC %d: Expect missing keys %v, actual %v",
				i,
				c.ExpectMissingKeys,
				missingKeys,
			)
		}
	}
}