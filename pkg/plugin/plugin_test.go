package plugin

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	githubql "github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
	"k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/github"
	"sigs.k8s.io/yaml"

	"cncf.io/infra/verify-conformance-release/pkg/common"
	"cncf.io/infra/verify-conformance-release/pkg/suite"
)

var (
	log = logrus.StandardLogger().WithField("plugin", "verify-conformance-release")

	//go:embed testdata/TestGetJunitSubmittedConformanceTests-coolkube-v1-27-junit_01.xml
	testGetJunitSubmittedConformanceTestsCoolkubeV127Junit_01xml string
)

type prContext struct {
	PullRequestQuery *suite.PullRequestQuery
	SupportingFiles  []*suite.PullRequestFile
}

type FakeGitHubClient struct {
	PopulatedPullRequests []*prContext
}

func NewFakeGitHubClient(p []*prContext) *FakeGitHubClient {
	return &FakeGitHubClient{
		PopulatedPullRequests: p,
	}
}

func (f *FakeGitHubClient) CreateStatus(org string, repo string, headRefOID string, status github.Status) error {
	return nil
}
func (f *FakeGitHubClient) GetCombinedStatus(org, repo, ref string) (*github.CombinedStatus, error) {
	return nil, nil
}
func (f *FakeGitHubClient) GetIssueLabels(org, repo string, number int) ([]github.Label, error) {
	labels := []github.Label{}
	var prIndex *int
	for i := range f.PopulatedPullRequests {
		if f.PopulatedPullRequests[i].PullRequestQuery.Number == githubql.Int(number) {
			prIndex = &i
		}
	}
	if prIndex == nil {
		return []github.Label{}, fmt.Errorf("unable to find pr '%v'", number)
	}
	for _, l := range f.PopulatedPullRequests[*prIndex].PullRequestQuery.Labels.Nodes {
		labels = append(labels, github.Label{Name: string(l.Name)})
	}
	return labels, nil
}
func (f *FakeGitHubClient) CreateComment(org, repo string, number int, comment string) error {
	return nil
}
func (f *FakeGitHubClient) ListIssueCommentsWithContext(ctx context.Context, org, repo string, number int) ([]github.IssueComment, error) {
	return []github.IssueComment{}, nil
}
func (f *FakeGitHubClient) BotUserChecker() (func(candidate string) bool, error) {
	return func(string) bool { return false }, nil
}
func (f *FakeGitHubClient) AddLabel(org, repo string, number int, label string) error {
	var prIndex *int
	for i := range f.PopulatedPullRequests {
		if f.PopulatedPullRequests[i].PullRequestQuery.Number == githubql.Int(number) {
			prIndex = &i
		}
	}
	if prIndex == nil {
		return fmt.Errorf("unable to find label '%v' in pr number '%v'", label, number)
	}
	f.PopulatedPullRequests[*prIndex].PullRequestQuery.Labels.Nodes = append(f.PopulatedPullRequests[*prIndex].PullRequestQuery.Labels.Nodes, struct{ Name githubql.String }{githubql.String(label)})
	return nil
}
func (f *FakeGitHubClient) RemoveLabel(org, repo string, number int, label string) error {
	var prIndex *int
	for i := range f.PopulatedPullRequests {
		if f.PopulatedPullRequests[i].PullRequestQuery.Number == githubql.Int(number) {
			prIndex = &i
		}
	}
	if prIndex == nil {
		return fmt.Errorf("unable to find label '%v' in pr number '%v'", label, number)
	}
	var labelIndex *int
	for i, l := range f.PopulatedPullRequests[*prIndex].PullRequestQuery.Labels.Nodes {
		if l.Name == githubql.String(label) {
			labelIndex = &i
		}
	}
	if labelIndex == nil {
		return fmt.Errorf("unable to find label '%v' in pr number '%v'", label, number)
	}
	f.PopulatedPullRequests[*prIndex].PullRequestQuery.Labels.Nodes = append(f.PopulatedPullRequests[*prIndex].PullRequestQuery.Labels.Nodes[:*labelIndex], f.PopulatedPullRequests[*prIndex].PullRequestQuery.Labels.Nodes[*labelIndex+1:]...)
	return nil
}
func (f *FakeGitHubClient) DeleteStaleComments(org, repo string, number int, comments []github.IssueComment, isStale func(github.IssueComment) bool) error {
	return nil
}
func (f *FakeGitHubClient) QueryWithGitHubAppsSupport(ctx context.Context, sq interface{}, vars map[string]interface{}, org string) error {
	if org == "nil" {
		return fmt.Errorf("org does not exist")
	}
	// wrap each pull request in an array struct, as per search query nodes
	if len(f.PopulatedPullRequests) > 0 && f.PopulatedPullRequests[0] == nil {
		return fmt.Errorf("empty pr")
	}
	nodes := func() []struct {
		PullRequest suite.PullRequestQuery "graphql:\"... on PullRequest\""
	} {
		o := []struct {
			PullRequest suite.PullRequestQuery "graphql:\"... on PullRequest\""
		}{}
		for _, pr := range f.PopulatedPullRequests {
			if pr.PullRequestQuery == nil {
				continue
			}
			o = append(o, struct {
				PullRequest suite.PullRequestQuery "graphql:\"... on PullRequest\""
			}{
				PullRequest: *pr.PullRequestQuery,
			})
		}
		return o
	}()
	query, ok := sq.(*SearchQuery)
	if !ok {
		return fmt.Errorf("failed to case sq to SearchQuery")
	}
	hasNextPage := false
	// TODO tidy this
	searchCursor := func() string {
		s := vars["searchCursor"].(*githubql.String)
		if s == nil {
			return "1"
		}
		return string(*s)
	}()
	if searchCursor == "1" {
		hasNextPage = true
		searchCursor = "2"
	} else if searchCursor == "2" {
		hasNextPage = false
		searchCursor = "3"
	} else {
		searchCursor = "1"
	}
	*query = SearchQuery{
		RateLimit: struct {
			Cost      githubql.Int
			Remaining githubql.Int
		}{
			Cost:      githubql.Int(1),
			Remaining: githubql.Int(999999),
		},
		Search: struct {
			PageInfo struct {
				HasNextPage githubql.Boolean
				EndCursor   githubql.String
			}
			Nodes []struct {
				PullRequest suite.PullRequestQuery "graphql:\"... on PullRequest\""
			}
		}{
			PageInfo: struct {
				HasNextPage githubql.Boolean
				EndCursor   githubql.String
			}{
				HasNextPage: githubql.Boolean(hasNextPage),
				EndCursor:   githubql.String(searchCursor),
			},
			Nodes: nodes,
		},
	}
	return nil
}
func (f *FakeGitHubClient) GetPullRequest(org, repo string, number int) (*github.PullRequest, error) {
	return nil, nil
}
func (f *FakeGitHubClient) GetPullRequestChanges(org, repo string, number int) ([]github.PullRequestChange, error) {
	pr := &prContext{}
	prChanges := []github.PullRequestChange{}
	for _, n := range f.PopulatedPullRequests {
		if n.PullRequestQuery.Number == githubql.Int(number) {
			pr = n
			break
		}
	}
	for _, file := range pr.SupportingFiles {
		prChanges = append(prChanges, github.PullRequestChange{
			Filename: file.Name,
			BlobURL:  file.BlobURL,
		})
	}
	return prChanges, nil
}

