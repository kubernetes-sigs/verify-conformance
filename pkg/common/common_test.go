package common

import (
	_ "embed"
	"fmt"
	"testing"
)

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
	err := fmt.Errorf(inputText)
	safeError := SafeError(err)
	if safeError.Error() != expectedText {
		t.Fatalf("error: html escape not applied to '%v'", safeError.Error())
	}
}
