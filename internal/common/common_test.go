/*
Copyright 2024 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"testing"
)

func init() {
	if err := os.Setenv("KO_DATA_PATH", "./../../kodata"); err != nil {
		log.Fatalf("failed to set env: %v", err)
	}
}

func TestPointer(t *testing.T) {
	type testCase struct {
		Name           string
		Input          any
		ExpectedResult any
	}

	for _, tc := range []testCase{
		{
			Name:  "string to pointer",
			Input: "hello world",
			ExpectedResult: func() *string {
				i := "hello world"
				return &i
			},
		},
		{
			Name:  "int to pointer",
			Input: 123,
			ExpectedResult: func() *int {
				i := 123
				return &i
			},
		},
		{
			Name:           "nil to pointer",
			ExpectedResult: nil,
		},
	} {
		result := Pointer(tc.Input)
		if reflect.ValueOf(result).Kind() != reflect.Ptr {
			t.Fatalf("error: unexpected result from test '%v', with input as '%v' and expected input as '%v'", tc.Name, tc.Input, &tc.ExpectedResult)
		}
	}
}

func TestReadFile(t *testing.T) {
	content, err := ReadFile("./testdata/file.txt")
	if err != nil {
		t.Fatalf("error: reading file; %v", err)
	}
	if content != "Hello!\n" {
		t.Fatalf("error: file content does not match expected")
	}

	content, err = ReadFile("./testdata/non-existent-file.txt")
	if err == nil || content != "" {
		t.Fatalf("error: file should not exist")
	}
}

func TestSafeError(t *testing.T) {
	inputText := "<p>Hello</p>"
	expectedText := `&lt;p&gt;Hello&lt;/p&gt;`
	err := fmt.Errorf("%s", inputText)
	safeError := SafeError(err)
	if safeError.Error() != expectedText {
		t.Fatalf("error: html escape not applied to '%v'", safeError.Error())
	}
}

func TestGetStableTxt(t *testing.T) {
	version, err := GetStableTxt()
	if err != nil {
		t.Fatalf("error reading stable.txt: %v", err)
	}
	if version == "" {
		t.Fatalf("error: version is empty")
	}
	re, err := regexp.Compile(`^v[0-9]\.[0-9]{1,2}\.[0-9]{1,2}$`)
	if err != nil {
		t.Fatalf("error compiling regexp: %v", err)
	}
	if !re.Match([]byte(version)) {
		t.Fatalf("error: version (%v) doesn't match regexp", version)
	}
	if err := os.Setenv("KO_DATA_PATH", "./../../kodat"); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	_, err = GetStableTxt()
	if err == nil {
		t.Fatalf("error expected to not find stable.txt")
	}
}

func TestGetDataPath(t *testing.T) {
	type testCase struct {
		Name          string
		EnvKoDataPath *string
		ExpectedValue string
	}

	for _, tc := range []testCase{
		{
			Name:          "passing has env set",
			EnvKoDataPath: Pointer("a"),
			ExpectedValue: "a",
		},
		{
			Name:          "defaulting no value set",
			EnvKoDataPath: nil,
			ExpectedValue: "kodata",
		},
	} {
		checkpointValue := os.Getenv("KO_DATA_PATH")

		if tc.EnvKoDataPath != nil {
			t.Logf("testcase (%v) setting env to '%v'...", tc.Name, *tc.EnvKoDataPath)
			if err := os.Setenv("KO_DATA_PATH", *tc.EnvKoDataPath); err != nil {
				t.Errorf("error restoring env value: %v\n", err)
			}
		} else {
			// nolint:errcheck
			os.Unsetenv("KO_DATA_PATH")
		}

		val := GetDataPath()
		if !reflect.DeepEqual(val, tc.ExpectedValue) {
			t.Fatalf("error: testcase (%v) datapath value (%v) doesn't equal expected value (%v)", tc.Name, val, tc.ExpectedValue)
		}

		if err := os.Setenv("KO_DATA_PATH", checkpointValue); err != nil {
			t.Errorf("error restoring env value: %v\n", err)
		}
	}
}