func TestHelpProvider(t *testing.T) {
	hp, err := HelpProvider([]config.OrgRepo{})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if hp.Description != `The Verify Conformance Request plugin checks the content of PRs that request Conformance Certification for Kubernetes to see if they are internally consistent. So, for example, if the title of the PR contains a reference to a Kubernetes version then this plugin checks to see that the Sonobouy e2e test logs refer to the same version.` {
		t.Fatalf("error: HelpProvider description is unexpected; %v", hp.Description)
	}
}

func TestFetchFileFromURI(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`Hello!`))
		if err != nil {
			t.Fatalf("error: sending http response; %v", err)
		}
	}))
	defer svr.Close()
	content, resp, err := fetchFileFromURI(svr.URL)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if content != `Hello!` {
		t.Fatalf("error: content doesn't match what is expected")
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("error: response code doesn't match what was expected")
	}

}

func TestRawURLForBlobURL(t *testing.T) {
	type testCase struct {
		BlobURL           string
		RawUserContentURL string
	}
	for _, u := range []testCase{
		{
			BlobURL:           "https://github.com/smira/k8s-conformance/blob/2c25ea5963e88ad77a8035dc639c7e3a60b8fb0f/v1.27/talos/PRODUCT.yaml",
			RawUserContentURL: "https://raw.githubusercontent.com/smira/k8s-conformance/2c25ea5963e88ad77a8035dc639c7e3a60b8fb0f/v1.27/talos/PRODUCT.yaml",
		},
		{
			BlobURL:           "https://github.com/cncf/apisnoop/blob/main/505_output_coverage_jsons.sql",
			RawUserContentURL: "https://raw.githubusercontent.com/cncf/apisnoop/main/505_output_coverage_jsons.sql",
		},
		{
			BlobURL:           "https://github.com/cncf-infra/verify-conformance/blob/main/README.org",
			RawUserContentURL: "https://raw.githubusercontent.com/cncf-infra/verify-conformance/main/README.org",
		},
	} {
		output := rawURLForBlobURL(u.BlobURL)
		if output != u.RawUserContentURL {
			t.Fatalf("error: url string (%v) replace does not match what is expected (%v)", output, u.RawUserContentURL)
		}
	}
}

