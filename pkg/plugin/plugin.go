package plugin

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	githubql "github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
	"k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/pluginhelp"
	"k8s.io/test-infra/prow/plugins"
	"sigs.k8s.io/yaml"

	"cncf.io/infra/verify-conformance-release/pkg/suite"
)

const (
	PluginName = "verify-conformance-release"
)

var (
	productYAMLRequiredFieldDateTypes = []ProductYAMLField{
		{Field: "website_url"},
		{Field: "repo_url"},
		{Field: "documentation_url"},
		{Field: "product_logo_url"},
	}
	godogPaths = []string{"./features/", "./kodata/features/", "/var/run/ko/features/"}
)

type ProductYAMLField struct {
	Field string
}

type githubClient interface {
	GetIssueLabels(org, repo string, number int) ([]github.Label, error)
	CreateComment(org, repo string, number int, comment string) error
	BotUserChecker() (func(candidate string) bool, error)
	AddLabel(org, repo string, number int, label string) error
	RemoveLabel(org, repo string, number int, label string) error
	DeleteStaleCommentsWithContext(ctx context.Context, org, repo string, number int, comments []github.IssueComment, isStale func(github.IssueComment) bool) error
	QueryWithGitHubAppsSupport(context.Context, interface{}, map[string]interface{}, string) error
	GetPullRequest(org, repo string, number int) (*github.PullRequest, error)
	GetPullRequestChanges(org, repo string, number int) ([]github.PullRequestChange, error)
}

type commentPruner interface {
	PruneComments(shouldPrune func(github.IssueComment) bool)
}

type PullRequest struct {
	Number githubql.Int
	Author struct {
		Login githubql.String
	}
	Repository struct {
		Name  githubql.String
		Owner struct {
			Login githubql.String
		}
	}
	Labels struct {
		Nodes []struct {
			Name githubql.String
		}
	} `graphql:"labels(first:100)"`
	Files struct {
		Nodes []struct {
			Path githubql.String
		}
	} `graphql:"files(first:10)"`
	Title   githubql.String
	Commits struct {
		Nodes []struct {
			Commit struct {
				Oid githubql.String
			}
		}
	} `graphql:"commits(first:5)"`
}

type SearchQuery struct {
	RateLimit struct {
		Cost      githubql.Int
		Remaining githubql.Int
	}
	Search struct {
		PageInfo struct {
			HasNextPage githubql.Boolean
			EndCursor   githubql.String
		}
		Nodes []struct {
			PullRequest suite.PullRequestQuery `graphql:"... on PullRequest"`
		}
	} `graphql:"search(type: ISSUE, first: 100, after: $searchCursor, query: $query)"`
}

// HelpProvider constructs the PluginHelp for this plugin that takes into account enabled repositories.
// HelpProvider defines the type for the function that constructs the PluginHelp for plugins.
func HelpProvider(_ []config.OrgRepo) (*pluginhelp.PluginHelp, error) {
	return &pluginhelp.PluginHelp{
			Description: `The Verify Conformance Request plugin checks the content of PRs that request Conformance Certification for Kubernetes to see if they are internally consistent. So, for example, if the title of the PR contains a reference to a Kubernetes version then this plugin checks to see that the Sonobouy e2e test logs refer to the same version.`,
		},
		nil
}

func fetchFileFromURI(uri string) (content string, resp *http.Response, err error) {
	resp, err = http.Get(uri)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}
	return string(body), resp, nil
}

// takes a patchUrl from a githubClient.PullRequestChange and transforms it
// to produce the url that delivers the raw file associated with the patch.
// Tested for small files.
func rawURLForBlobURL(patchUrl string) string {
	fileUrl := strings.Replace(patchUrl, "github.com", "raw.githubusercontent.com", 1)
	fileUrl = strings.Replace(fileUrl, "/blob", "", 1)
	return fileUrl
}

// Executes the search query contained in q using the GitHub client ghc
func search(ctx context.Context, log *logrus.Entry, ghc githubClient, q string, org string) ([]suite.PullRequestQuery, error) {
	var ret []suite.PullRequestQuery
	vars := map[string]interface{}{
		"query":        githubql.String(q),
		"searchCursor": (*githubql.String)(nil),
	}
	var totalCost int
	var remaining int
	for {
		sq := SearchQuery{}
		log.Infof("query \"%s\" ", q)
		if err := ghc.QueryWithGitHubAppsSupport(ctx, &sq, vars, org); err != nil {
			return nil, err
		}
		totalCost += int(sq.RateLimit.Cost)
		remaining = int(sq.RateLimit.Remaining)
		for _, n := range sq.Search.Nodes {
			ret = append(ret, n.PullRequest)
		}
		if !sq.Search.PageInfo.HasNextPage {
			break
		}
		vars["searchCursor"] = githubql.NewString(sq.Search.PageInfo.EndCursor)
	}
	log.Infof("Search for query \"%s\" cost %d point(s). %d remaining.", q, totalCost, remaining)
	return ret, nil
}

