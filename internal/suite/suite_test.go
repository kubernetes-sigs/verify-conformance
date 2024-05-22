package suite

import (
	"bytes"
	_ "embed"
	"log"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"

	githubql "github.com/shurcooL/githubv4"
	"sigs.k8s.io/verify-conformance/internal/common"
)

// TODO add Gomega https://onsi.github.io/gomega/

var (
	//go:embed testdata/TestGetJunitSubmittedConformanceTests-coolkube-v1-30-junit_01.xml
	testGetJunitSubmittedConformanceTestsCoolkubeV130Junit_01xml string
	//go:embed testdata/TestGetJunitSubmittedConformanceTests-coolkube-v1-30-junit_01-with-1-test-failed.xml
	testGetJunitSubmittedConformanceTestsCoolkubeV130Junit_01WithOneTestFailedxml string
	//go:embed testdata/TestGetJunitSubmittedConformanceTests-coolkube-v1-30-junit_01-with-1-test-missing.xml
	testGetJunitSubmittedConformanceTestsCoolkubeV130Junit_01WithOneTestMissingxml string
	//go:embed testdata/TestGetJunitSubmittedConformanceTests-coolkube-v1-30-junit_01-with-1-extra-test.xml
	testGetJunitSubmittedConformanceTestsCoolkubeV130Junit_01xmlWithOneExtraTest string
)

func init() {
	if err := os.Setenv("KO_DATA_PATH", "./../../kodata"); err != nil {
		log.Fatalf("failed to set env: %v", err)
	}
}

func TestNewPRSuite(t *testing.T) {
	for _, pr := range []*PullRequest{
		{
			PullRequestQuery: PullRequestQuery{
				Number: githubql.Int(1),
				Title:  githubql.String("Conformance results for SOMETHING/v1.30"),
				Author: struct{ Login githubql.String }{
					Login: githubql.String("BobyMCbobs"),
				},
				Repository: struct {
					Name  githubql.String
					Owner struct{ Login githubql.String }
				}{
					Name: githubql.String("k8s-conformance"),
					Owner: struct{ Login githubql.String }{
						Login: githubql.String("cncf"),
					},
				},
				Labels: struct {
					Nodes []struct{ Name githubql.String }
				}{
					Nodes: []struct{ Name githubql.String }{},
				},
				Files: struct {
					Nodes []struct{ Path githubql.String }
				}{
					Nodes: []struct{ Path githubql.String }{
						{
							Path: githubql.String("e2e.log"),
						},
						{
							Path: githubql.String("junit_01.xml"),
						},
						{
							Path: githubql.String("README.md"),
						},
						{
							Path: githubql.String("PRODUCT.yaml"),
						},
					},
				},
				Commits: struct {
					Nodes []struct {
						Commit struct {
							Oid    githubql.String
							Status struct {
								Contexts []struct {
									Context githubql.String
									State   githubql.String
								}
							}
						}
					}
				}{
					Nodes: []struct {
						Commit struct {
							Oid    githubql.String
							Status struct {
								Contexts []struct {
									Context githubql.String
									State   githubql.String
								}
							}
						}
					}{
						{
							Commit: struct {
								Oid    githubql.String
								Status struct {
									Contexts []struct {
										Context githubql.String
										State   githubql.String
									}
								}
							}{
								Oid: githubql.String(""),
								Status: struct {
									Contexts []struct {
										Context githubql.String
										State   githubql.String
									}
								}{
									Contexts: []struct {
										Context githubql.String
										State   githubql.String
									}{
										{
											Context: githubql.String(""),
											State:   githubql.String(""),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	} {
		prSuite := NewPRSuite(pr)
		if prSuite.PR == nil {
			t.Fatalf("error: PR is empty")
		}
		if len(prSuite.Labels) != 1 {
			t.Fatalf("error: PR must start with one label (%v)", prSuite.Labels)
		}
		if prSuite.MetadataFolder != path.Join(".", "..", "..", "kodata", "conformance-testdata") {
			t.Fatalf("error: metadata folder not as expected (%v)", prSuite.MetadataFolder)
		}
		if prSuite.buffer.Len() != 0 {
			t.Fatalf("error: buffer is not nil")
		}
	}
}

func TestNewTestSuite(t *testing.T) {
	prSuite := NewPRSuite(&PullRequest{})
	testSuite := prSuite.NewTestSuite(PRSuiteOptions{})
	if testSuite.Name != "how-are-the-prs" {
		t.Fatalf("unexpected test suite name: %v", testSuite.Name)
	}
}

func TestSetMetadataFolder(t *testing.T) {
	newMetadataFolder := "abc/123/cool/test/path"
	prSuite := NewPRSuite(&PullRequest{})
	prSuiteCopy := &PRSuite{}
	*prSuiteCopy = *prSuite
	prSuiteCopy = prSuiteCopy.SetMetadataFolder(newMetadataFolder)
	if prSuite.MetadataFolder == prSuiteCopy.MetadataFolder {
		t.Fatalf("error: metadata folder not changed and matches original (%v)", prSuite.MetadataFolder)
	}
	if prSuiteCopy.MetadataFolder != newMetadataFolder {
		t.Fatalf("error: metadata folder not set to %v", newMetadataFolder)
	}
}

func TestThePRTitleIsNotEmpty(t *testing.T) {
	type testCase struct {
		Name                string
		PullRequest         *PullRequest
		ExpectedErrorString string
	}

	for _, tc := range []testCase{
		{
			PullRequest: &PullRequest{
				PullRequestQuery: PullRequestQuery{
					Title: githubql.String("Conformance results for coolkube/v1.30"),
				},
			},
		},
		{
			PullRequest:         &PullRequest{},
			ExpectedErrorString: "title is empty",
		},
	} {
		prSuite := NewPRSuite(tc.PullRequest)
		if err := prSuite.thePRTitleIsNotEmpty(); err != nil && !strings.Contains(err.Error(), tc.ExpectedErrorString) {
			t.Fatalf("error: %v", err)
		}
	}
}

func TestIsIncludedInItsFileList(t *testing.T) {
	type testCase struct {
		Name         string
		PullRequest  *PullRequest
		MissingFiles []string
		ExtraFiles   []string
	}

	requiredFiles := []string{
		"README.md",
		"PRODUCT.yaml",
		"e2e.log",
		"junit_01.xml",
	}
	nonRelatedFiles := []string{
		"something.sh",
		"recipes.org",
		"index.js",
		"main.go",
	}
	for _, item := range []testCase{
		{
			Name: "contains all correct files and nothing more",
			PullRequest: &PullRequest{
				PullRequestQuery: PullRequestQuery{
					Title: githubql.String("Conformance results for v1.30/coolkube"),
				},
				SupportingFiles: []*PullRequestFile{
					{
						Name:     "v1.30/coolkube/README.md",
						BaseName: "README.md",
					},
					{
						Name:     "v1.30/coolkube/PRODUCT.yaml",
						BaseName: "PRODUCT.yaml",
					},
					{
						Name:     "v1.30/coolkube/e2e.log",
						BaseName: "e2e.log",
					},
					{
						Name:     "v1.30/coolkube/junit_01.xml",
						BaseName: "junit_01.xml",
					},
				},
			},
		},
		{
			Name:         "missing e2e.log and contains main.go",
			MissingFiles: []string{"e2e.log"},
			PullRequest: &PullRequest{
				PullRequestQuery: PullRequestQuery{
					Title: githubql.String("Conformance results for v1.30/badkube"),
				},
				SupportingFiles: []*PullRequestFile{
					{
						Name:     "v1.30/badkube/README.md",
						BaseName: "README.md",
					},
					{
						Name:     "v1.30/coolkube/PRODUCT.yaml",
						BaseName: "PRODUCT.yaml",
					},
					{
						Name:     "v1.30/coolkube/junit_01.xml",
						BaseName: "junit_01.xml",
					},
					{
						Name:     "v1.30/coolkube/main.go",
						BaseName: "main.go",
					},
				},
			},
		},
	} {
		prSuite := NewPRSuite(item.PullRequest)
		filesMissingFromPR := []string{}
		for _, f := range requiredFiles {
			if err := prSuite.isIncludedInItsFileList(f); err != nil {
				filesMissingFromPR = append(filesMissingFromPR, f)
			}
		}
		missingFileCount := 0
		for _, fm := range filesMissingFromPR {
			for _, e := range item.MissingFiles {
				if e == fm {
					missingFileCount++
				}
			}
		}
		if missingFileCount != len(item.MissingFiles) {
			t.Fatalf("error: missing file count (%v) doesn't match expected (%v)", missingFileCount, len(item.MissingFiles))
		}
		filesNonRelatedInPR := []string{}
		for _, f := range nonRelatedFiles {
			if err := prSuite.isIncludedInItsFileList(f); err == nil {
				filesNonRelatedInPR = append(filesNonRelatedInPR, f)
			}
		}
		notRelatedFileCount := 0
		for _, fm := range filesNonRelatedInPR {
			for _, e := range item.ExtraFiles {
				if e == fm {
					notRelatedFileCount++
				}
			}
		}
		if notRelatedFileCount != len(item.ExtraFiles) {
			t.Fatalf("error: notRelated file count (%v) doesn't match expected (%v)", notRelatedFileCount, len(item.ExtraFiles))
		}
	}
}

func TestFileFolderStructureMatchesRegex(t *testing.T) {
	type testSuite struct {
		Name                string
		PullRequest         *PullRequest
		ExpectedErrorString string
	}

	folderStructureRegexp := `(v1.[0-9]{2})/(.*)`

	for _, tc := range []testSuite{
		{
			Name: "valid file paths",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						Name: "v1.30/coolkube/README.md",
					},
					{
						Name: "v1.30/coolkube/PRODUCT.yaml",
					},
					{
						Name: "v1.30/coolkube/junit_01.xml",
					},
					{
						Name: "v1.30/coolkube/e2e.log",
					},
				},
			},
		},
		{
			Name: "invalid file paths with edit outside pr",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						Name: "v1.30/coolkube/README.md",
					},
					{
						Name: "v1.30/coolkube/PRODUCT.yaml",
					},
					{
						Name: "v1.30/coolkube/junit_01.xml",
					},
					{
						Name: "v1.30/coolkube/e2e.log",
					},
					{
						Name: "README.md",
					},
				},
			},
			ExpectedErrorString: "not allowed.",
		},
		{
			Name: "invalid file paths missing distroname",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						Name: "v1.30//README.md",
					},
					{
						Name: "v1.30//PRODUCT.yaml",
					},
					{
						Name: "v1.30//junit_01.xml",
					},
					{
						Name: "v1.30//e2e.log",
					},
				},
			},
			ExpectedErrorString: "not allowed.",
		},
		{
			Name: "invalid file paths ",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						Name: "README.md",
					},
				},
			},
			ExpectedErrorString: "your product submission PR must be in folders structured like",
		},
	} {
		prSuite := NewPRSuite(tc.PullRequest)
		if err := prSuite.fileFolderStructureMatchesRegex(folderStructureRegexp); err != nil && !strings.Contains(err.Error(), tc.ExpectedErrorString) {
			t.Fatalf("error with testcase '%v'; %v", tc.Name, err)
		}
	}
}

