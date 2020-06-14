/*
Copyright 2020 CNCF TODO Check how this code should be licensed

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package plugin

import (
	"bytes"
	"context"
	"fmt"
        "regexp"
        "strings"
	"time"

	githubql "github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"

	"k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/pluginhelp"
	"k8s.io/test-infra/prow/plugins"
)

const (
	PluginName     = "verify-conformance-request"
	needsVersionReview = "Please ensure that the logs provided correspond to the version referenced in the title of this PR."
	verifyLabel    = "Verified version"
)

var sleep = time.Sleep

type githubClient interface {
	GetIssueLabels(org, repo string, number int) ([]github.Label, error)
	CreateComment(org, repo string, number int, comment string) error
	BotName() (string, error)
	AddLabel(org, repo string, number int, label string) error
	RemoveLabel(org, repo string, number int, label string) error
	DeleteStaleComments(org, repo string, number int, comments []github.IssueComment, isStale func(github.IssueComment) bool) error
	Query(context.Context, interface{}, map[string]interface{}) error
	GetPullRequest(org, repo string, number int) (*github.PullRequest, error)
	GetPullRequestChanges(org, repo string, number int) ([]github.PullRequestChange, error)
}

type commentPruner interface {
	PruneComments(shouldPrune func(github.IssueComment) bool)
}

// HelpProvider constructs the PluginHelp for this plugin that takes into account enabled repositories.
// HelpProvider defines the type for the function that constructs the PluginHelp for plugins.
func HelpProvider(_ []config.OrgRepo) (*pluginhelp.PluginHelp, error) {
	return &pluginhelp.PluginHelp{
			Description: `The Verify Conformance Request plugin checks the content of PRs that request Conformance Certification for Kubernetes to see if they are internally consistent. So, for example, if the title of the PR contains a reference to a Kubernetes version then this plugin checks to see that the Sonobouy e2e test logs refer to the same version.`,
		},
		nil
}

// HandlePullRequestEvent handles a GitHub pull request event
func HandlePullRequestEvent(log *logrus.Entry, ghc githubClient, pre *github.PullRequestEvent) error {
	log.Infof("HandlePullRequestEvent")
	if pre.Action != github.PullRequestActionOpened && pre.Action != github.PullRequestActionSynchronize && pre.Action != github.PullRequestActionReopened {
		return nil
	}

	return handle(log, ghc, &pre.PullRequest)
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

	return handle(log, ghc, pr)
}

// handle checks a Conformance Certification PR to determine if the contents of the PR pass sanity checks.
// Adds a comment to indicate whther or not the version in the PR title occurs in the supplied logs.
func handle(log *logrus.Entry, ghc githubClient, pr *github.PullRequest) error {
	log.Infof("handle")
	if pr.Merged {
		return nil
	}

	org := pr.Base.Repo.Owner.Login
	repo := pr.Base.Repo.Name
	number := pr.Number
	sha := pr.Head.SHA

	verifiable, releaseVersion, err := HasReleaseInPrTitle(log, ghc, string(pr.Title))
	log.Infof("verifiable is %v, commit sha is %q, release version is %v", verifiable, sha, releaseVersion)
	if err != nil {
		return err
	}
	issueLabels, err := ghc.GetIssueLabels(org, repo, number)
	log.Infof("IssueLabels %v ", issueLabels)
	if err != nil {
		return err
	}
	return nil // takeAction(log, ghc, org, repo, number, pr.User.Login, hasLabel, verifiable)
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

	if len(orgs) == 0 && len(repos) == 0 {
		log.Warnf("HandleAll : No repos have been configured for the %s plugin", PluginName)
		return nil
	}

        // TODO simplify queryOpenPRs
        //      - more general than required
        //      - we deal with a single org and repo
        //      - we target k8s conformance requests sent to the cncf
	var queryOpenPRs bytes.Buffer
	fmt.Fprint(&queryOpenPRs, "archived:false is:pr is:open")
	for _, org := range orgs {
		fmt.Fprintf(&queryOpenPRs, " org:\"%s\"", org)
	}
	for _, repo := range repos {
		fmt.Fprintf(&queryOpenPRs, " repo:\"%s\"", repo)
	}
	prs, err := search(context.Background(), log, ghc, queryOpenPRs.String())

	if err != nil {
		return err
	}
	log.Infof("Considering %d PRs.", len(prs))

	for _, pr := range prs {
		org := string(pr.Repository.Owner.Login)
		repo := string(pr.Repository.Name)
		prNumber := int(pr.Number)
		sha := string(pr.Commits.Nodes[0].Commit.Oid)

                hasReleaseInTitle, releaseVersion, err := HasReleaseInPrTitle(log,ghc,string(pr.Title))

                hasReleaseLabel, err := HasReleaseLabel(log, org, repo, prNumber, ghc, "release-"+releaseVersion)

		prLogger := log.WithFields(logrus.Fields{
			"org":  org,
			"repo": repo,
			"prNumber":   prNumber,
                        "release": releaseVersion,
		})

                if err != nil {
                        prLogger.WithError(err).Error("Failed to find a release in title")
                        githubClient.CreateComment(ghc, org, repo, prNumber, "Please include the release in the title of this Pull Request" )
                }

                if hasReleaseInTitle && !hasReleaseLabel {
                        logsHaveSpecifiedRelease, err := checkLogsForK8sRelease(prLogger, ghc, org, repo, prNumber, sha, releaseVersion)

                        if err != nil {
                                prLogger.WithError(err).Error("Failed to find a releaseVersion in files")
                        }

                        if logsHaveSpecifiedRelease && !hasReleaseLabel {
                                githubClient.AddLabel(ghc, org, repo, prNumber, "verifiable")
                                githubClient.AddLabel(ghc, org, repo, prNumber, "release-"+releaseVersion)
                                githubClient.CreateComment(ghc, org, repo, prNumber, "Found " + releaseVersion + " in logs" )
                        } else {
                                githubClient.AddLabel(ghc, org, repo, prNumber, "not-verifiable")
                                githubClient.CreateComment(ghc, org, repo, prNumber, "This request is not yet verifiable. We cannot find a reference to " + releaseVersion + "in the logs you supplied with this PR")
                        }
                } else {
                        githubClient.AddLabel(ghc, org, repo, prNumber, "not verifiable")
                        githubClient.CreateComment(ghc, org, repo, prNumber, "This conformance request is not yet verifiable. Please ensure that PR Title refernces the Kubernetes Release and that the supplied logs refer to the specified Release")
		}
        }
	return nil
}


func HasReleaseLabel(prLogger *logrus.Entry, org,repo string, prNumber int, ghc githubClient, releaseLabel string ) (bool,error) {
        hasReleaseLabel := false
	labels, err := ghc.GetIssueLabels(org, repo, prNumber)

        if err != nil {
                prLogger.WithError(err).Error("Failed to find labels")
        }

        for foundLabel := range labels {
                releaseCheck := strings.Compare(labels[foundLabel].Name,releaseLabel)
                if releaseCheck == 0 {
			hasReleaseLabel := true
                        break
                }
        }

        return hasReleaseLabel, err
}
// TODO make this fn more cohesive and fix name
func HasReleaseInPrTitle(log *logrus.Entry, ghc githubClient, prTitle string)  (bool, string, error) {
        hasReleaseInTitle := false
        k8sRelease := ""
        log.WithFields(logrus.Fields{
                "PR Title": prTitle,
        })
	log.Infof("IsVerifiable: title of PR is %q", prTitle)
        k8sVerRegExp := regexp.MustCompile(`v[0-9]\.[0-9][0-9]*`)
        titleContainsVersion, err := regexp.MatchString(`v[0-9]\.[0-9][0-9]*`, prTitle)
        if err != nil {
                log.WithError(err).Error("Error matching k8s version in PR title")
        }
        if (titleContainsVersion) {
                k8sRelease = k8sVerRegExp.FindString(prTitle)
                log.WithFields(logrus.Fields{
                        "Version": k8sRelease,
                })
                hasReleaseInTitle = true
        }
        return hasReleaseInTitle, k8sRelease, nil
}

// takeAction adds or removes the "preliminary_verified" label based on the current
// state of the PR (hasLabel and isVerified). It also handles adding and
// removing GitHub comments notifying the PR author that the request has been verified
func takeAction(log *logrus.Entry, ghc githubClient, org, repo string, num int, author string, hasLabel, verifiable bool) error {
	if !verifiable && !hasLabel {
		if err := ghc.AddLabel(org, repo, num, verifyLabel); err != nil {
			log.WithError(err).Errorf("Failed to add %q label.", verifyLabel)
		}
		msg := plugins.FormatSimpleResponse(author, "Version Mismatch")
		return ghc.CreateComment(org, repo, num, msg)
	} else if verifiable && hasLabel {
		// remove label and prune comment
		if err := ghc.RemoveLabel(org, repo, num, "Version Mismatch"); err != nil {
			log.WithError(err).Errorf("Failed to remove %q label.", "")
		}
		botName, err := ghc.BotName()
		if err != nil {
			return err
		}
		return ghc.DeleteStaleComments(org, repo, num, nil, shouldPrune(botName))
	}
	return nil
}

// Checks changes associated with the supplied sha to see if the contain a reference to k8sRelease
// returns true if k8sRelease found , false otherwise
func checkLogsForK8sRelease(prLogger *logrus.Entry, ghc githubClient, org, repo string, prNumber int, sha, k8sRelease string ) (bool,error) {
	logsHaveStatedRelease := false
	changes, err := ghc.GetPullRequestChanges(org, repo, prNumber)
	if err != nil {
		return logsHaveStatedRelease, err
	}

	prLogger.Infof("checkLogsForK8sRelease: %+v", changes)

	for _ , change := range changes {
		// https://developer.github.com/v3/pulls/#list-pull-requests-files
		patchContainsVersion, err := regexp.MatchString(`v[0-9]\.[0-9][0-9]*`, change.Patch)
		if err != nil {
			return logsHaveStatedRelease, err
		}

		if (patchContainsVersion){
			logsHaveStatedRelease =true
		}

	}
	return logsHaveStatedRelease , err
}

func shouldPrune(botName string) func(github.IssueComment) bool {
	return func(ic github.IssueComment) bool {
		return github.NormLogin(botName) == github.NormLogin(ic.User.Login) &&
			strings.Contains(ic.Body, needsVersionReview)
	}
}
// Executes the search query contained in q using the GitHub client ghc
func search(ctx context.Context, log *logrus.Entry, ghc githubClient, q string) ([]PullRequest, error) {
	var ret []PullRequest
	vars := map[string]interface{}{
		"query":        githubql.String(q),
		"searchCursor": (*githubql.String)(nil),
	}
	var totalCost int
	var remaining int
	for {
		sq := SearchQuery{}
		if err := ghc.Query(ctx, &sq, vars); err != nil {
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
	Title githubql.String
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
			PullRequest PullRequest `graphql:"... on PullRequest"`
		}
	} `graphql:"search(type: ISSUE, first: 100, after: $searchCursor, query: $query)"`
}
