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
	"path"
	"net/http"
	"io/ioutil"
	"github.com/golang-collections/collections/set"
)

const (
	PluginName     = "verify-conformance-request"
	needsVersionReview = "Please ensure that the logs provided correspond to the version referenced in the title of this PR."
	verifyLabel    = "release consistent"
)

var sleep = time.Sleep
//var requiredProductFieldsSet = set.New("vendor", "name", "version", "website_url", "repo_url", "documentation_url", "product_logo_url", "type", "description")
var requiredProductFieldsSet = set.New("vendor", "name", "version", "website_url", "documentation_url", "product_logo_url", "type", "description")
var requiredProductFields = map[string]string {
	"vendor" : "Name of the legal entity that is certifying. This entity must have a signed participation form on file with the CNCF",
	"name" : "Name of the product being certified.",
	"version" : "The version of the product being certified (not the version of Kubernetes it runs).",
	"website_url" : "URL to the product information website",
	//"repo_url" : "If your product is open source, this field is necessary to point to the primary GitHub repo containing the source. It's OK if this is a mirror. OPTIONAL",
	"documentation_url" : "URL to the product documentation",
	"product_logo_url" : "URL to the product's logo, (must be in SVG, AI or EPS format -- not a PNG -- and include the product name). OPTIONAL. If not supplied, we'll use your company logo. Please see logo guidelines",
	"type" : "Is your product a distribution, hosted platform, or installer (see definitions)",
	"description" :	"One sentence description of your offering",
}

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
	//	fmt.Fprint(&queryOpenPRs, "archived:false is:pr is:open -label:verifiable")
	fmt.Fprint(&queryOpenPRs, "archived:false is:pr is:open ")
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
			//"org":  org,
			//"repo": repo,
			"pr":   prNumber,
                        "title": pr.Title,
                        "statedRelease": releaseVersion,
		})

                if err != nil {
                        prLogger.WithError(err).Error("Failed to find a release in title")
                        githubClient.CreateComment(ghc, org, repo, prNumber, "Please include the release in the title of this Pull Request" )
                }

		hasNotVerifiableLabel, err := HasNotVerifiableLabel(log, org, repo, prNumber, ghc)
                if hasReleaseInTitle && !hasReleaseLabel {
                        changesHaveSpecifiedRelease, err := checkChangesHaveStatedK8sRelease(prLogger, ghc, org, repo, prNumber, sha, releaseVersion)

                        if err != nil {
                                prLogger.WithError(err)
                        }

			log.Infof("cHSR returns %v", changesHaveSpecifiedRelease)
			if changesHaveSpecifiedRelease && !hasReleaseLabel {
				//   githubClient.AddLabel(ghc, org, repo, prNumber, "verifiable")
                                //githubClient.AddLabel(ghc, org, repo, prNumber, "release-"+releaseVersion)
                                githubClient.AddLabel(ghc, org, repo, prNumber, "release-documents-checked")
                                githubClient.AddLabel(ghc, org, repo, prNumber, "release-"+releaseVersion)
                                githubClient.CreateComment(ghc, org, repo, prNumber, "Found " + releaseVersion + " in logs" )
                                if hasNotVerifiableLabel {
                                        githubClient.RemoveLabel(ghc, org, repo, prNumber, "not-verifiable")
                                }
                        } else { // specifiedRelease not present in logs
                                if !hasNotVerifiableLabel {
                                        // githubClient.AddLabel(ghc, org, repo, prNumber, "not-verifiable")
					// githubClient.CreateComment(ghc, org, repo, prNumber, "This request is not yet verifiable.")

					// TODO move changesHaveSpecifiedRelease back into handleall
					// I need to report on individual failures to apply the correct lable
					// the following code is a repeat of the same code we declared in changesHaveSpecifiedRelease
					e2eLogHasRelease := false
					productYamlCorrect := false
					foldersCorrect := false
					productYamlDiff := set.New()


					changes, err := ghc.GetPullRequestChanges(org, repo, prNumber)
					if err != nil {
						prLogger.WithError(err)
					}
					var supportingFiles = make ( map[string] github.PullRequestChange  )
					for _ , change := range changes {
						// https://developer.github.com/v3/pulls/#list-pull-requests-files
						supportingFiles[path.Base(change.Filename)] = change
						//		prLogger.Infof("cCHSKR: %+v", supportingFiles[path.Base(change.Filename)])
					}

					productYamlCorrect, productYamlDiff = checkProductYAMLHasRequiredFields(prLogger,supportingFiles["PRODUCT.yaml"])
					foldersCorrect = checkFilesAreInCorrectFolders(prLogger,supportingFiles, releaseVersion)
					e2eLogHasRelease = checkE2eLogHasRelease(prLogger,supportingFiles["e2e.log"], releaseVersion)

					// This is why I repeat the code above, I need to be able to write individual lables based on failure reason

					if !productYamlCorrect {
						var prodYamlDiffString = fmt.Sprintf("%v[1]", productYamlDiff)
						prLogger.Infof("pYC in HANDLEALL productYamlCorrect returned %v\n",productYamlCorrect)

						githubClient.CreateComment(ghc, org, repo, prNumber, "This request is not yet verifiable, please confirm that your product file ( PRODUCT.yaml ) is named correctly and have all the fields listed in  [How to submit conformance results](https://github.com/cncf/k8s-conformance/blob/master/instructions.md#productyaml) . Please make sure you included the following fields:"+prodYamlDiffString)
						//	githubClient.CreateComment(ghc, org, repo, prNumber, "You are missing the following fields"+prodYamlDiffString)
						if !hasNotVerifiableLabel {
							githubClient.AddLabel(ghc, org, repo, prNumber, "not-verifiable")
						}
					}
					if !e2eLogHasRelease {
						prLogger.Infof("eLHR in HANDLEALL e2eLogHasRelease returned %v\n",e2eLogHasRelease)
						githubClient.CreateComment(ghc, org, repo, prNumber, "This request is not yet verifiable, please confirm that your e2e logs reference the release you are submitting for")
						if !hasNotVerifiableLabel {
							githubClient.AddLabel(ghc, org, repo, prNumber, "not-verifiable")
						}
					}
					if !foldersCorrect{
						prLogger.Infof("fC in HANDLEALL foldersCorrect returned %v\n",foldersCorrect)
						githubClient.CreateComment(ghc, org, repo, prNumber, "This request is not yet verifiable, please confirm that your supporting files are in the correct folder.")
						if !hasNotVerifiableLabel {
							githubClient.AddLabel(ghc, org, repo, prNumber, "not-verifiable")
						}
					}
                                }
                        }
                } else if !hasNotVerifiableLabel && !hasReleaseLabel {
                        githubClient.AddLabel(ghc, org, repo, prNumber, "not-verifiable")
                        githubClient.CreateComment(ghc, org, repo, prNumber, "This conformance request is not yet verifiable. Please ensure that PR Title refernces the Kubernetes Release and that the supplied logs refer to the specified Release")
		} //else {
		//   break
		//	}
        }
	return nil
}

