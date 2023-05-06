package plugin

import (
	"os"
	"regexp"
	"testing"
)

func TestGetStableTxt(t *testing.T) {
	if err := os.Setenv("KO_DATA_PATH", "./../../kodata"); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
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
}
