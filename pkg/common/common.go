package common

import (
	"fmt"
	"html"
	"os"
	"path"
	"strings"
)

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

func GetStableTxt() (string, error) {
	content, err := ReadFile(path.Join(os.Getenv("KO_DATA_PATH"), "metadata", "stable.txt"))
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(content, "\n"), nil
}
