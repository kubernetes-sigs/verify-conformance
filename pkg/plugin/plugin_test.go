package plugin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	githubql "github.com/shurcooL/githubv4"
	"k8s.io/test-infra/prow/config"

	"cncf.io/infra/verify-conformance-release/pkg/suite"
)

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

}

func TestNewPRSuiteForPR(t *testing.T) {

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

}

func TestNewPullRequestQueryForGithubPullRequest(t *testing.T) {

}

func TestHandlePullRequestEvent(t *testing.T) {

}

func TestHandleIssueCommentEvent(t *testing.T) {

}

func TestHandleAll(t *testing.T) {

}