func TestThereIsOnlyOnePathOfFolders(t *testing.T) {
	type testSuite struct {
		Name                string
		PullRequest         *PullRequest
		ExpectedErrorString string
	}

	for _, tc := range []testSuite{
		{
			Name: "valid file paths",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						Name: "v1.30/coolkube/README.md",
					},
					{
						Name: "v1.30/coolkube/PRODUCT.yaml",
					},
					{
						Name: "v1.30/coolkube/junit_01.xml",
					},
					{
						Name: "v1.30/coolkube/e2e.log",
					},
				},
			},
		},
		{
			Name: "invalid file paths with edit outside pr",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						Name: "v1.30/coolkube/README.md",
					},
					{
						Name: "v1.30/coolkube/PRODUCT.yaml",
					},
					{
						Name: "v1.30/coolkube/junit_01.xml",
					},
					{
						Name: "v1.30/coolkube/e2e.log",
					},
					{
						Name: "README.md",
					},
				},
			},
			ExpectedErrorString: "there should be a single set of products in the submission",
		},
		{
			Name: "invalid file paths with multiple submissions",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						Name: "v1.30/coolkube/README.md",
					},
					{
						Name: "v1.30/coolkube/PRODUCT.yaml",
					},
					{
						Name: "v1.30/coolkube/junit_01.xml",
					},
					{
						Name: "v1.30/coolkube/e2e.log",
					},
					{
						Name: "v1.30/coolerkube/README.md",
					},
					{
						Name: "v1.30/coolerkube/PRODUCT.yaml",
					},
					{
						Name: "v1.30/coolerkube/junit_01.xml",
					},
					{
						Name: "v1.30/coolerkube/e2e.log",
					},
				},
			},
			ExpectedErrorString: "there should be a single set of products in the submission",
		},
	} {
		prSuite := NewPRSuite(tc.PullRequest)
		if err := prSuite.thereIsOnlyOnePathOfFolders(); err != nil && !strings.Contains(err.Error(), tc.ExpectedErrorString) {
			t.Fatalf("error with testcase '%v'; %v", tc.Name, err)
		}
	}
}

func TestTheTitleOfThePR(t *testing.T) {
	type testSuite struct {
		Name                string
		PullRequest         *PullRequest
		ExpectedErrorString string
	}

	for _, tc := range []testSuite{
		{
			Name: "valid title",
			PullRequest: &PullRequest{
				PullRequestQuery: PullRequestQuery{
					Title: githubql.String("Conformance results for v1.30/coolkube"),
				},
			},
		},
		{
			Name: "invalid empty title",
			PullRequest: &PullRequest{
				PullRequestQuery: PullRequestQuery{},
			},
			ExpectedErrorString: "title is empty",
		},
	} {
		prSuite := NewPRSuite(tc.PullRequest)
		if err := prSuite.theTitleOfThePR(); err != nil && !strings.Contains(err.Error(), tc.ExpectedErrorString) {
			t.Fatalf("error on testcase '%v'; %v", tc.Name, err)
		}
	}
}

func TestTheTitleOfThePRMatches(t *testing.T) {
	type testSuite struct {
		Name                string
		PullRequest         *PullRequest
		ExpectedErrorString string
	}
	titleRegexp := `(.*) (v1.[0-9]{2})[ /](.*)`

	for _, tc := range []testSuite{
		{
			Name: "valid title",
			PullRequest: &PullRequest{
				PullRequestQuery: PullRequestQuery{
					Title: githubql.String("Conformance results for v1.30/coolkube"),
				},
			},
		},
		{
			Name: "invalid title without period in version",
			PullRequest: &PullRequest{
				PullRequestQuery: PullRequestQuery{
					Title: githubql.String("Conformance results for v130/coolkube"),
				},
			},
			ExpectedErrorString: "title must be formatted like",
		},
		{
			Name: "invalid title with non-conformant text",
			PullRequest: &PullRequest{
				PullRequestQuery: PullRequestQuery{
					Title: githubql.String("test test test test aaaand fail"),
				},
			},
			ExpectedErrorString: "title must be formatted like",
		},
	} {
		prSuite := NewPRSuite(tc.PullRequest)
		if err := prSuite.theTitleOfThePRMatches(titleRegexp); err != nil && !strings.Contains(err.Error(), tc.ExpectedErrorString) {
			t.Fatalf("error on testcase '%v'; %v", tc.Name, err)
		}
	}
}

func TestTheFilesInThePR(t *testing.T) {
	type testSuite struct {
		Name                string
		PullRequest         *PullRequest
		ExpectedErrorString string
	}

	for _, tc := range []testSuite{
		{
			Name: "valid with files",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						BaseName: "README.md",
						Contents: "# Hi!",
					},
				},
			},
		},
		{
			Name:                "invalid without files",
			PullRequest:         &PullRequest{},
			ExpectedErrorString: "there were no files found in the submission",
		},
	} {
		prSuite := NewPRSuite(tc.PullRequest)
		if err := prSuite.theFilesInThePR(); err != nil && !strings.Contains(err.Error(), tc.ExpectedErrorString) {
			t.Fatalf("error on testcase '%v'; %v", tc.Name, err)
		}
	}
}