func TestSearch(t *testing.T) {
	type testCase struct {
		Name                string
		PullRequestQuery    *suite.PullRequestQuery
		ExpectedErrorString string
	}
	for _, tc := range []testCase{
		{
			Name: "complete result",
			PullRequestQuery: &suite.PullRequestQuery{
				Number: githubql.Int(1),
				Author: struct{ Login githubql.String }{
					Login: githubql.String("cncf"),
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
			},
		},
		{
			Name: "org does not exist",
			PullRequestQuery: &suite.PullRequestQuery{
				Number: githubql.Int(1),
				Author: struct{ Login githubql.String }{
					Login: githubql.String("nil"),
				},
				Repository: struct {
					Name  githubql.String
					Owner struct{ Login githubql.String }
				}{
					Name: githubql.String("k8s-conformance"),
					Owner: struct{ Login githubql.String }{
						Login: githubql.String("nil"),
					},
				},
			},
			ExpectedErrorString: "org does not exist",
		},
		{
			Name:                "empty",
			PullRequestQuery:    nil,
			ExpectedErrorString: "empty pr",
		},
	} {
		ghc := NewFakeGitHubClient([]*prContext{
			{
				PullRequestQuery: tc.PullRequestQuery,
			},
		})
		var org string
		if tc.PullRequestQuery != nil {
			org = string(tc.PullRequestQuery.Repository.Owner.Login)
		}
		prs, err := search(context.TODO(), log, ghc, "archived:false is:pr is:open repo:\"k8s-conformance\"", org)
		t.Logf("tc(%v) has error %v", tc.Name, err != nil)
		if err != nil && !strings.Contains(err.Error(), tc.ExpectedErrorString) {
			t.Fatalf("error searching for PRs: %v", err)
		}
		t.Logf("tc(%v) %+v\n", tc.Name, prs)
	}
}

func TestNewPRSuiteForPR(t *testing.T) {
	type testCase struct {
		Name                string
		PullRequestQuery    *suite.PullRequestQuery
		Labes               []github.Label
		SupportingFiles     []*suite.PullRequestFile
		ExpectedErrorString string
	}

	common.DataPathPrefix = "../../"

	for _, tc := range []testCase{
		{
			Name: "valid pull request entry",
			Labes: []github.Label{
				{
					Name: "conformance-product-submission",
				},
			},
			PullRequestQuery: &suite.PullRequestQuery{
				Number: githubql.Int(1),
				Repository: struct {
					Name  githubql.String
					Owner struct{ Login githubql.String }
				}{
					Name: githubql.String("cncf-ci"),
					Owner: struct{ Login githubql.String }{
						Login: githubql.String("cncf-ci"),
					},
				},
			},
			SupportingFiles: []*suite.PullRequestFile{
				{
					Name:     "v1.27/coolkube/README.md",
					BaseName: "README.md",
					BlobURL:  "README.md",
					Contents: `# CoolKube`,
				},
				{
					Name:     "v1.27/coolkube/PRODUCT.yaml",
					BaseName: "PRODUCT.yaml",
					BlobURL:  "PRODUCT.yaml",
					Contents: `vendor: "cool"
name: "coolkube"
version: "v1.27"
type: "distribution"
description: "it's just all-round cool and probably the best k8s, idk"
website_url: "website_url"
documentation_url: "docs"
contact_email_address: "sales@coolkubernetes.com"`,
				},
				{
					Name:     "v1.27/coolkube/junit_01.xml",
					BaseName: "junit_01.xml",
					BlobURL:  "junit_01.xml",
					Contents: testGetJunitSubmittedConformanceTestsCoolkubeV127Junit_01xml,
				},
				{
					Name:     "v1.27/coolkube/e2e.log",
					BaseName: "e2e.log",
					BlobURL:  "e2e.log",
					Contents: `cool!`,
				},
			},
		},
	} {
		productYAML := map[string]string{}
		var productYAMLSupportingFile string
		for _, file := range tc.SupportingFiles {
			if file.BaseName == "PRODUCT.yaml" {
				productYAMLSupportingFile = file.Contents
			}
		}
		if productYAMLSupportingFile != "" {
			if err := yaml.Unmarshal([]byte(productYAMLSupportingFile), &productYAML); err != nil {
				t.Fatalf("error: unmarshalling from PRODUCT.yaml supporting file: %v", err)
			}
		}

		svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Logf("requesting path '%v'", r.URL.Path)
			supportingFile := &suite.PullRequestFile{}
			for _, file := range tc.SupportingFiles {
				if r.URL.Path == "/"+file.BaseName || r.URL.Path == "/"+file.Name {
					supportingFile = file
				}
			}
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(supportingFile.Contents))
			if err != nil {
				t.Fatalf("error: sending http response; %v", err)
			}
		}))
		defer svr.Close()
		for _, field := range []string{"website_url", "documentation_url"} {
			if productYAML[field] != "" {
				productYAML[field] = svr.URL + "/" + productYAML[field]
			}
		}
		productYAMLBytes, err := yaml.Marshal(productYAML)
		if err != nil {
			t.Fatalf("error: marshalling new product yaml: %v", err)
		}
		for i := range tc.SupportingFiles {
			tc.SupportingFiles[i].BlobURL = svr.URL + "/" + tc.SupportingFiles[i].BlobURL
			if tc.SupportingFiles[i].BaseName == "PRODUCT.yaml" {
				tc.SupportingFiles[i].Contents = string(productYAMLBytes)
			}
		}
		ghc := NewFakeGitHubClient([]*prContext{
			{
				PullRequestQuery: tc.PullRequestQuery,
				SupportingFiles:  tc.SupportingFiles,
			},
		})
		prSuite, err := NewPRSuiteForPR(log, ghc, tc.PullRequestQuery)
		if err != nil && strings.Contains(err.Error(), tc.ExpectedErrorString) {
			t.Fatalf("unexpected error in testcase '%v': %v", tc.Name, err)
		}
		t.Logf("prSuite: %+v\n", prSuite)
	}
}

