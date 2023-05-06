package suite

import (
	_ "embed"
	"log"
	"os"
	"path"
	"testing"

	githubql "github.com/shurcooL/githubv4"
)

var (
	//go:embed testdata/TestGetJunitSubmittedConformanceTests-coolkube-v1-27-junit_01.xml
	testGetJunitSubmittedConformanceTestsCoolkubeV127Junit_01xml string
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
				Title:  githubql.String("Conformance results for SOMETHING/v1.27"),
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
	prSuite := NewPRSuite(&PullRequest{
		PullRequestQuery: PullRequestQuery{
			Title: githubql.String("Conformance results for coolkube/v1.27"),
		},
	})
	if err := prSuite.thePRTitleIsNotEmpty(); err != nil {
		t.Fatalf("error: %v", err)
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
					Title: githubql.String("Conformance results for v1.27/coolkube"),
				},
				SupportingFiles: []*PullRequestFile{
					{
						Name:     "v1.27/coolkube/README.md",
						BaseName: "README.md",
					},
					{
						Name:     "v1.27/coolkube/PRODUCT.yaml",
						BaseName: "PRODUCT.yaml",
					},
					{
						Name:     "v1.27/coolkube/e2e.log",
						BaseName: "e2e.log",
					},
					{
						Name:     "v1.27/coolkube/junit_01.xml",
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
					Title: githubql.String("Conformance results for v1.27/badkube"),
				},
				SupportingFiles: []*PullRequestFile{
					{
						Name:     "v1.27/badkube/README.md",
						BaseName: "README.md",
					},
					{
						Name:     "v1.27/coolkube/PRODUCT.yaml",
						BaseName: "PRODUCT.yaml",
					},
					{
						Name:     "v1.27/coolkube/junit_01.xml",
						BaseName: "junit_01.xml",
					},
					{
						Name:     "v1.27/coolkube/main.go",
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

}

func TestThereIsOnlyOnePathOfFolders(t *testing.T) {

}

func TestTheTitleOfThePR(t *testing.T) {

}

func TestTheTitleOfThePRMatches(t *testing.T) {

}

func TestTheFilesInThePR(t *testing.T) {

}

func TestAFile(t *testing.T) {
	prSuite := NewPRSuite(&PullRequest{
		PullRequestQuery: PullRequestQuery{
			Title: githubql.String("Conformance results for v1.27/coolkube"),
		},
		SupportingFiles: []*PullRequestFile{
			{
				Name:     "v1.27/coolkube/junit_01.xml",
				BaseName: "junit_01.xml",
				Contents: testGetJunitSubmittedConformanceTestsCoolkubeV127Junit_01xml,
			},
		},
	})
	if err := prSuite.aFile("junit_01.xml"); err != nil {
		t.Fatalf("error: %v", err)
	}
}

func TestGetFileByFileName(t *testing.T) {
	prSuite := NewPRSuite(&PullRequest{
		PullRequestQuery: PullRequestQuery{
			Title: githubql.String("Conformance results for v1.27/coolkube"),
		},
		SupportingFiles: []*PullRequestFile{
			{
				Name:     "v1.27/coolkube/junit_01.xml",
				BaseName: "junit_01.xml",
				Contents: testGetJunitSubmittedConformanceTestsCoolkubeV127Junit_01xml,
			},
		},
	})
	if file := prSuite.GetFileByFileName("junit_01.xml"); file == nil {
		t.Fatalf("error: file 'junit_01.xml' is empty and should not be")
	}
}

func TestTheYamlFileContainsTheRequiredAndNonEmptyField(t *testing.T) {

}

func TestIsNotEmpty(t *testing.T) {

}

func TestALineOfTheFileMatches(t *testing.T) {

}

func TestAListOfCommits(t *testing.T) {

}

func TestThereIsOnlyOneCommit(t *testing.T) {

}

func TestThatVersionMatchesTheSameKubernetesReleaseVersionAsInTheFolderStructure(t *testing.T) {

}

func TestAListOfLabelsInThePR(t *testing.T) {

}

func TestTheLabelPrefixedWithAndEndingWithKubernetesReleaseVersionShouldBePresent(t *testing.T) {

}

func TestTheContentOfTheInTheValueOfIsAValid(t *testing.T) {

}

func TestTheContentOfTheUrlInTheValueOfMatches(t *testing.T) {

}

func TestSetSubmissionMetadatafromFolderStructure(t *testing.T) {

}

func TestTheReleaseVersionMatchesTheReleaseVersionInTheTitle(t *testing.T) {

}

func TestTheReleaseVersion(t *testing.T) {

}

func TestItIsAValidAndSupportedRelease(t *testing.T) {

}

func TestGetRequiredTests(t *testing.T) {

}

func TestGetMissingJunitTestsFromPRSuite(t *testing.T) {

}

func TestDetermineSuccessfulTests(t *testing.T) {
	prSuite := NewPRSuite(&PullRequest{
		PullRequestQuery: PullRequestQuery{
			Title: githubql.String("Conformance results for v1.27/coolkube"),
		},
		SupportingFiles: []*PullRequestFile{
			{
				Name:     "v1.27/coolkube/junit_01.xml",
				BaseName: "junit_01.xml",
				Contents: testGetJunitSubmittedConformanceTestsCoolkubeV127Junit_01xml,
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
			Title: githubql.String("Conformance results for v1.27/coolkube"),
		},
		SupportingFiles: []*PullRequestFile{
			{
				Name:     "v1.27/coolkube/junit_01.xml",
				BaseName: "junit_01.xml",
				Contents: testGetJunitSubmittedConformanceTestsCoolkubeV127Junit_01xml,
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
	prSuite := NewPRSuite(&PullRequest{
		PullRequestQuery: PullRequestQuery{
			Title: githubql.String("Conformance results for v1.27/coolkube"),
		},
		SupportingFiles: []*PullRequestFile{
			{
				Name:     "v1.27/coolkube/junit_01.xml",
				BaseName: "junit_01.xml",
				Contents: testGetJunitSubmittedConformanceTestsCoolkubeV127Junit_01xml,
			},
		},
	})
	tests, err := prSuite.getJunitSubmittedConformanceTests()
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(tests) < 1 {
		t.Fatal("error: no tests found")
	}
}

func TestTheTestsPassAndAreSuccessful(t *testing.T) {

}

func TestAllRequiredTestsInArePresent(t *testing.T) {

}

func TestIsValidYaml(t *testing.T) {

}

func TestIsValid(t *testing.T) {

}

func TestAPRTitle(t *testing.T) {

}

func TestGetLabelsAndCommentsFromSuiteResultsBuffer(t *testing.T) {

}