func TestTheFilesIncludedInThePRAreOnly(t *testing.T) {
	type testSuite struct {
		Name                string
		PullRequest         *PullRequest
		FilesString         string
		ExpectedErrorString string
	}

	for _, tc := range []testSuite{
		{
			Name: "valid submission",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						BaseName: "README.md",
					},
					{
						BaseName: "e2e.log",
					},
					{
						BaseName: "PRODUCT.yaml",
					},
					{
						BaseName: "junit_01.xml",
					},
				},
			},
			FilesString: "README.md, e2e.log, PRODUCT.yaml, junit_01.xml",
		},
		{
			Name: "invalid submission with extra files",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						BaseName: "README.md",
					},
					{
						BaseName: "e2e.log",
					},
					{
						BaseName: "PRODUCT.yaml",
					},
					{
						BaseName: "scenic-photo.png",
					},
					{
						BaseName: "soup-recommendation.ogg",
					},
					{
						BaseName: "caleb-was-here.txt",
					},
				},
			},
			FilesString:         "README.md, e2e.log, PRODUCT.yaml, junit_01.xml",
			ExpectedErrorString: "it appears that there are 3 non-required file(s) included in the submission: scenic-photo.png, soup-recommendation.ogg, caleb-was-here.txt",
		},
	} {
		prSuite := NewPRSuite(tc.PullRequest)
		err := prSuite.theFilesIncludedInThePRAreOnly(tc.FilesString)
		if err != nil && !strings.Contains(err.Error(), tc.ExpectedErrorString) {
			t.Errorf("unexpected error in testcase (%v): %v", tc.Name, err)
		}
		t.Logf("ran testcase (%v)", tc.Name)
	}
}

func TestAFile(t *testing.T) {
	prSuite := NewPRSuite(&PullRequest{
		PullRequestQuery: PullRequestQuery{
			Title: githubql.String("Conformance results for v1.30/coolkube"),
		},
		SupportingFiles: []*PullRequestFile{
			{
				Name:     "v1.30/coolkube/junit_01.xml",
				BaseName: "junit_01.xml",
				Contents: testGetJunitSubmittedConformanceTestsCoolkubeV130Junit_01xml,
			},
		},
	})
	if err := prSuite.aFile("junit_01.xml"); err != nil {
		t.Fatalf("error: %v", err)
	}
	if err := prSuite.aFile("README.md"); err != nil && !strings.Contains(err.Error(), "missing required file") {
		t.Fatalf("error expected missing file 'README.md'; %v", err)
	}
}

func TestGetFileByFileName(t *testing.T) {
	prSuite := NewPRSuite(&PullRequest{
		PullRequestQuery: PullRequestQuery{
			Title: githubql.String("Conformance results for v1.30/coolkube"),
		},
		SupportingFiles: []*PullRequestFile{
			{
				Name:     "v1.30/coolkube/junit_01.xml",
				BaseName: "junit_01.xml",
				Contents: testGetJunitSubmittedConformanceTestsCoolkubeV130Junit_01xml,
			},
		},
	})
	if file := prSuite.GetFileByFileName("junit_01.xml"); file == nil {
		t.Fatalf("error: file 'junit_01.xml' is empty and should not be")
	}
}

func TestTheYamlFileContainsTheRequiredAndNonEmptyField(t *testing.T) {
	type testCase struct {
		Name                string
		PullRequest         *PullRequest
		ExpectedErrorString string
	}
	requiredKeys := []string{"vendor", "name", "version", "type", "description", "website_url", "documentation_url", "contact_email_address"}
	for _, tc := range []testCase{
		{
			Name: "valid PRODUCT.yaml",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						Name:     "v1.30/coolkube/PRODUCT.yaml",
						BaseName: "PRODUCT.yaml",
						Contents: `vendor: "cool"
name: "coolkube"
version: "v1.30"
type: "distribution"
description: "it's just all-round cool and probably the best k8s, idk"
website_url: "https://coolkubernetes.com"
documentation_url: "https://coolkubernetes.com/docs"
contact_email_address: "sales@coolkubernetes.com"`,
					},
				},
			},
		},
		{
			Name: "invalid PRODUCT.yaml missing documentation_url",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						Name:     "v1.30/coolkube/PRODUCT.yaml",
						BaseName: "PRODUCT.yaml",
						Contents: `vendor: "cool"
name: "coolkube"
version: "v1.30"
type: "distribution"
description: "it's just all-round cool and probably the best k8s, idk"
website_url: "https://coolkubernetes.com"
contact_email_address: "sales@coolkubernetes.com"`,
					},
				},
			},
			ExpectedErrorString: "missing or empty field &#39;documentation_url&#39;",
		},
		{
			Name: "invalid PRODUCT.yaml missing many fields",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						Name:     "v1.30/coolkube/PRODUCT.yaml",
						BaseName: "PRODUCT.yaml",
						Contents: `vendor: "cool"
name: "coolkube"
version: "v1.30"
description: "it's just all-round cool and probably the best k8s, idk"
website_url: "https://coolkubernetes.com"
contact_email_address: "sales@coolkubernetes.com"`,
					},
				},
			},
			ExpectedErrorString: "missing or empty field",
		},
		{
			Name:                "missing PRODUCT.yaml",
			ExpectedErrorString: "missing required file",
			PullRequest:         &PullRequest{},
		},
		{
			Name:                "invalid PRODUCT.yaml unable to parse",
			ExpectedErrorString: "unable to read file",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						Name:     "v1.30/coolkube/PRODUCT.yaml",
						BaseName: "PRODUCT.yaml",
						Contents: `v"`,
					},
				},
			},
		},
	} {
		prSuite := NewPRSuite(tc.PullRequest)
	k:
		for _, k := range requiredKeys {
			err := prSuite.theYamlFileContainsTheRequiredAndNonEmptyField("PRODUCT.yaml", k)
			if err != nil && strings.Contains(err.Error(), tc.ExpectedErrorString) {
				continue k
			} else if err != nil {
				t.Fatalf("error: %v", err)
			}
		}
	}
}

func TestIsNotEmpty(t *testing.T) {
	type testCase struct {
		BaseName              string
		FileContents          string
		ExpectedErrorContains string
	}

	for _, tc := range []testCase{
		{
			BaseName:     "A",
			FileContents: "abc123",
		},
		{
			BaseName:              "B",
			FileContents:          "",
			ExpectedErrorContains: "is empty",
		},
		{
			BaseName:              "",
			FileContents:          "",
			ExpectedErrorContains: "unable to find file",
		},
	} {
		pr := &PullRequest{}
		if tc.BaseName != "" {
			pr.SupportingFiles = append(pr.SupportingFiles, &PullRequestFile{
				BaseName: tc.BaseName,
				Contents: tc.FileContents,
			})
		}
		prSuite := NewPRSuite(pr)
		if err := prSuite.isNotEmpty(tc.BaseName); err != nil {
			if err != nil && strings.Contains(err.Error(), tc.ExpectedErrorContains) {
				continue
			} else if err != nil {
				t.Logf("files: %v", prSuite.PR.SupportingFiles[0])
				t.Fatalf("error: with file name '%v'; supporting files: %+v; %v", tc.BaseName, pr.SupportingFiles[0], err)
			}
		}
	}
}