func TestGetGodogPaths(t *testing.T) {
	paths := GetGodogPaths()
	found := false
	for _, p := range paths {
		if p == "../../kodata/features/" || p == "./kodata/features/" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("error: unable to find features folder for godog paths")
	}
}

func TestLabelIsManaged(t *testing.T) {
	type testCase struct {
		Label          string
		ExpectedResult bool
	}

	for _, tc := range []testCase{
		{
			Label:          "conformance-product-submission",
			ExpectedResult: true,
		},
		{
			Label:          "not-verifiable",
			ExpectedResult: true,
		},
		{
			Label:          "release-documents-checked",
			ExpectedResult: true,
		},
		{
			Label:          "required-tests-missing",
			ExpectedResult: true,
		},
		{
			Label:          "evidence-missing",
			ExpectedResult: true,
		},
		{
			Label:          "unable-to-process",
			ExpectedResult: true,
		},
		{
			Label:          "some-kinda-label",
			ExpectedResult: false,
		},
		{
			Label:          "non-managed",
			ExpectedResult: false,
		},
	} {
		if result := labelIsManaged(tc.Label); result != tc.ExpectedResult {
			t.Fatalf("error: label (%v) not expected to be managed", tc.Label)
		}
	}
}

func TestLabelIsVersionLabel(t *testing.T) {
	type testCase struct {
		Label          string
		Version        string
		ExpectedResult bool
	}

	for _, tc := range []testCase{
		{
			Label:          "release-v1.27",
			Version:        "v1.27",
			ExpectedResult: true,
		},
		{
			Label:          "release-v1.26",
			Version:        "v1.26",
			ExpectedResult: true,
		},
		{
			Label:          "no-failed-tests-v1.27",
			Version:        "v1.27",
			ExpectedResult: true,
		},
		{
			Label:          "no-failed-tests-v1.26",
			Version:        "v1.26",
			ExpectedResult: true,
		},
		{
			Label:          "tests-verified-v1.27",
			Version:        "v1.27",
			ExpectedResult: true,
		},
		{
			Label:          "am-i-a-label-v1.27",
			Version:        "v1.27",
			ExpectedResult: false,
		},
		{
			Label:          "thing",
			Version:        "v1.27",
			ExpectedResult: false,
		},
	} {
		if result := labelIsVersionLabel(tc.Label, tc.Version); result != tc.ExpectedResult {
			t.Fatalf("error: version label (%v) does not match expected result (%v)", tc.Label, tc.ExpectedResult)
		}
	}
}

