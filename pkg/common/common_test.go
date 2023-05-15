package common

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"regexp"
	"testing"
)

func init() {
	if err := os.Setenv("KO_DATA_PATH", "./../../kodata"); err != nil {
		log.Fatalf("failed to set env: %v", err)
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
	err := fmt.Errorf(inputText)
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