func TestAListOfCommits(t *testing.T) {
	type testCase struct {
		PullRequest         *PullRequest
		ExpectedErrorString string
	}
	for _, tc := range []testCase{
		{
			PullRequest:         &PullRequest{},
			ExpectedErrorString: "no commits were found",
		},
		{
			PullRequest: &PullRequest{
				PullRequestQuery: PullRequestQuery{
					Commits: struct {
						Nodes []struct {
							Commit struct {
								Oid    githubql.String
								Status struct {
									Contexts []struct {
										Context githubql.String
										State   githubql.String
									}
								}
							}
						}
					}{
						Nodes: []struct {
							Commit struct {
								Oid    githubql.String
								Status struct {
									Contexts []struct {
										Context githubql.String
										State   githubql.String
									}
								}
							}
						}{
							{
								Commit: struct {
									Oid    githubql.String
									Status struct {
										Contexts []struct {
											Context githubql.String
											State   githubql.String
										}
									}
								}{
									Oid: githubql.String(""),
									Status: struct {
										Contexts []struct {
											Context githubql.String
											State   githubql.String
										}
									}{
										Contexts: []struct {
											Context githubql.String
											State   githubql.String
										}{
											{
												Context: githubql.String(""),
												State:   githubql.String(""),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	} {
		prSuite := NewPRSuite(tc.PullRequest)
		if err := prSuite.aListOfCommits(); err != nil && err.Error() != tc.ExpectedErrorString {
			t.Fatalf("error unexpected while listing commits: %v", err)
		}
	}
}

func TestThereIsOnlyOneCommit(t *testing.T) {
	type testCase struct {
		PullRequest         *PullRequest
		ExpectedErrorString string
	}
	for _, tc := range []testCase{
		{
			PullRequest: &PullRequest{
				PullRequestQuery: PullRequestQuery{
					Commits: struct {
						Nodes []struct {
							Commit struct {
								Oid    githubql.String
								Status struct {
									Contexts []struct {
										Context githubql.String
										State   githubql.String
									}
								}
							}
						}
					}{
						Nodes: []struct {
							Commit struct {
								Oid    githubql.String
								Status struct {
									Contexts []struct {
										Context githubql.String
										State   githubql.String
									}
								}
							}
						}{
							{
								Commit: struct {
									Oid    githubql.String
									Status struct {
										Contexts []struct {
											Context githubql.String
											State   githubql.String
										}
									}
								}{
									Oid: githubql.String(""),
									Status: struct {
										Contexts []struct {
											Context githubql.String
											State   githubql.String
										}
									}{
										Contexts: []struct {
											Context githubql.String
											State   githubql.String
										}{
											{
												Context: githubql.String(""),
												State:   githubql.String(""),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			PullRequest: &PullRequest{
				PullRequestQuery: PullRequestQuery{
					Commits: struct {
						Nodes []struct {
							Commit struct {
								Oid    githubql.String
								Status struct {
									Contexts []struct {
										Context githubql.String
										State   githubql.String
									}
								}
							}
						}
					}{
						Nodes: []struct {
							Commit struct {
								Oid    githubql.String
								Status struct {
									Contexts []struct {
										Context githubql.String
										State   githubql.String
									}
								}
							}
						}{
							{
								Commit: struct {
									Oid    githubql.String
									Status struct {
										Contexts []struct {
											Context githubql.String
											State   githubql.String
										}
									}
								}{
									Oid: githubql.String(""),
									Status: struct {
										Contexts []struct {
											Context githubql.String
											State   githubql.String
										}
									}{
										Contexts: []struct {
											Context githubql.String
											State   githubql.String
										}{
											{
												Context: githubql.String(""),
												State:   githubql.String(""),
											},
										},
									},
								},
							},
							{
								Commit: struct {
									Oid    githubql.String
									Status struct {
										Contexts []struct {
											Context githubql.String
											State   githubql.String
										}
									}
								}{
									Oid: githubql.String(""),
									Status: struct {
										Contexts []struct {
											Context githubql.String
											State   githubql.String
										}
									}{
										Contexts: []struct {
											Context githubql.String
											State   githubql.String
										}{
											{
												Context: githubql.String(""),
												State:   githubql.String(""),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			ExpectedErrorString: "more than one commit was found; only one commit is allowed.",
		},
	} {
		prSuite := NewPRSuite(tc.PullRequest)
		if err := prSuite.thereIsOnlyOneCommit(); err != nil && err.Error() != tc.ExpectedErrorString {
			t.Fatalf("error unexpected while listing commits: %v", err)
		}
	}
}

func TestAListOfLabelsInThePR(t *testing.T) {
	type testCase struct {
		PullRequest         *PullRequest
		ExpectedErrorString string
	}

	for _, tc := range []testCase{
		{
			PullRequest: &PullRequest{
				Labels: []string{"conformance-product-submission"},
			},
		},
		{
			PullRequest:         &PullRequest{},
			ExpectedErrorString: "there are no labels found",
		},
	} {
		prSuite := NewPRSuite(tc.PullRequest)
		if err := prSuite.aListOfLabelsInThePR(); err != nil && err.Error() != tc.ExpectedErrorString {
			t.Fatalf("error: %v", err)
		}
	}
}

func TestTheLabelPrefixedWithAndEndingWithKubernetesReleaseVersionShouldBePresent(t *testing.T) {
	type testCase struct {
		PullRequest         *PullRequest
		TestLabel           string
		ExpectedErrorString string
	}

	for _, tc := range []testCase{
		{
			PullRequest: &PullRequest{
				Labels: []string{"release-v1.30"},
			},
			TestLabel: "release-",
		},
		{
			PullRequest: &PullRequest{
				Labels: []string{"release-"},
			},
			TestLabel:           "release-",
			ExpectedErrorString: "required label",
		},
		{
			PullRequest: &PullRequest{
				Labels: []string{},
			},
			TestLabel:           "release-",
			ExpectedErrorString: "required label",
		},
	} {
		prSuite := NewPRSuite(tc.PullRequest)
		prSuite.KubernetesReleaseVersion = "v1.30"
		for _, l := range tc.PullRequest.Labels {
			if err := prSuite.theLabelPrefixedWithAndEndingWithKubernetesReleaseVersionShouldBePresent(tc.TestLabel); err != nil && !strings.Contains(err.Error(), tc.ExpectedErrorString) {
				t.Fatalf("error with labels '%v': %v", l, err)
			}
		}
	}
}

func TestTheContentOfTheInTheValueOfIsAValid(t *testing.T) {
	type testCase struct {
		Name                string
		PullRequest         *PullRequest
		Field               string
		FieldType           string
		ExpectedErrorString string
	}
	content := `vendor: "CoolKube"
name: "Kubernetes - The Cool Way"
version: "1.2.3"
website_url: "https://cool.kube"
repo_url: "https://cool.kube"
documentation_url: "https://docs-for.coo.kube"
product_logo_url: "http://localhost:8081/logo.svg"
type: "installer"
description: "it's just cool OK"
contact_email_address: "greetings@cool.kube"`

	for _, tc := range []testCase{
		{
			Field: "name",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						BaseName: "PRODUCT.yaml",
						Contents: content,
					},
				},
			},
		},
		{
			Field:     "website_url",
			FieldType: "URL",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						BaseName: "PRODUCT.yaml",
						Contents: content,
					},
				},
			},
		},
		{
			Field:     "contact_email_address",
			FieldType: "email",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						BaseName: "PRODUCT.yaml",
						Contents: content,
					},
				},
			},
		},
		{
			Name:                "invalid missing file PRODUCT.yaml",
			Field:               "contact_email_address",
			FieldType:           "email",
			PullRequest:         &PullRequest{},
			ExpectedErrorString: "missing required file",
		},
		{
			Name:      "invalid unable to parse PRODUCT.yaml",
			Field:     "contact_email_address",
			FieldType: "email",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						BaseName: "PRODUCT.yaml",
						Contents: `v"`,
					},
				},
			},
			ExpectedErrorString: "unable to read file",
		},
		{
			Field:     "a",
			FieldType: "text-or-something",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						BaseName: "PRODUCT.yaml",
						Contents: `a: ""`,
					},
				},
			},
			ExpectedErrorString: "",
		},
		{
			Field:     "site_url",
			FieldType: "URL",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						BaseName: "PRODUCT.yaml",
						Contents: `site_url: a`,
					},
				},
			},
			ExpectedErrorString: "in PRODUCT.yaml is not a valid URL",
		},
		{
			Name:      "invalid empty value for field",
			Field:     "email",
			FieldType: "email",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						BaseName: "PRODUCT.yaml",
						Contents: `email: a`,
					},
				},
			},
			ExpectedErrorString: "in PRODUCT.yaml is not a valid address",
		},
		{
			Name:      "invalid url field type",
			Field:     "repo_url",
			FieldType: "URL",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						BaseName: "PRODUCT.yaml",
						Contents: `repo_url: '?'`,
					},
				},
			},
			ExpectedErrorString: "is not a valid URL",
		},
		{
			Name:      "invalid email field type",
			Field:     "contact_email_address",
			FieldType: "email",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						BaseName: "PRODUCT.yaml",
						Contents: `contact_email_address: '?'`,
					},
				},
			},
		},
	} {
		prSuite := NewPRSuite(tc.PullRequest)
		if err := prSuite.theContentOfTheInTheValueOfIsAValid(tc.FieldType, tc.Field); err != nil && !strings.Contains(err.Error(), tc.ExpectedErrorString) {
			t.Fatalf("error with PRODUCT.yaml content field '%v' (type %v); error: %v", tc.Field, tc.FieldType, err)
		}
	}
}

func TestTheContentOfTheUrlInTheValueOfMatches(t *testing.T) {
	type testCase struct {
		Name                string
		PullRequest         *PullRequest
		Field               string
		FieldType           string
		ExpectedErrorString string
	}
	content := `vendor: "CoolKube"
name: "Kubernetes - The Cool Way"
version: "1.2.3"
website_url: "https://cool.kube"
repo_url: "https://cool.kube"
documentation_url: "https://docs-for.coo.kube"
product_logo_url: "http://localhost:8081/logo.svg"
type: "installer"
description: "it's just cool OK"
contact_email_address: "greetings@cool.kube"`
	productYAMLURLDataTypes := map[string]string{
		"vendor":            "string",
		"name":              "string",
		"version":           "string",
		"type":              "string",
		"description":       "string",
		"website_url":       "text/html",
		"repo_url":          "text/html",
		"documentation_url": "text/html",
		"product_logo_url":  "image/svg",
	}

	for _, tc := range []testCase{
		{
			Field:     "name",
			FieldType: "string",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						BaseName: "PRODUCT.yaml",
						Contents: content,
					},
				},
			},
		},
		{
			Field:     "website_url",
			FieldType: "URL",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						BaseName: "PRODUCT.yaml",
						Contents: content,
					},
				},
			},
		},
		{
			Field:     "contact_email_address",
			FieldType: "string",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						BaseName: "PRODUCT.yaml",
						Contents: content,
					},
				},
			},
		},
		{
			Field:     "contact_email_address",
			FieldType: "thing1",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						BaseName: "PRODUCT.yaml",
						Contents: content,
					},
				},
			},
			ExpectedErrorString: "resolving content type",
		},
	} {
		prSuite := NewPRSuite(tc.PullRequest)
		prSuite.PR.ProductYAMLURLDataTypes = productYAMLURLDataTypes
		if err := prSuite.theContentOfTheUrlInTheValueOfMatches(tc.Field, tc.FieldType); err != nil && !strings.Contains(err.Error(), tc.ExpectedErrorString) {
			t.Fatalf("error with PRODUCT.yaml content field '%v' (type %v)", tc.Field, tc.FieldType)
		}
	}
}

