package suite

import (
	_ "embed"
	"log"
	"os"
	"path"
	"strings"
	"testing"

	githubql "github.com/shurcooL/githubv4"
)

// TODO add Gomega https://onsi.github.io/gomega/

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
	type testCase struct {
		Contents              string
		ExpectedErrorContains string
	}
	requiredKeys := []string{"vendor", "name", "version", "type", "description", "website_url", "documentation_url", "contact_email_address"}
	for _, tc := range []testCase{
		{
			Contents: `vendor: "cool"
name: "coolkube"
version: "v1.27"
type: "distribution"
description: "it's just all-round cool and probably the best k8s, idk"
website_url: "https://coolkubernetes.com"
documentation_url: "https://coolkubernetes.com/docs"
contact_email_address: "sales@coolkubernetes.com"`,
		},
		{
			Contents: `vendor: "cool"
name: "coolkube"
version: "v1.27"
type: "distribution"
description: "it's just all-round cool and probably the best k8s, idk"
website_url: "https://coolkubernetes.com"
contact_email_address: "sales@coolkubernetes.com"`,
			ExpectedErrorContains: "missing or empty field &#39;documentation_url&#39;",
		},
		{
			Contents: `vendor: "cool"
name: "coolkube"
version: "v1.27"
description: "it's just all-round cool and probably the best k8s, idk"
website_url: "https://coolkubernetes.com"
contact_email_address: "sales@coolkubernetes.com"`,
			ExpectedErrorContains: "missing or empty field",
		},
	} {
		prSuite := NewPRSuite(&PullRequest{
			SupportingFiles: []*PullRequestFile{
				{
					Name:     "v1.27/coolkube/PRODUCT.yaml",
					BaseName: "PRODUCT.yaml",
					Contents: tc.Contents,
				},
			},
		})
	k:
		for _, k := range requiredKeys {
			err := prSuite.theYamlFileContainsTheRequiredAndNonEmptyField("PRODUCT.yaml", k)
			if err != nil && strings.Contains(err.Error(), tc.ExpectedErrorContains) {
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
				Labels: []string{"release-v1.27"},
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
		prSuite.KubernetesReleaseVersion = "v1.27"
		for _, l := range tc.PullRequest.Labels {
			if err := prSuite.theLabelPrefixedWithAndEndingWithKubernetesReleaseVersionShouldBePresent(tc.TestLabel); err != nil && !strings.Contains(err.Error(), tc.ExpectedErrorString) {
				t.Fatalf("error with labels '%v': %v", l, err)
			}
		}
	}
}

func TestTheContentOfTheInTheValueOfIsAValid(t *testing.T) {
	type testCase struct {
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
	} {
		prSuite := NewPRSuite(tc.PullRequest)
		if err := prSuite.theContentOfTheInTheValueOfIsAValid(tc.FieldType, tc.Field); err != nil && strings.Contains(err.Error(), tc.ExpectedErrorString) {
			t.Fatalf("error with PRODUCT.yaml content field '%v' (type %v)", tc.Field, tc.FieldType)
		}
	}
}

func TestTheContentOfTheUrlInTheValueOfMatches(t *testing.T) {
	type testCase struct {
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
			Field:     "name",
			FieldType: "text",
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
	} {
		prSuite := NewPRSuite(tc.PullRequest)
		if err := prSuite.theContentOfTheUrlInTheValueOfMatches(tc.Field, tc.FieldType); err != nil && strings.Contains(err.Error(), tc.ExpectedErrorString) {
			t.Fatalf("error with PRODUCT.yaml content field '%v' (type %v)", tc.Field, tc.FieldType)
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
						Name: "v1.27/coolkube/junit_01.xml",
					},
					{
						Name: "v1.27/coolkube/README.md",
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
					Title: githubql.String("conformance results for v1.27/coolkube"),
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
		prSuite.KubernetesReleaseVersion = "v1.27"
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
			Version: "v1.27",
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
		Version             string
		VersionLatest       string
		ExpectedErrorString string
	}

	for _, tc := range []testCase{
		{
			Version:       "v1.27",
			VersionLatest: "v1.27.0",
		},
		{
			Version:             "v1.14",
			VersionLatest:       "v1.27.0",
			ExpectedErrorString: "unable to use version",
		},
	} {
		prSuite := NewPRSuite(&PullRequest{})
		prSuite.KubernetesReleaseVersion = tc.Version
		prSuite.KubernetesReleaseVersionLatest = tc.VersionLatest
		if err := prSuite.itIsAValidAndSupportedRelease(); err != nil && !strings.Contains(err.Error(), tc.ExpectedErrorString) {
			t.Fatalf("error: %v", err)
		}
	}
}

func TestGetRequiredTests(t *testing.T) {
	type testCase struct {
		Version             string
		ExpectedTestsCount  int
		ExpectedErrorString string
	}

	for _, tc := range []testCase{
		{
			Version:            "v1.27",
			ExpectedTestsCount: 378,
		},
		{
			Version:             "v1.notfound",
			ExpectedTestsCount:  0,
			ExpectedErrorString: "Malformed version",
		},
		{
			Version:            "v1.26",
			ExpectedTestsCount: 368,
		},
	} {
		prSuite := NewPRSuite(&PullRequest{})
		prSuite.KubernetesReleaseVersion = tc.Version
		tests, err := prSuite.GetRequiredTests()
		if err != nil && !strings.Contains(err.Error(), tc.ExpectedErrorString) {
			t.Fatalf("error: %v", err)
		}
		if len(tests) != tc.ExpectedTestsCount {
			t.Fatalf("error: test count for version %v is expected to be at %v but instead found at %v", tc.Version, tc.ExpectedTestsCount, len(tests))
		}
	}
}

func TestGetMissingJunitTestsFromPRSuite(t *testing.T) {
	type testCase struct {
		Title                     string
		Version                   string
		PullRequest               *PullRequest
		ExpectedTestsMissingCount int
		ExpectedErrorString       string
	}

	for _, tc := range []testCase{
		{
			Title:   `valid junit`,
			Version: "v1.27",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						Name:     "v1.27/coolkube/junit_01.xml",
						BaseName: "junit_01.xml",
						Contents: testGetJunitSubmittedConformanceTestsCoolkubeV127Junit_01xml,
					},
				},
			},
			ExpectedTestsMissingCount: 0,
		},
		{
			Title:   `empty junit`,
			Version: "v1.27",
			PullRequest: &PullRequest{
				SupportingFiles: []*PullRequestFile{
					{
						Name:     "v1.27/coolkube/junit_01.xml",
						BaseName: "junit_01.xml",
						Contents: ``,
					},
				},
			},
			ExpectedTestsMissingCount: 0, // skip since invalid junit anyways
			ExpectedErrorString:       "unable to parse junit_01.xml file",
		},
	} {
		t.Logf("%v", tc.Title)
		prSuite := NewPRSuite(tc.PullRequest)
		prSuite.KubernetesReleaseVersion = tc.Version
		tests, err := prSuite.GetMissingJunitTestsFromPRSuite()
		if err != nil && !strings.Contains(err.Error(), tc.ExpectedErrorString) {
			t.Fatalf("error: %v", err)
		}
		if len(tests) != tc.ExpectedTestsMissingCount {
			t.Fatalf("error: missing test count for version %v is expected to be at %v but instead found at %v", tc.Version, tc.ExpectedTestsMissingCount, len(tests))
		}
	}
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