func NewPRSuiteForPR(log *logrus.Entry, ghc githubClient, pr *suite.PullRequestQuery) (prSuite *suite.PRSuite, err error) {
	prSuite = suite.NewPRSuite(&suite.PullRequest{PullRequestQuery: *pr})
	issueLabels, err := ghc.GetIssueLabels(string(pr.Repository.Owner.Login), string(pr.Repository.Name), int(pr.Number))
	if err != nil {
		return &suite.PRSuite{}, fmt.Errorf("error fetching PR issue labels for issue (%v), %v ", pr.Number, err)
	}
	for _, l := range issueLabels {
		prSuite.PR.Labels = append(prSuite.PR.Labels, l.Name)
	}

	var productYAMLContent string
	changes, err := ghc.GetPullRequestChanges(string(pr.Repository.Owner.Login), string(pr.Repository.Name), int(pr.Number))
	if err != nil {
		return &suite.PRSuite{}, fmt.Errorf("error fetching PR (%v) changes, %v", pr.Number, err)
	}
	for _, c := range changes {
		content, _, err := fetchFileFromURI(rawURLForBlobURL(c.BlobURL))
		if err != nil {
			return &suite.PRSuite{}, fmt.Errorf("error fetching content of '%v' in PR (%v) via '%v', %v", c.Filename, pr.Number, c.BlobURL, err)
		}

		baseName := path.Base(c.Filename)
		prFile := &suite.PullRequestFile{
			Name:     c.Filename,
			BaseName: baseName,
			BlobURL:  c.BlobURL,
			Contents: content,
		}
		prSuite.PR.SupportingFiles = append(prSuite.PR.SupportingFiles, prFile)

		if baseName == "PRODUCT.yaml" {
			productYAMLContent = content
		}
	}
	if productYAMLContent == "" {
		return &suite.PRSuite{}, fmt.Errorf("failed to find PRODUCT.yaml from the list of files in the PR (%v)", pr.Number)
	}

	productYAML := map[string]string{}
	err = yaml.Unmarshal([]byte(productYAMLContent), &productYAML)
	if err != nil {
		return &suite.PRSuite{}, fmt.Errorf("failed to parse content of PRODUCT.yaml in PR (%v), %v", pr.Number, err)
	}

	for _, f := range productYAMLRequiredFieldDateTypes {
		uri := productYAML[f.Field]
		u, err := url.Parse(uri)
		if err != nil {
			return &suite.PRSuite{}, fmt.Errorf("failed to parse url '%v' of the field '%v' in PRODUCT.yaml in PR (%v), %v", uri, pr.Number, err)
		}
		resp, err := http.Head(u.String())
		if err != nil {
			return &suite.PRSuite{}, fmt.Errorf("failed to make a HEAD request to url '%v' from the field '%v' in PRODUCT.yaml in PR (%v), %v", u, pr.Number, err)
		}
		contentType := resp.Header.Get("Content-Type")
		if prSuite.PR.ProductYAMLURLDataTypes == nil {
			prSuite.PR.ProductYAMLURLDataTypes = map[string]string{}
		}
		prSuite.PR.ProductYAMLURLDataTypes[f.Field] = contentType
	}

	return prSuite, nil
}

func GetGodogPaths() (paths []string) {
	for _, p := range godogPaths {
		if _, err := os.Stat(p); os.IsNotExist(err) == true {
			continue
		}
		paths = append(paths, p)
	}
	return paths
}

// handle checks a Conformance Certification PR to determine if the contents of the PR pass sanity checks.
// Adds a comment to indicate whether or not the version in the PR title occurs in the supplied logs.
func handle(log *logrus.Entry, ghc githubClient, pr *suite.PullRequestQuery) error {
	godogFeaturePaths := GetGodogPaths()
	prSuite, err := NewPRSuiteForPR(log, ghc, pr)
	if err != nil {
		return err
	}

	prSuite.SetSubmissionMetadatafromFolderStructure()
	prSuite.NewTestSuite(suite.PRSuiteOptions{Paths: godogFeaturePaths}).Run()

	finalComment, labels, err := prSuite.GetLabelsAndCommentsFromSuiteResultsBuffer()
	if err != nil {
		return err
	}

	fmt.Println("PR title:", prSuite.PR.Title)
	fmt.Println("Release Version:", prSuite.KubernetesReleaseVersion)
	fmt.Println("Labels:", strings.Join(labels, ", "))
	fmt.Println(finalComment)
	return nil
}