func TestTheTypeFieldInPRODUCTyamlIsValid(t *testing.T) {
	type testCase struct {
		Name          string
		Field         string
		Values        string
		PullRequest   *PullRequest
		ExpectedError bool
	}

	for _, tc := range []testCase{
		{
			Name:   "valid installer",
			Field:  "type",
			Values: "installer, distribution, hosted platform",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						BaseName: "PRODUCT.yaml",
						Contents: `---
type: "installer"
`,
					},
				},
			},
		},
		{
			Name:   "valid distribution",
			Field:  "type",
			Values: "installer, distribution, hosted platform",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						BaseName: "PRODUCT.yaml",
						Contents: `---
type: "distribution"
`,
					},
				},
			},
		},
		{
			Name:   "valid hosted platform",
			Field:  "type",
			Values: "installer, distribution, hosted platform",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						BaseName: "PRODUCT.yaml",
						Contents: `---
type: "hosted platform"
`,
					},
				},
			},
		},
		{
			Name:   "invalid type",
			Field:  "type",
			Values: "installer, distribution, hosted platform",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						BaseName: "PRODUCT.yaml",
						Contents: `---
type: "soup"
`,
					},
				},
			},
			ExpectedError: true,
		},
		{
			Name:   "field not found",
			Field:  "type",
			Values: "installer, distribution, hosted platform",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						BaseName: "PRODUCT.yaml",
						Contents: `---
soup: "minestrone"
`,
					},
				},
			},
			ExpectedError: true,
		},
		{
			Name:   "bad yaml",
			Field:  "type",
			Values: "installer, distribution, hosted platform",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						BaseName: "PRODUCT.yaml",
						Contents: `%^(%^&(%&^%))`,
					},
				},
			},
			ExpectedError: true,
		},
		{
			Name:   "file not found",
			Field:  "type",
			Values: "installer, distribution, hosted platform",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{},
			},
			ExpectedError: true,
		},
	} {
		tc := tc
		prSuite := NewPRSuite(tc.PullRequest)
		if err := prSuite.theFieldMatchesOneOfTheFollowingValues(tc.Field, tc.Values); err != nil && (err != nil) != tc.ExpectedError {
			t.Fatalf("error with PRODUCT.yaml field type is invalid: %v", err)
		}
	}
}

func TestSetSubmissionMetadatafromFolderStructure(t *testing.T) {
	type testCase struct {
		PullRequest    *PullRequest
		ExpectedResult bool
	}
	for _, tc := range []testCase{
		{
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						Name: "v1.30/coolkube/junit_01.xml",
					},
					{
						Name: "v1.30/coolkube/README.md",
					},
				},
			},
			ExpectedResult: true,
		},
		{
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{},
			},
			ExpectedResult: false,
		},
	} {
		prSuite := NewPRSuite(tc.PullRequest)
		prSuite.SetSubmissionMetadatafromFolderStructure()
		if (prSuite.KubernetesReleaseVersion == "" || prSuite.ProductName == "") && tc.ExpectedResult {
			t.Fatalf("error unexpected result of metadata being set (%v) intended case being (%v)", prSuite.ProductName, tc.ExpectedResult)
		}
	}
}

func TestTheReleaseVersionMatchesTheReleaseVersionInTheTitle(t *testing.T) {
	type testCase struct {
		PullRequest         *PullRequest
		ExpectedErrorString string
	}

	for _, tc := range []testCase{
		{
			PullRequest: &PullRequest{
				PullRequestQuery: PullRequestQuery{
					Title: githubql.String("conformance results for v1.30/coolkube"),
				},
			},
		},
		{
			PullRequest: &PullRequest{
				PullRequestQuery: PullRequestQuery{
					Title: githubql.String("I WANT CONFORMANCE AND I WANT IT NOW"),
				},
			},
			ExpectedErrorString: "the Kubernetes release version in the title",
		},
	} {
		prSuite := NewPRSuite(tc.PullRequest)
		prSuite.KubernetesReleaseVersion = "v1.30"
		if err := prSuite.theReleaseVersionMatchesTheReleaseVersionInTheTitle(); err != nil && !strings.Contains(err.Error(), tc.ExpectedErrorString) {
			t.Fatalf("error unexpected error matching the release version in the title: %v", err)
		}
	}
}

func TestTheReleaseVersion(t *testing.T) {
	type testCase struct {
		Version             string
		ExpectedErrorString string
	}

	for _, tc := range []testCase{
		{
			Version: "v1.30",
		},
		{
			Version:             "a",
			ExpectedErrorString: "unable to find a Kubernetes release version in the title",
		},
		{
			Version:             "",
			ExpectedErrorString: "unable to find a Kubernetes release version in the title",
		},
	} {
		prSuite := NewPRSuite(&PullRequest{})
		prSuite.KubernetesReleaseVersion = tc.Version
		if err := prSuite.theReleaseVersion(); err != nil && !strings.Contains(err.Error(), tc.ExpectedErrorString) {
			t.Fatalf("error unexpected error finding the release version in the title: %v", err)
		}
	}
}

func TestItIsAValidAndSupportedRelease(t *testing.T) {
	type testCase struct {
		Name                string
		Version             string
		VersionLatest       string
		ExpectedErrorString string
	}

	for _, tc := range []testCase{
		{
			Name:          "valid",
			Version:       "v1.30",
			VersionLatest: "v1.30.0",
		},
		{
			Name:                "invalid unsupported release",
			Version:             "v1.14",
			VersionLatest:       "v1.30.0",
			ExpectedErrorString: "unable to use version",
		},
		{
			Name:                "invalid future release",
			Version:             "v1.208",
			VersionLatest:       "v1.30.0",
			ExpectedErrorString: "unable to use version",
		},
		{
			Name:                "invalid version latest string",
			Version:             "v1.30",
			VersionLatest:       "????",
			ExpectedErrorString: "unable to parse latest release version",
		},
		{
			Name:                "invalid version string",
			Version:             "????",
			VersionLatest:       "v1.30.0",
			ExpectedErrorString: "unable to parse release version",
		},
	} {
		prSuite := NewPRSuite(&PullRequest{})
		prSuite.KubernetesReleaseVersion = tc.Version
		prSuite.KubernetesReleaseVersionLatest = tc.VersionLatest
		if err := prSuite.itIsAValidAndSupportedRelease(); err != nil && !strings.Contains(err.Error(), tc.ExpectedErrorString) {
			t.Fatalf("error: unexpected error in test case '%v': %v", tc.Name, err)
		}
	}
}

