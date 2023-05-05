package common

import (
	"fmt"
	"html"
	"os"
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