func TestLabelIsFileLabel(t *testing.T) {
	type testCase struct {
		Name           string
		Label          string
		MissingFiles   []string
		ExpectedResult bool
	}

	for _, tc := range []testCase{
		{
			Label:          "missing-file-README.md",
			MissingFiles:   []string{"README.md"},
			ExpectedResult: true,
		},
		{
			Label:          "missing-file-e2e.log",
			MissingFiles:   []string{"e2e.log"},
			ExpectedResult: true,
		},
		{
			Label:          "missing-file-junit_01.xml",
			MissingFiles:   []string{"junit_01.xml"},
			ExpectedResult: true,
		},
		{
			Label:          "missing-file-PRODUCT.yaml",
			MissingFiles:   []string{"PRODUCT.yaml"},
			ExpectedResult: true,
		},
		{
			Label:          "missing-file-README.md",
			ExpectedResult: true,
		},
		{
			Label:          "missing-file-e2e.log",
			ExpectedResult: true,
		},
		{
			Label:          "missing-file-junit_01.xml",
			ExpectedResult: true,
		},
		{
			Label:          "missing-file-PRODUCT.yaml",
			ExpectedResult: true,
		},
		{
			Label:          "hi-im-a-label",
			ExpectedResult: false,
		},
		{
			Label:          "missing-fil-PRODUCT.yaml",
			ExpectedResult: false,
		},
	} {
		if result := labelIsFileLabel(tc.Label, tc.MissingFiles); result != tc.ExpectedResult {
			t.Fatalf("error: file label is not expected for %v (%v) with result (%v)", tc.Label, tc.ExpectedResult, result)
		}
	}
}

func TestUpdateLabels(t *testing.T) {

}

func TestUpdateComments(t *testing.T) {

}

func TestRemoveSliceOfStringsFromStringSlice(t *testing.T) {
	type testCase struct {
		Input          []string
		Remove         []string
		ExpectedOutput []string
	}

	for _, tc := range []testCase{
		{
			Input:          []string{"a", "b", "c", "d", "e"},
			Remove:         []string{"b", "e"},
			ExpectedOutput: []string{"a", "c", "d"},
		},
	} {
		result := removeSliceOfStringsFromStringSlice(tc.Input, tc.Remove)
		if len(result) != len(tc.ExpectedOutput) {
			t.Fatalf("error: slice (%v) length is mismatching (%v)", len(result), len(tc.ExpectedOutput))
		}
		itemsMatching := []string{}
		for _, i := range tc.ExpectedOutput {
			for _, r := range result {
				if i == r {
					itemsMatching = append(itemsMatching, r)
				}
			}
		}
		if len(itemsMatching) != len(tc.ExpectedOutput) {
			t.Fatalf("error: items matching count (%v) doesn't equal the expected output count (%v)", len(itemsMatching), len(tc.ExpectedOutput))
		}
	}
}

func TestIsConformancePR(t *testing.T) {
	type testCase struct {
		PR             *suite.PullRequestQuery
		ExpectedResult bool
	}

	for _, tc := range []testCase{
		{
			PR: &suite.PullRequestQuery{
				Title: githubql.String("conformance results for v1.27/thing"),
			},
			ExpectedResult: true,
		},
		{
			PR: &suite.PullRequestQuery{
				Title: githubql.String("conformance results for v1.27/kubernetes"),
			},
			ExpectedResult: true,
		},
		{
			PR: &suite.PullRequestQuery{
				Title: githubql.String("conformanc"),
			},
			ExpectedResult: false,
		},
	} {
		if isConformancePR(tc.PR) != tc.ExpectedResult {
			t.Fatalf("error: conformance PR with name (%v) is expected to be a conformance PR", tc.PR.Title)
		}
	}
}

func TestUpdateStatus(t *testing.T) {

}