func TestGetRequiredTests(t *testing.T) {
	type testCase struct {
		Name                string
		Version             string
		ExpectedTestsCount  int
		ExpectedErrorString string
		MetadataFolder      *string
	}

	for _, tc := range []testCase{
		{
			Name:               "valid",
			Version:            "v1.30",
			ExpectedTestsCount: 402, // NOTE magic number is count of tests form conformance.yaml
		},
		{
			Name:               "valid alternate version",
			Version:            "v1.29",
			ExpectedTestsCount: 388, // NOTE magic number is count of tests form conformance.yaml
		},
		{
			Name:               "valid with test with version above pr version",
			Version:            "v1.31",
			MetadataFolder:     common.Pointer("testdata/metadata/version-of-test-higher"),
			ExpectedTestsCount: 0,
		},
		{
			Name:                "invalid with malformed version",
			Version:             "v1.notfound",
			ExpectedTestsCount:  0,
			ExpectedErrorString: "Malformed version",
		},
		{
			Name:                "invalid unable to parse conformance.yaml",
			Version:             "v1.123",
			ExpectedTestsCount:  0,
			MetadataFolder:      common.Pointer("testdata/metadata/unable-to-parse"),
			ExpectedErrorString: "cannot unmarshal string into Go value of type []suite.ConformanceTestMetadata",
		},
		{
			Name:                "invalid version in conformance.yaml test",
			Version:             "v1.27",
			ExpectedTestsCount:  0,
			MetadataFolder:      common.Pointer("testdata/metadata/bad-version-in-conformance.yaml"),
			ExpectedErrorString: "Malformed version",
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			prSuite := NewPRSuite(&PullRequest{})
			prSuite.KubernetesReleaseVersion = tc.Version
			if tc.MetadataFolder != nil {
				prSuite.MetadataFolder = *tc.MetadataFolder
			}
			tests, err := prSuite.GetRequiredTests()
			if err != nil && !strings.Contains(err.Error(), tc.ExpectedErrorString) {
				t.Fatalf("error: unexpected error in test case '%v': %v", tc.Name, err)
			}
			if len(tests) != tc.ExpectedTestsCount {
				t.Fatalf("error: unexpected test count for version %v in test case '%v' is expected to be at %v but instead found at %v", tc.Version, tc.Name, tc.ExpectedTestsCount, len(tests))
			}
		})
	}
}

func TestGetMissingJunitTestsFromPRSuite(t *testing.T) {
	type testCase struct {
		Name                      string
		Version                   string
		MetadataFolder            *string
		PullRequest               *PullRequest
		ExpectedTestsMissingCount int
		ExpectedErrorString       string
	}

	for _, tc := range []testCase{
		{
			Name:    `valid junit`,
			Version: "v1.30",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						Name:     "v1.30/coolkube/junit_01.xml",
						BaseName: "junit_01.xml",
						Contents: testGetJunitSubmittedConformanceTestsCoolkubeV130Junit_01xml,
					},
				},
			},
			ExpectedTestsMissingCount: 0,
		},
		{
			Name:    "valid junit but with one extra test",
			Version: "v1.30",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						Name:     "v1.30/coolkube/junit_01.xml",
						BaseName: "junit_01.xml",
						Contents: testGetJunitSubmittedConformanceTestsCoolkubeV130Junit_01xmlWithOneExtraTest,
					},
				},
			},
			ExpectedTestsMissingCount: 0,
		},
		{
			Name:    "invalid with a metadata folder pointing to nowhere",
			Version: "v1.30",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						Name:     "v1.30/coolkube/junit_01.xml",
						BaseName: "junit_01.xml",
						Contents: testGetJunitSubmittedConformanceTestsCoolkubeV130Junit_01xml,
					},
				},
			},
			ExpectedTestsMissingCount: 0,
			MetadataFolder:            common.Pointer("nowhere"),
			ExpectedErrorString:       "open nowhere/v1.30/conformance.yaml",
		},
		{
			Name:    `empty junit`,
			Version: "v1.30",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						Name:     "v1.30/coolkube/junit_01.xml",
						BaseName: "junit_01.xml",
						Contents: ``,
					},
				},
			},
			ExpectedTestsMissingCount: 0, // skip since invalid junit anyways
			ExpectedErrorString:       "unable to parse junit_01.xml file",
		},
	} {
		prSuite := NewPRSuite(tc.PullRequest)
		prSuite.KubernetesReleaseVersion = tc.Version
		if tc.MetadataFolder != nil {
			prSuite.MetadataFolder = *tc.MetadataFolder
		}
		tests, err := prSuite.GetMissingJunitTestsFromPRSuite()
		if err != nil && !strings.Contains(err.Error(), tc.ExpectedErrorString) {
			t.Fatalf("error: unexpected error in test case '%v': %v", tc.Name, err)
		}
		if len(tests) != tc.ExpectedTestsMissingCount {
			t.Fatalf("error: missing test count for version %v is expected to be at %v but instead found at %v", tc.Version, tc.ExpectedTestsMissingCount, len(tests))
		}
	}
}

func TestDetermineSuccessfulTests(t *testing.T) {
	prSuite := NewPRSuite(&PullRequest{
		PullRequestQuery: PullRequestQuery{
			Title: githubql.String("Conformance results for v1.30/coolkube"),
		},
		SupportingFiles: []*PullRequestFile{
			{
				Name:     "v1.30/coolkube/junit_01.xml",
				BaseName: "junit_01.xml",
				Contents: testGetJunitSubmittedConformanceTestsCoolkubeV130Junit_01xml,
			},
		},
	})
	success, passed, tests, err := prSuite.DetermineSuccessfulTests()
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !success || passed != len(tests) {
		t.Fatalf("error: all tests must be successful")
	}
}

func TestDetermineSuccessfulTestsv125AndAbove(t *testing.T) {
	prSuite := NewPRSuite(&PullRequest{
		PullRequestQuery: PullRequestQuery{
			Title: githubql.String("Conformance results for v1.30/coolkube"),
		},
		SupportingFiles: []*PullRequestFile{
			{
				Name:     "v1.30/coolkube/junit_01.xml",
				BaseName: "junit_01.xml",
				Contents: testGetJunitSubmittedConformanceTestsCoolkubeV130Junit_01xml,
			},
		},
	})
	success, passed, tests, err := prSuite.determineSuccessfulTestsv125AndAbove()
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !success || passed != len(tests) {
		t.Fatalf("error: all tests must be successful")
	}
}

func TestGetJunitSubmittedConformanceTests(t *testing.T) {
	tests, err := NewPRSuite(&PullRequest{
		PullRequestQuery: PullRequestQuery{
			Title: githubql.String("Conformance results for v1.30/coolkube"),
		},
		SupportingFiles: []*PullRequestFile{
			{
				Name:     "v1.30/coolkube/junit_01.xml",
				BaseName: "junit_01.xml",
				Contents: testGetJunitSubmittedConformanceTestsCoolkubeV130Junit_01xml,
			},
		},
	}).getJunitSubmittedConformanceTests()
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(tests) < 1 {
		t.Fatal("error: no tests found")
	}
	_, err = NewPRSuite(&PullRequest{}).getJunitSubmittedConformanceTests()
	if err == nil {
		t.Fatalf("error unexpectedly nil")
	}
	if strings.Contains(err.Error(), "unable to find file junit_01.xml") != true {
		t.Fatalf("error with unexpected content: %v", err)
	}
}