func NewPullRequestQueryForGithubPullRequest(orgName string, repoName string, number int, pr *github.PullRequest) *suite.PullRequestQuery {
	return &suite.PullRequestQuery{
		Title:  githubql.String(pr.Title),
		Number: githubql.Int(number),
		Author: struct {
			Login githubql.String
		}{
			Login: githubql.String(pr.User.Login),
		},
		Repository: struct {
			Name  githubql.String
			Owner struct {
				Login githubql.String
			}
		}{
			Name: githubql.String(repoName),
			Owner: struct {
				Login githubql.String
			}{
				Login: githubql.String(pr.User.Login),
			},
		},
	}
}

// HandlePullRequestEvent handles a GitHub pull request event
func HandlePullRequestEvent(log *logrus.Entry, ghc githubClient, pre *github.PullRequestEvent) error {
	log.Infof("HandlePullRequestEvent")
	if pre.Action != github.PullRequestActionOpened && pre.Action != github.PullRequestActionSynchronize && pre.Action != github.PullRequestActionReopened {
		return nil
	}

	return handle(log, ghc, NewPullRequestQueryForGithubPullRequest(pre.Repo.Owner.Login, pre.Repo.Name, pre.Number, &pre.PullRequest))
}

// HandleIssueCommentEvent handles a GitHub issue comment event and adds or removes a
// message indicating that there are inconsitencies in the version of Kubernetes
// referenced in the title of the PR versus the log file evidence supplied in the associated commit.
func HandleIssueCommentEvent(log *logrus.Entry, ghc githubClient, ice *github.IssueCommentEvent) error {
	log.Infof("HandleIssueCommentEvent")
	if !ice.Issue.IsPullRequest() {
		return nil
	}
	pr, err := ghc.GetPullRequest(ice.Repo.Owner.Login, ice.Repo.Name, ice.Issue.Number)
	if err != nil {
		return err
	}

	return handle(log, ghc, NewPullRequestQueryForGithubPullRequest(ice.Repo.Owner.Login, ice.Repo.Name, ice.Issue.Number, pr))
}

// HandleAll is called periodically and the period is setup in main.go
// It runs a Github Query to get all open PRs for this repo which contains k8s conformance requests
//
// Each PR is checked in turn, we check
//   - for the presence of a Release Version in the PR title
//- then we take that version and verify that the e2e test logs refer to that same release version.
//
// if all is in order then we add the verifiable label and a release-Vx.y label
// if there is an inconsistency we add a comment that explains the problem
// and tells the PR submitter to review the documentation
func HandleAll(log *logrus.Entry, ghc githubClient, config *plugins.Configuration) error {
	log.Infof("%v : HandleAll : Checking all PRs for handling", PluginName)

	orgs, repos := config.EnabledReposForExternalPlugin(PluginName) // TODO : Overkill see below
	log.Infof("orgs: %#v, repos: %#v", orgs, repos)

	if len(orgs) == 0 && len(repos) == 0 {
		log.Warnf("HandleAll : No repos have been configured for the %s plugin", PluginName)
		return nil
	}

	var queryOpenPRs bytes.Buffer
	//	fmt.Fprint(&queryOpenPRs, "archived:false is:pr is:open -label:verifiable")
	fmt.Fprint(&queryOpenPRs, "archived:false is:pr is:open ")
	for _, repo := range repos {
		slashSplit := strings.Split(repo, "/")
		if n := len(slashSplit); n != 2 {
			log.WithField("repo", repo).Warn("Found repo that was not in org/repo format, ignoring...")
			continue
		}
		org := slashSplit[0]
		orgs = append(orgs, org)
		fmt.Fprintf(&queryOpenPRs, " repo:\"%s\"", repo)
	}
	for _, org := range orgs {
		fmt.Fprintf(&queryOpenPRs, " org:\"%s\"", org)
	}

	prs := []suite.PullRequestQuery{}
	for _, org := range orgs {
		prSearch, err := search(context.Background(), log, ghc, queryOpenPRs.String(), org)
		if err != nil {
			return err
		}
		prs = append(prs, prSearch...)
	}
	log.Infof("Considering %d PRs.", len(prs))

	for _, pr := range prs {
		err := handle(log, ghc, &pr)
		if err != nil {
			log.Infof("error running checks on PR: %v", err)
		}
	}
	return nil
}
