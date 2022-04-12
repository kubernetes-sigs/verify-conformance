package common

import (
	"fmt"
	"html"
	"io/ioutil"
)

func ReadFile(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func SafeError(input error) (output error) {
	return fmt.Errorf(html.EscapeString(input.Error()))
}