func TestAllRequiredTestsInJunitXmlArePresent(t *testing.T) {
	type testCase struct {
		Name                string
		PullRequest         *PullRequest
		ExpectedErrorString string
		ExpectedLabels      []string
	}

	for _, tc := range []testCase{
		{
			Name: "valid and all tests pass and are successful",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						Name:     "v1.30/coolkube/junit_01.xml",
						BaseName: "junit_01.xml",
						Contents: testGetJunitSubmittedConformanceTestsCoolkubeV130Junit_01xml,
					},
				},
			},
			ExpectedLabels: []string{"conformance-product-submission", "tests-verified-v1.30"},
		},
		{
			Name: "invalid with one test not passing and successful",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						Name:     "v1.30/coolkube/junit_01.xml",
						BaseName: "junit_01.xml",
						Contents: testGetJunitSubmittedConformanceTestsCoolkubeV130Junit_01WithOneTestFailedxml,
					},
				},
			},
			ExpectedLabels:      []string{"conformance-product-submission", "required-tests-missing"},
			ExpectedErrorString: "the following test(s) are missing",
		},
		{
			Name:                "invalid with no junit_01.xml",
			PullRequest:         &PullRequest{},
			ExpectedLabels:      []string{"conformance-product-submission"},
			ExpectedErrorString: "unable to find file junit_01.xml",
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			prSuite := NewPRSuite(tc.PullRequest)
			prSuite.KubernetesReleaseVersion = "v1.30"
			if err := prSuite.allRequiredTestsInJunitXmlArePresent(); err != nil && !strings.Contains(err.Error(), tc.ExpectedErrorString) {
				t.Fatalf("error with testcase '%v'; %v", tc.Name, err)
			}
			foundLabelCount := 0
			for _, l := range tc.ExpectedLabels {
				for _, tcl := range prSuite.Labels {
					if l == tcl {
						foundLabelCount++
					}
				}
			}
			if foundLabelCount != len(tc.ExpectedLabels) {
				t.Fatalf("error: with testcase '%v' did not find all expected labels (%+v) instead found (%+v)", tc.Name, tc.ExpectedLabels, prSuite.Labels)
			}
		})
	}

}

func TestTheTestsPassAndAreSuccessful(t *testing.T) {
	type testCase struct {
		Name                string
		PullRequest         *PullRequest
		ExpectedErrorString string
		ExpectedLabels      []string
	}

	for _, tc := range []testCase{
		{
			Name: "valid and all tests pass and are successful",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						Name:     "v1.30/coolkube/junit_01.xml",
						BaseName: "junit_01.xml",
						Contents: testGetJunitSubmittedConformanceTestsCoolkubeV130Junit_01xml,
					},
				},
			},
			ExpectedLabels: []string{"conformance-product-submission", "no-failed-tests-v1.30"},
		},
		{
			Name: "invalid with one test not passing and successful",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						Name:     "v1.30/coolkube/junit_01.xml",
						BaseName: "junit_01.xml",
						Contents: testGetJunitSubmittedConformanceTestsCoolkubeV130Junit_01WithOneTestFailedxml,
					},
				},
			},
			ExpectedLabels:      []string{"conformance-product-submission", "evidence-missing"},
			ExpectedErrorString: "it appears that there are failures in some tests",
		},
		{
			Name: "invalid with missing junit_01.xml",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{},
			},
			ExpectedLabels:      []string{"conformance-product-submission"},
			ExpectedErrorString: "unable to find file junit_01.xml",
		},
	} {
		prSuite := NewPRSuite(tc.PullRequest)
		prSuite.KubernetesReleaseVersion = "v1.30"
		if err := prSuite.theTestsPassAndAreSuccessful(); err != nil && !strings.Contains(err.Error(), tc.ExpectedErrorString) {
			t.Fatalf("error with testcase '%v'; %v", tc.Name, err)
		}
		foundLabelCount := 0
		for _, l := range tc.ExpectedLabels {
			for _, tcl := range prSuite.Labels {
				if l == tcl {
					foundLabelCount++
				}
			}
		}
		if foundLabelCount != len(tc.ExpectedLabels) {
			t.Fatalf("error: with testcase '%v' did not find all expected labels (%+v) instead found (%+v)", tc.Name, tc.ExpectedLabels, prSuite.Labels)
		}
	}
}

func TestAllRequiredTestsInArePresent(t *testing.T) {
	type testCase struct {
		Name                string
		PullRequest         *PullRequest
		MetadataFolder      *string
		ExpectedErrorString string
	}

	for _, tc := range []testCase{
		{
			Name: "valid and all tests pass and are successful",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						Name:     "v1.30/coolkube/junit_01.xml",
						BaseName: "junit_01.xml",
						Contents: testGetJunitSubmittedConformanceTestsCoolkubeV130Junit_01xml,
					},
				},
			},
		},
		{
			Name: "valid and all tests pass and are successful but with one extra test",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						Name:     "v1.30/coolkube/junit_01.xml",
						BaseName: "junit_01.xml",
						Contents: testGetJunitSubmittedConformanceTestsCoolkubeV130Junit_01xmlWithOneExtraTest,
					},
				},
			},
		},
		{
			Name: "invalid with one test missing",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						Name:     "v1.30/coolkube/junit_01.xml",
						BaseName: "junit_01.xml",
						Contents: testGetJunitSubmittedConformanceTestsCoolkubeV130Junit_01WithOneTestMissingxml,
					},
				},
			},
			ExpectedErrorString: "there appears to be 1 tests missing",
		},
		{
			Name: "invalid with missing junit_01.xml",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{},
			},
			ExpectedErrorString: "unable to find file junit_01.xml",
		},
		{
			Name: "invalid with a metadata folder pointing to nowhere",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						Name:     "v1.30/coolkube/junit_01.xml",
						BaseName: "junit_01.xml",
						Contents: testGetJunitSubmittedConformanceTestsCoolkubeV130Junit_01WithOneTestMissingxml,
					},
				},
			},
			MetadataFolder:      common.Pointer("nowhere"),
			ExpectedErrorString: "open nowhere/v1.30/conformance.yaml",
		},
	} {
		prSuite := NewPRSuite(tc.PullRequest)
		prSuite.KubernetesReleaseVersion = "v1.30"
		if tc.MetadataFolder != nil {
			prSuite.MetadataFolder = *tc.MetadataFolder
		}
		if err := prSuite.allRequiredTestsInArePresent(); err != nil && !strings.Contains(err.Error(), tc.ExpectedErrorString) {
			t.Fatalf("error with testcase '%v'; %v", tc.Name, err)
		}
	}
}

func TestIsValidYaml(t *testing.T) {
	type testCase struct {
		Name                string
		Content             string
		ExpectedErrorString string
	}

	for _, tc := range []testCase{
		{
			Name: "valid yaml",
			Content: `---
a: b
b: c
d: e
`,
		},
		{
			Name:                "invalid yaml 1",
			Content:             `a`,
			ExpectedErrorString: "cannot unmarshal string into Go value of type map[string]interface",
		},
		{
			Name:                "invalid yaml 2",
			Content:             `1`,
			ExpectedErrorString: "cannot unmarshal number into Go value of type map[string]interface",
		},
		{
			Name:                "invalid yaml 3",
			Content:             `:`,
			ExpectedErrorString: "error converting YAML to JSON: yaml: did not find expected key",
		},
	} {
		if err := IsValidYaml([]byte(tc.Content)); err != nil && !strings.Contains(err.Error(), tc.ExpectedErrorString) {
			t.Fatalf("error on test '%v'; %v", tc.Name, err)
		}
	}
}

func TestIsValid(t *testing.T) {
	type testCase struct {
		Name                string
		PullRequest         *PullRequest
		File                string
		FileType            string
		ExpectedErrorString string
	}

	for _, tc := range []testCase{
		{
			Name:     "valid yaml",
			File:     "PRODUCT.yaml",
			FileType: "yaml",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						BaseName: "PRODUCT.yaml",
						Contents: `vendor: "CoolKube"
name: "Kubernetes - The Cool Way"
version: "1.2.3"
website_url: "https://cool.kube"
repo_url: "https://cool.kube"
documentation_url: "https://docs-for.coo.kube"
product_logo_url: "http://localhost:8081/logo.svg"
type: "installer"
description: "it's just cool OK"
contact_email_address: "greetings@cool.kube"`,
					},
				},
			},
		},
		{
			Name:     "invalid yaml",
			File:     "PRODUCT.yaml",
			FileType: "yaml",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						BaseName: "PRODUCT.yaml",
						Contents: `a`,
					},
				},
			},
			ExpectedErrorString: "cannot unmarshal string into Go value of type map[string]interface",
		},
		{
			Name:     "empty yaml",
			File:     "PRODUCT.yaml",
			FileType: "yaml",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						BaseName: "PRODUCT.yaml",
						Contents: ``,
					},
				},
			},
			ExpectedErrorString: "is empty",
		},
		{
			Name:     "valid markdown",
			File:     "README.md",
			FileType: "markdown",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						BaseName: "README.md",
						Contents: `# Hi!`,
					},
				},
			},
		},
		{
			Name:     "missing README.md",
			File:     "README.md",
			FileType: "markdown",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{},
			},
			ExpectedErrorString: "unable to find file",
		},
	} {
		prSuite := NewPRSuite(tc.PullRequest)
		if err := prSuite.IsValid(tc.File, tc.FileType); err != nil && !strings.Contains(err.Error(), tc.ExpectedErrorString) {
			t.Fatalf("error with PRODUCT.yaml content file '%v' (type %v) on test '%v'; %v", tc.File, tc.FileType, tc.Name, err)
		}
	}
}

