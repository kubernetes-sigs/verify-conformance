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
	"fmt"
	"html"
	"os"
	"path"
	"strings"
)

var (
	DataPathPrefix = ""
)

func Pointer[V any](input V) *V {
	return &input
}

func ReadFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func SafeError(input error) (output error) {
	return fmt.Errorf(html.EscapeString(input.Error()))
}

func GetDataPath() string {
	if val, ok := os.LookupEnv("KO_DATA_PATH"); ok {
		return val
	}
	return path.Join(DataPathPrefix, "./kodata")
}

func GetStableTxt() (string, error) {
	content, err := ReadFile(path.Join(GetDataPath(), "metadata", "stable.txt"))
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(content, "\n"), nil
}