func TestHandle(t *testing.T) {
	type testCase struct {
		Name                    string
		KubernetesVersion       *string
		KubernetesVersionLatest *string
		PullRequestQuery        *suite.PullRequestQuery
		SupportingFiles         []*suite.PullRequestFile
		Labels                  []string
		ExpectedLabels          []string
		ExpectedComment         string
		ExpectedStatus          string
		ExpectedError           string
	}

	for _, tc := range []testCase{
		{
			Name:                    "valid submission",
			Labels:                  []string{"conformance-product-submission"},
			KubernetesVersion:       common.Pointer("v1.27"),
			KubernetesVersionLatest: common.Pointer("v1.27"),
			SupportingFiles: []*suite.PullRequestFile{
				{
					Name:     "v1.27/coolkube/README.md",
					BaseName: "README.md",
					Contents: `# coolkube
> the coolest Kubernetes distribution

## Generating conformance results

1. create a coolkube cluster
2. sonobuoy run --wait && sonobuoy results "$(sonobuoy retrieve)" && sonobuoy delete --wait`,
					BlobURL: "README.md",
				},
				{
					Name:     "v1.27/coolkube/PRODUCT.yaml",
					BaseName: "PRODUCT.yaml",
					Contents: `vendor: "cool"
name: "coolkube"
version: "v1.27"
type: "distribution"
description: "it's just all-round cool and probably the best k8s, idk"
website_url: "website_url"
documentation_url: "docs"
contact_email_address: "sales@coolkubernetes.com"`,
					BlobURL: "PRODUCT.yaml",
				},
				{
					Name:     "v1.27/coolkube/e2e.log",
					BaseName: "e2e.log",
					Contents: "",
					BlobURL:  "e2e.log",
				},
				{
					Name:     "v1.27/coolkube/junit_01.xml",
					BaseName: "junit_01.xml",
					Contents: testGetJunitSubmittedConformanceTestsCoolkubeV127Junit_01xml,
					BlobURL:  "junit_01.xml",
				},
			},
			PullRequestQuery: &suite.PullRequestQuery{
				Title: githubql.String("Conformance results for v1.27/coolkube"),
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
		productYAML := map[string]string{}
		var productYAMLSupportingFile string
		for _, file := range tc.SupportingFiles {
			if file.BaseName == "PRODUCT.yaml" {
				productYAMLSupportingFile = file.Contents
			}
		}
		if productYAMLSupportingFile != "" {
			if err := yaml.Unmarshal([]byte(productYAMLSupportingFile), &productYAML); err != nil {
				t.Fatalf("error: unmarshalling from PRODUCT.yaml supporting file: %v", err)
			}
		}
		svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Logf("requesting path '%v'", r.URL.Path)
			supportingFile := &suite.PullRequestFile{}
			for _, file := range tc.SupportingFiles {
				if r.URL.Path == "/"+file.BaseName || r.URL.Path == "/"+file.Name {
					supportingFile = file
				}
			}
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(supportingFile.Contents))
			if err != nil {
				t.Fatalf("error: sending http response; %v", err)
			}
		}))
		defer svr.Close()
		for _, field := range []string{"website_url", "documentation_url"} {
			if productYAML[field] != "" {
				productYAML[field] = svr.URL + "/" + productYAML[field]
			}
		}
		productYAMLBytes, err := yaml.Marshal(productYAML)
		if err != nil {
			t.Fatalf("error: marshalling new product yaml: %v", err)
		}
		for i := range tc.SupportingFiles {
			tc.SupportingFiles[i].BlobURL = svr.URL + "/" + tc.SupportingFiles[i].BlobURL
			if tc.SupportingFiles[i].BaseName == "PRODUCT.yaml" {
				tc.SupportingFiles[i].Contents = string(productYAMLBytes)
			}
		}
		ghc := NewFakeGitHubClient([]*prContext{
			{
				PullRequestQuery: tc.PullRequestQuery,
				SupportingFiles:  tc.SupportingFiles,
			},
		})
		if err := handle(log, ghc, tc.PullRequestQuery); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// TODO check comment, labels and status
	}
}

func TestNewPullRequestQueryForGithubPullRequest(t *testing.T) {
	if prq := NewPullRequestQueryForGithubPullRequest(
		"cncf",
		"k8s-conformance",
		0,
		&github.PullRequest{
			User: github.User{
				Login: "cncf-ci",
			},
		},
	); prq == nil {
		t.Fatalf("PullRequestQuery must never be empty")
	}
}

func TestHandlePullRequestEvent(t *testing.T) {

}

func TestHandleIssueCommentEvent(t *testing.T) {

}

func TestHandleAll(t *testing.T) {

}