func TestAPRTitle(t *testing.T) {
	if err := aPRTitle(); err != nil {
		t.Fatalf("error: %v", err)
	}
}

func TestGetLabelsAndCommentsFromSuiteResultsBuffer(t *testing.T) {
	type testCase struct {
		Name                    string
		PullRequest             *PullRequest
		KubernetesVersion       *string
		KubernetesVersionLatest *string
		Buffer                  *bytes.Buffer
		ExpectedComment         *string
		ExpectedLabels          []string
		ExpectedState           *string
		ExpectedErrorString     string
	}

	for _, tc := range []testCase{
		{
			Name: "invalid empty PR",
			PullRequest: &PullRequest{
				PullRequestQuery:        PullRequestQuery{},
				Labels:                  []string{},
				SupportingFiles:         []*PullRequestFile{},
				ProductYAMLURLDataTypes: map[string]string{},
			},
			ExpectedErrorString: "Malformed version",
		},
		{
			Name:              "invalid with KubernetesVersion",
			KubernetesVersion: common.Pointer("v1.30"),
			PullRequest: &PullRequest{
				PullRequestQuery:        PullRequestQuery{},
				Labels:                  []string{},
				SupportingFiles:         []*PullRequestFile{},
				ProductYAMLURLDataTypes: map[string]string{},
			},
		},
		{
			Name:                    "invalid with KubernetesVersion and KubernetesVersionLatest",
			KubernetesVersion:       common.Pointer("v1.30"),
			KubernetesVersionLatest: common.Pointer("v1.30"),
			PullRequest: &PullRequest{
				PullRequestQuery:        PullRequestQuery{},
				Labels:                  []string{},
				SupportingFiles:         []*PullRequestFile{},
				ProductYAMLURLDataTypes: map[string]string{},
			},
			ExpectedLabels: []string{"conformance-product-submission", "missing-file-README.md", "missing-file-PRODUCT.yaml", "missing-file-e2e.log", "missing-file-junit_01.xml", "release-v1.30", "not-verifiable"},
		},
		{
			Name:                    "invalid with non-cuke contents",
			KubernetesVersion:       common.Pointer("v1.30"),
			KubernetesVersionLatest: common.Pointer("v1.30"),
			Buffer:                  bytes.NewBuffer([]byte(`hiiii`)),
			PullRequest: &PullRequest{
				PullRequestQuery:        PullRequestQuery{},
				Labels:                  []string{},
				SupportingFiles:         []*PullRequestFile{},
				ProductYAMLURLDataTypes: map[string]string{},
			},
			ExpectedErrorString: "invalid character 'h' looking for beginning of value",
		},
		{
			Name:                    "invalid with missing supporting conformance.yaml for KubernetesVersion",
			KubernetesVersion:       common.Pointer("v1.123"),
			KubernetesVersionLatest: common.Pointer("v1.123"),
			PullRequest: &PullRequest{
				PullRequestQuery:        PullRequestQuery{},
				Labels:                  []string{},
				SupportingFiles:         []*PullRequestFile{},
				ProductYAMLURLDataTypes: map[string]string{},
			},
			ExpectedLabels:  []string{"conformance-product-submission", "unable-to-process"},
			ExpectedState:   common.Pointer("pending"),
			ExpectedComment: common.Pointer("The release version v1.123 is unable to be processed at this time; Please wait as this version may become available soon."),
		},
		{
			Name:                    "valid pull request",
			KubernetesVersion:       common.Pointer("v1.30"),
			KubernetesVersionLatest: common.Pointer("v1.30"),
			PullRequest: &PullRequest{
				PullRequestQuery: PullRequestQuery{
					Title: githubql.String("Conformance results for v1.30/coolkube"),
					Commits: struct {
						Nodes []struct {
							Commit struct {
								Oid    githubql.String
								Status struct {
									Contexts []struct {
										Context githubql.String
										State   githubql.String
									}
								}
							}
						}
					}{
						Nodes: []struct {
							Commit struct {
								Oid    githubql.String
								Status struct {
									Contexts []struct {
										Context githubql.String
										State   githubql.String
									}
								}
							}
						}{
							{
								Commit: struct {
									Oid    githubql.String
									Status struct {
										Contexts []struct {
											Context githubql.String
											State   githubql.String
										}
									}
								}{
									Oid: githubql.String(""),
									Status: struct {
										Contexts []struct {
											Context githubql.String
											State   githubql.String
										}
									}{
										Contexts: []struct {
											Context githubql.String
											State   githubql.String
										}{
											{
												Context: githubql.String(""),
												State:   githubql.String(""),
											},
										},
									},
								},
							},
						},
					},
				},
				SupportingFiles: []*PullRequestFile{
					{
						Name:     "v1.30/coolkube/PRODUCT.yaml",
						BaseName: "PRODUCT.yaml",
						Contents: `vendor: "cool"
name: "coolkube"
version: "v1.30"
type: "distribution"
description: "it's just all-round cool and probably the best k8s, idk"
website_url: "https://coolkubernetes.com"
documentation_url: "https://coolkubernetes.com/docs"
contact_email_address: "sales@coolkubernetes.com"`,
					},
					{
						Name:     "v1.30/coolkube/README.md",
						BaseName: "README.md",
						Contents: `# v1.30/coolkube`,
					},
					{
						Name:     "v1.30/coolkube/e2e.log",
						BaseName: "e2e.log",
						Contents: `stuff here`,
					},
					{
						Name:     "v1.30/coolkube/junit_01.xml",
						BaseName: "junit_01.xml",
						Contents: testGetJunitSubmittedConformanceTestsCoolkubeV130Junit_01xml,
					},
				},
				ProductYAMLURLDataTypes: map[string]string{},
			},
			ExpectedLabels:  []string{"conformance-product-submission", "tests-verified-v1.30", "no-failed-tests-v1.30", "release-v1.30", "release-documents-checked"},
			ExpectedComment: common.Pointer("All requirements (15) have passed for the submission!\n"),
		},
	} {
		prSuite := NewPRSuite(tc.PullRequest)
		if tc.KubernetesVersion != nil {
			prSuite.KubernetesReleaseVersion = *tc.KubernetesVersion
		}
		if tc.KubernetesVersionLatest != nil {
			prSuite.KubernetesReleaseVersionLatest = *tc.KubernetesVersionLatest
		}
		prSuite.SetSubmissionMetadatafromFolderStructure()
		prSuite.NewTestSuite(PRSuiteOptions{Paths: []string{"../../kodata/features/verify-conformance.feature"}}).Run()
		if tc.Buffer != nil {
			prSuite.buffer = *tc.Buffer
		}
		comment, labels, state, err := prSuite.GetLabelsAndCommentsFromSuiteResultsBuffer()
		if err != nil && !strings.Contains(err.Error(), tc.ExpectedErrorString) {
			t.Fatalf("error: unexpected error string in test case '%v': %v", tc.Name, err)
		}
		if tc.ExpectedComment != nil && comment != *tc.ExpectedComment {
			t.Fatalf("error: comment in test case '%v' result '%v' does not match expected '%v'", tc.Name, comment, *tc.ExpectedComment)
		}
		if len(tc.ExpectedLabels) != 0 && !reflect.DeepEqual(labels, tc.ExpectedLabels) {
			t.Fatalf("error: labels in test case '%v' result '%+v' does not match expected '%+v'", tc.Name, labels, tc.ExpectedLabels)
		}
		if tc.ExpectedState != nil && state != *tc.ExpectedState {
			t.Fatalf("error: state in test case '%v' result '%v' does not match expected '%v'", tc.Name, state, *tc.ExpectedState)
		}
	}
}

func TestInitializeScenario(t *testing.T) {
	prSuite := NewPRSuite(&PullRequest{})
	prSuite.NewTestSuite(PRSuiteOptions{Paths: []string{"../../kodata/features/verify-conformance.feature"}})
	if code := prSuite.Suite.Run(); code != 1 {
		t.Fatalf("error intended failure code of '1', but found to be '%v'", code)
	}
}