// TODO Consolodate this and the next function to cerate a map of labels
func HasNotVerifiableLabel(prLogger *logrus.Entry, org,repo string, prNumber int, ghc githubClient) (bool,error) {
        hasNotVerifiableLabel := false
	labels, err := ghc.GetIssueLabels(org, repo, prNumber)

        if err != nil {
                prLogger.WithError(err).Error("Failed to find labels")
        }

        for foundLabel := range labels {
                notVerifiableCheck := strings.Compare(labels[foundLabel].Name,"not-verifiable")
                if notVerifiableCheck == 0 {
			hasNotVerifiableLabel = true
                        break
                }
        }

        return hasNotVerifiableLabel, err
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
			hasReleaseLabel = true
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

// Checks that changes associated with the pull request contain correct references to k8sRelease
// returns true if k8sRelease found in both the paths to the files and the files themselves, false otherwise
// error contains information as to where the release was missing
func checkChangesHaveStatedK8sRelease(prLogger *logrus.Entry, ghc githubClient, org, repo string, prNumber int, sha, k8sRelease string ) (bool,error) {
	changesHaveStatedRelease := false

	e2eLogHasRelease := false
	productYamlCorrect := false
	foldersCorrect := false
	productYamlDiff := set.New()

	missingProductFields := set.New()
	changes, err := ghc.GetPullRequestChanges(org, repo, prNumber)

	if err != nil {
		return changesHaveStatedRelease, err
		prLogger.WithError(err)
	}

	// Create a map of filenames to changes associated with the filename
	// we are doing this so that we can handle all of the file specific
	// checks that we need to do
        var supportingFiles = make ( map[string] github.PullRequestChange  )
	for _ , change := range changes {
		// https://developer.github.com/v3/pulls/#list-pull-requests-files
		supportingFiles[path.Base(change.Filename)] = change
		//		prLogger.Infof("cCHSKR: %+v", supportingFiles[path.Base(change.Filename)])
	}

	// Do all our checks
	// e2eLogHasRelease = checkPatchContainsRelease(prLogger,supportingFiles["e2e.log"], k8sRelease)
	var prodYamlDiffString = fmt.Sprintf("%v", productYamlDiff)
	var gitCommentProductYaml = fmt.Sprintf("You are missing the following fields %v .", prodYamlDiffString)
	productYamlCorrect, productYamlDiff = checkProductYAMLHasRequiredFields(prLogger,supportingFiles["PRODUCT.yaml"])
	foldersCorrect = checkFilesAreInCorrectFolders(prLogger,supportingFiles, k8sRelease)
	e2eLogHasRelease = checkE2eLogHasRelease(prLogger,supportingFiles["e2e.log"], k8sRelease)

	prLogger.Infof("pYC productYamlCorrect returned %v\n",gitCommentProductYaml)
	prLogger.Infof("pYC productYamlCorrect returned %v\n",productYamlCorrect)
	prLogger.Infof("fC foldersCorrect returned %v\n",foldersCorrect)
	prLogger.Infof("eLHR e2eLogHasRelease %v\n",e2eLogHasRelease)

	if ( e2eLogHasRelease && productYamlCorrect && foldersCorrect) {
		changesHaveStatedRelease = true
	} else {
		// TODO we may want to consider more maintainable and more complete error reporting here
		// Leave this for now implemnt checks first
		var errMsg strings.Builder
		if !e2eLogHasRelease {
			fmt.Fprintf(&errMsg, "Release %s missing from e2e.log file\n", k8sRelease)
		}

		if !productYamlCorrect {
			fmt.Fprintf(&errMsg, "Product.YAML is missing the following required fields %v", missingProductFields)
			prLogger.Infof("fC should be logging the fields missing from  %v\n", missingProductFields)
		}

		if !foldersCorrect {
			fmt.Fprintf(&errMsg, "The files supplied for release %s are not in the correct folders", k8sRelease )
		}

		err = fmt.Errorf(errMsg.String())
	}

	prLogger.WithError(err)
	return changesHaveStatedRelease, err
}

func checkPatchContainsRelease(log *logrus.Entry, change github.PullRequestChange, k8sRelease string)(bool){
	log.Infof("checkPatchContainsRelease: patch is %v\n ",change.Patch)
	return strings.Contains(change.Patch, k8sRelease)
}

func checkFilesAreInCorrectFolders(log *logrus.Entry, changes map[string] github.PullRequestChange, k8sRelease string)(bool){
	filesAreInCorrectReleaseFoldersBool := false

	for _ , change := range changes {
		filesAreInCorrectReleaseFolders := strings.Contains(change.Filename, k8sRelease)
		if filesAreInCorrectReleaseFolders {
			log.Infof("cFAICF found files only in stated  release folder %s", k8sRelease)
                        filesAreInCorrectReleaseFoldersBool = true
			break
		}
	}
	return filesAreInCorrectReleaseFoldersBool

}

// takes a patchUrl from a githubClient.PullRequestChange and transforms it
// to produce the url that delivers the raw file associated with the patch.
// Tested for small files.
func patchUrlToFileUrl(patchUrl string) (string){
	fileUrl := strings.Replace(patchUrl, "github.com", "raw.githubusercontent.com", 1)
	fileUrl = strings.Replace(fileUrl, "/blob", "", 1)
        return fileUrl
}
// Retrieves e2eLogfile and checks that it contains k8sRelease
func checkE2eLogHasRelease(log *logrus.Entry, e2eChange github.PullRequestChange, k8sRelease string) (bool){
        e2eLogHasStatedRelease := false

        fileUrl := patchUrlToFileUrl(e2eChange.BlobURL)
	resp, err := http.Get(fileUrl)
	if err != nil {
		log.Errorf("cELHR : %+v",err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)


	// Make a set that contains all the key fields in the Product YAML file
        // TODO Check to see if string(body) performant
	for _, line := range strings.Split(string(body), "\n") {
                if strings.Contains(line, k8sRelease){
                        log.Infof("cELHR found stated release!! %s",line)
                        e2eLogHasStatedRelease = true
                        break
                }
        }
        return e2eLogHasStatedRelease

}

func checkProductYAMLHasRequiredFields(log *logrus.Entry, productYaml github.PullRequestChange)(bool, *set.Set){
	allRequiredFieldsPresent := false
	productFields := set.New()
	// ref https://github.com/cncf/k8s-conformance/blob/master/instructions.md#productyaml
        difference := set.New()

	if productYaml.BlobURL != "" {
	// TODO return a list of the missing fields
		// missingFields  := make([]string, len(requiredProductFields))
		log.Infof("cPYHRf: PY CHANGE %+v\n",productYaml)

		fileUrl := patchUrlToFileUrl(productYaml.BlobURL)

		log.Infof("cPYHRf: PY PATH  %+v\n",fileUrl)

		resp, err := http.Get(fileUrl)
		if resp.StatusCode > 199 && resp.StatusCode < 300 {
			// TODO check body for 404
			if err != nil {
				log.Errorf("Error retrieving conformance tests metadata from : %s", fileUrl)
				log.Errorf("HTTP Reponse was: %+v", resp)
				log.Errorf("getRequiredTests : %+v", err)
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Errorf("cPYHRf : %+v",err)
			}
			// Make a set that contains all the key fields in the Product YAML file
			for _, line := range strings.Split(string(body), "\n") {
				// extract the key field regEx start of line to first occurance of :
				key := strings.Split(line,":")
				// Add key to fieldSet
				if len(key[0]) > 0 {
					log.Infof("%s", key[0])
					productFields.Insert(key[0])
				}
			}
			// Difference the requiredFieldsSet against productFields found here
			difference = requiredProductFieldsSet.Difference(productFields)

			if difference.Len() == 0 {
				allRequiredFieldsPresent = true
			} else {
				log.Infof("THESE FIELDS ARE MISSING! %v", difference)
			}
		}
	}
	return allRequiredFieldsPresent, difference

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
		log.Infof("query \"%s\" ", q)
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
