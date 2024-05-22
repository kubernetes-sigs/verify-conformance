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
