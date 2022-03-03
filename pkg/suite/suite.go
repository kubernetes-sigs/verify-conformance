package suite

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/cucumber/godog"
	semver "github.com/hashicorp/go-version"
	githubql "github.com/shurcooL/githubv4"
	"sigs.k8s.io/yaml"
	// "k8s.io/test-infra/prow/github"

	"cncf.io/infra/verify-conformance-release/internal/types"
)

type ResultPrepare struct {
	Name  string
	Hints []string
}

type PullRequestQuery struct {
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

type PullRequestFile struct {
	BlobURL  string
	Name     string
	BaseName string
	Contents string
}

type PullRequest struct {
	PullRequestQuery

	Labels                  []string
	SupportingFiles         []*PullRequestFile
	ProductYAMLURLDataTypes map[string]string
}

type PRSuite struct {
	PR                             *PullRequest
	KubernetesReleaseVersion       string
	KubernetesReleaseVersionLatest string
	ProductName                    string
	MissingFiles                   []string
	IsNotConformanceSubmission     bool

	Suite  godog.TestSuite
	buffer bytes.Buffer
}

type PRSuiteOptions struct {
	Paths []string
}

func NewPRSuite(PR *PullRequest) *PRSuite {
	return &PRSuite{
		PR: PR,

		buffer: *bytes.NewBuffer(nil),
	}
}

func (s *PRSuite) NewTestSuite(opts PRSuiteOptions) godog.TestSuite {
	s.Suite = godog.TestSuite{
		Name: "how-are-the-prs",
		Options: &godog.Options{
			// Format: "pretty",
			Format: "cucumber",
			Output: &s.buffer,
			Paths:  opts.Paths,
		},
		ScenarioInitializer: s.InitializeScenario,
	}
	return s.Suite
}

func (s *PRSuite) aConformanceProductSubmissionPR() error {
	if s.PR == nil {
		return fmt.Errorf("unable to find PR from query")
	}
	if strings.Contains(strings.ToLower(string(s.PR.Title)), "conformance") != true {
		s.IsNotConformanceSubmission = true
		return godog.ErrPending
	}
	return nil
}

func (s *PRSuite) thePRTitleIsNotEmpty() error {
	if len(s.PR.Title) == 0 {
		return fmt.Errorf("title is empty")
	}
	return nil
}

func (s *PRSuite) isIncludedInItsFileList(file string) error {
	for _, f := range s.PR.SupportingFiles {
		if f.BaseName == file {
			return nil
		}
	}
	s.MissingFiles = append(s.MissingFiles, file)
	return fmt.Errorf("missing file '%v'", file)
}

func (s *PRSuite) fileFolderStructureMatchesRegex(match string) error {
	pattern := regexp.MustCompile(match)

	anyProductFoldersFound := false
	for _, file := range s.PR.SupportingFiles {
		if matches := pattern.MatchString(path.Dir(file.Name)); matches == true {
			anyProductFoldersFound = true
		}
	}
	if anyProductFoldersFound == false {
		s.IsNotConformanceSubmission = true
		return godog.ErrPending
	}
	failureError := fmt.Errorf("your product submission PR be in folders like $KubernetesReleaseVersion/$ProductName, e.g: v1.23/averycooldistro")
	for _, file := range s.PR.SupportingFiles {
		if matches := pattern.MatchString(path.Dir(file.Name)); matches != true {
			return fmt.Errorf("file '%v' not allowed. %v", file.Name, failureError)
		}
		allIndexes := pattern.FindAllSubmatchIndex([]byte(path.Dir(file.Name)), -1)
		for _, loc := range allIndexes {
			baseFolder := string(file.Name[loc[2]:loc[3]])
			distroName := string(file.Name[loc[4]:loc[5]])

			if baseFolder == "" || distroName == "" {
				return failureError
			}
		}
	}
	return nil
}

func (s *PRSuite) thereIsOnlyOnePathOfFolders() error {
	if len(s.PR.SupportingFiles) == 0 {
		return godog.ErrPending
	}
	paths := []string{}
	for _, file := range s.PR.SupportingFiles {
		filePath := path.Dir(file.Name)

		foundInPaths := false
		for _, p := range paths {
			if p == filePath {
				foundInPaths = true
			}
		}
		if filePath == "." {
			continue
		}
		if foundInPaths == false {
			paths = append(paths, filePath)
		}
	}

	if len(paths) != 1 {
		return fmt.Errorf("only one product must be submitted at a time, will use '%v'. Please remove the following: '%v'", paths[0], strings.Join(paths[1:], ", "))
	}

	return nil
}

func (s *PRSuite) theTitleOfThePR() error {
	if s.PR.Title == "" {
		return fmt.Errorf("unable to use product submission PR, as it appears to not have a title")
	}
	return nil
}

func (s *PRSuite) theTitleOfThePRMatches(match string) error {
	pattern := regexp.MustCompile(match)
	if pattern.MatchString(string(s.PR.Title)) != true {
		return fmt.Errorf("title must be formatted like 'Conformance results for $KubernetesReleaseVersion $ProductName' (e.g: Conformance results for v1.23 CoolKubernetes)")
	}
	return nil
}

func (s *PRSuite) theFilesInThePR() error {
	if len(s.PR.SupportingFiles) == 0 {
		return fmt.Errorf("there were no files found in the submission")
	}
	return nil
}

func (s *PRSuite) aFile(fileName string) error {
	if s.ProductName == "" || s.KubernetesReleaseVersion == "" {
		return godog.ErrPending
	}
	file := s.GetFileByFileName(fileName)
	if file == nil {
		return fmt.Errorf("unable to find required file '%v' in list files in product submission PR", fileName)
	}
	return nil
}

func (s *PRSuite) GetFileByFileName(fileName string) *PullRequestFile {
	for _, f := range s.PR.SupportingFiles {
		if strings.ToLower(f.BaseName) == strings.ToLower(fileName) {
			return f
		}
	}
	return nil
}

func (s *PRSuite) theYamlFileContainsTheRequiredAndNonEmptyField(fileName, fieldName string) error {
	var parsedContent map[string]*interface{}
	file := s.GetFileByFileName(fileName)
	if file == nil {
		return fmt.Errorf("unable to find file '%v'", fileName)
	}
	err := yaml.Unmarshal([]byte(file.Contents), &parsedContent)
	if err != nil {
		return fmt.Errorf("unable to read file '%v'", fileName)
	}
	if parsedContent[fieldName] == nil {
		return fmt.Errorf("missing or empty field '%v' in file '%v'", fieldName, fileName)
	}
	return nil
}

func (s *PRSuite) isNotEmpty(fileName string) error {
	file := s.GetFileByFileName(fileName)
	if file == nil {
		return fmt.Errorf("unable to find file '%v'", fileName)
	}
	if file.Contents == "" {
		return fmt.Errorf("file '%v' is empty", fileName)
	}
	return nil
}

func (s *PRSuite) aLineOfTheFileMatches(fileName, match string) error {
	pattern := regexp.MustCompile(match)
	file := s.GetFileByFileName(fileName)
	if file == nil {
		return fmt.Errorf("unable to find file '%v'", fileName)
	}
	lines := strings.Split(file.Contents, "\n")
	foundMatchingLine := false
lineLoop:
	for _, line := range lines {
		foundMatchingLine = pattern.MatchString(line)
		if foundMatchingLine == true {
			break lineLoop
		}
	}
	if foundMatchingLine == false {
		return fmt.Errorf("the file '%v' does not contain a release version of Kubernetes in it", fileName)
	}
	return nil
}

func (s *PRSuite) aListOfLabelsInThePR() error {
	if s.KubernetesReleaseVersion == "" {
		return godog.ErrPending
	}
	if len(s.PR.Labels) == 0 {
		return fmt.Errorf("there are no labels found")
	}
	return nil
}

func (s *PRSuite) theLabelPrefixedWithAndEndingWithKubernetesReleaseVersionShouldBePresent(label string) error {
	labelWithReleaseAttached := label + s.KubernetesReleaseVersion
	foundLabel := false
	for _, l := range s.PR.Labels {
		if l == labelWithReleaseAttached {
			foundLabel = true
		}
	}
	if foundLabel != true {
		return fmt.Errorf("unable to find required label '%v' on this PR. It may be safe to ignore and wait for it to appear if everything else is passing", labelWithReleaseAttached)
	}
	return nil
}

func (s *PRSuite) ifIsSetToUrlTheContentOfTheUrlInTheValueOfMatchesIts(contentType, field, dataType string) error {
	if contentType != "url" {
		return nil
	}
	foundDataType := false
	for _, dt := range strings.Split(dataType, " ") {
		foundDataType = strings.Contains(s.PR.ProductYAMLURLDataTypes[field], dt) == true
		if foundDataType == true {
			break
		}
	}
	if foundDataType == false {
		return fmt.Errorf("unable to use field '%v' in PRODUCT.yaml as the data in the resolved content (%v) doesn't match what is expected (%v). Please ensure that the urls resolve to exact resources intended (especially for an image the exact image url)", field, s.PR.ProductYAMLURLDataTypes[field], strings.Join(strings.Split(dataType, " "), ", or "))
	}
	return nil
}

func (s *PRSuite) SetSubmissionMetadatafromFolderStructure() *PRSuite {
	pattern := regexp.MustCompile(`(v1.[0-9]{2})/(.*)/.*`)

filesLoop:
	for _, file := range s.PR.SupportingFiles {
		allIndexes := pattern.FindAllSubmatchIndex([]byte(file.Name), -1)
		for _, loc := range allIndexes {
			releaseVersion := string(file.Name[loc[2]:loc[3]])
			distroName := string(file.Name[loc[4]:loc[5]])
			s.KubernetesReleaseVersion = releaseVersion
			s.ProductName = distroName
			break filesLoop
		}
	}
	return s
}

func (s *PRSuite) GetLabelsAndCommentsFromSuiteResultsBuffer() (comment string, labels []string, err error) {
	if s.IsNotConformanceSubmission == true {
		return "", []string{}, nil
	}
	cukeFeatures := []types.CukeFeatureJSON{}
	err = json.Unmarshal([]byte(s.buffer.String()), &cukeFeatures)
	if err != nil {
		return "", []string{}, err
	}
	uniquelyNamedStepsRun := []string{}
	resultPrepares := []ResultPrepare{}
	for _, c := range cukeFeatures {
		for _, e := range c.Elements {
			foundNameInStepsRun := false
			resultPrepare := ResultPrepare{}
			hasFails := false
			foundExistingResultTitle := false
			for _, u := range uniquelyNamedStepsRun {
				if u == e.Name {
					foundNameInStepsRun = true
				}
			}
			if foundNameInStepsRun == false {
				uniquelyNamedStepsRun = append(uniquelyNamedStepsRun, e.Name)
			}
		steps:
			for _, s := range e.Steps {
				if s.Result.Status != "failed" {
					continue steps
				}
				hasFails = true
				hint := s.Result.Error
				for ri, r := range resultPrepares {
					hintAlreadyPresentInResult := false
					for _, h := range resultPrepares[ri].Hints {
						if h == hint {
							hintAlreadyPresentInResult = true
						}
					}
					if r.Name == e.Name {
						foundExistingResultTitle = true
					}
					if foundExistingResultTitle && !hintAlreadyPresentInResult {
						resultPrepares[ri].Hints = append(resultPrepares[ri].Hints, hint)
					}
				}
				if foundExistingResultTitle == false {
					resultPrepare.Hints = append(resultPrepare.Hints, hint)
				}
			}
			if hasFails == true && foundExistingResultTitle == false {
				resultPrepare.Name = strings.TrimSpace(e.Description)
				resultPrepares = append(resultPrepares, resultPrepare)
			}
		}
	}

	finalComment := fmt.Sprintf("All requirements (%v) have passed for the submission!", len(uniquelyNamedStepsRun))
	labels = []string{"conformance-product-submission"}
	for _, f := range s.MissingFiles {
		labels = append(labels, "missing-file-"+f)
	}
	if s.KubernetesReleaseVersion != "" {
		labels = append(labels, "release-"+s.KubernetesReleaseVersion)
	}
	if len(resultPrepares) > 0 {
		finalComment = fmt.Sprintf("%v of %v requirements have passed. Please review the following:", len(uniquelyNamedStepsRun)-len(resultPrepares), len(uniquelyNamedStepsRun))
		for _, r := range resultPrepares {
			finalComment += "\n- [FAIL] " + r.Name
			for _, h := range r.Hints {
				finalComment += "\n  - " + h
			}
		}
		labels = append(labels, []string{"not-verifiable"}...)
	} else {
		labels = append(labels, "release-documents-checked")
	}
	finalComment += "\n"

	return finalComment, labels, nil
}

func (s *PRSuite) theReleaseVersionMatchesTheReleaseVersionInTheTitle() error {
	pattern := regexp.MustCompile(`(.*) (v1.[0-9]{2})[ /](.*)`)

	var titleReleaseVersion string
	allIndexes := pattern.FindAllSubmatchIndex([]byte(s.PR.Title), -1)
	for _, loc := range allIndexes {
		titleReleaseVersion = string(s.PR.Title[loc[4]:loc[5]])
		if titleReleaseVersion != "" {
			break
		}
	}
	if titleReleaseVersion != s.KubernetesReleaseVersion {
		return fmt.Errorf("there is a mismatch between the release version in the title (%v) and the release version in the in the folder structure (%v)", titleReleaseVersion, s.KubernetesReleaseVersion)
	}
	return nil
}

func (s *PRSuite) theReleaseVersion() error {
	if s.KubernetesReleaseVersion == "" {
		return godog.ErrPending
	}
	return nil
}

func (s *PRSuite) itIsAValidAndSupportedRelease() error {
	latestVersion, err := semver.NewSemver(s.KubernetesReleaseVersionLatest)
	if err != nil {
		fmt.Printf("error with go-version parsing latestVersion '%v': %v\n", s.KubernetesReleaseVersionLatest, err)
		return fmt.Errorf("unable to parse latest release version")
	}
	currentVersion, err := semver.NewSemver(s.KubernetesReleaseVersion)
	if err != nil {
		fmt.Printf("error with go-version parsing currentVersion '%v': %v\n", currentVersion, err)
		return fmt.Errorf("unable to parse latest release version")
	}
	latestVersionSegments := latestVersion.Segments()
	latestVersionSegments[1] -= 3
	oldestVersion := fmt.Sprintf("v%v.%v", latestVersionSegments[0], latestVersionSegments[1])
	oldestSupportedVersion, err := semver.NewSemver(oldestVersion)
	if err != nil {
		fmt.Printf("error with go-version parsing oldest release version '%v': %v\n", latestVersionSegments, err)
		return fmt.Errorf("unable to parse oldest supported release version")
	}

	if currentVersion.GreaterThan(latestVersion) {
		return fmt.Errorf("unable to use version '%v' because it is newer than the current supported release (%v)", s.KubernetesReleaseVersion, s.KubernetesReleaseVersionLatest)
	} else if currentVersion.LessThan(oldestSupportedVersion) {
		return fmt.Errorf("unable to use version '%v' because it is older than the last currently supported release (%v)", s.KubernetesReleaseVersion, oldestVersion)
	}
	return nil
}

func (s *PRSuite) InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^a conformance product submission PR$`, s.aConformanceProductSubmissionPR)
	ctx.Step(`^the PR title is not empty$`, s.thePRTitleIsNotEmpty)
	ctx.Step(`^"([^"]*)" is included in its file list$`, s.isIncludedInItsFileList)
	ctx.Step(`^the files in the PR`, s.theFilesInThePR)
	ctx.Step(`^file folder structure matches "([^"]*)"$`, s.fileFolderStructureMatchesRegex)
	ctx.Step(`^the title of the PR$`, s.theTitleOfThePR)
	ctx.Step(`^the title of the PR matches "([^"]*)"$`, s.theTitleOfThePRMatches)
	ctx.Step(`^the yaml file "([^"]*)" contains the required and non-empty "([^"]*)"$`, s.theYamlFileContainsTheRequiredAndNonEmptyField)
	ctx.Step(`^a "([^"]*)" file$`, s.aFile)
	ctx.Step(`^"([^"]*)" is not empty$`, s.isNotEmpty)
	ctx.Step(`^a line of the file "([^"]*)" matches "([^"]*)"$`, s.aLineOfTheFileMatches)
	ctx.Step(`^a list of labels in the PR$`, s.aListOfLabelsInThePR)
	ctx.Step(`^the label prefixed with "([^"]*)" and ending with Kubernetes release version should be present$`, s.theLabelPrefixedWithAndEndingWithKubernetesReleaseVersionShouldBePresent)
	ctx.Step(`^if "([^"]*)" is set to url, the content of the url in the value of "([^"]*)" matches it\'s "([^"]*)"$`, s.ifIsSetToUrlTheContentOfTheUrlInTheValueOfMatchesIts)
	ctx.Step(`^there is only one path of folders$`, s.thereIsOnlyOnePathOfFolders)
	ctx.Step(`^the release version matches the release version in the title$`, s.theReleaseVersionMatchesTheReleaseVersionInTheTitle)
	ctx.Step(`^the release version$`, s.theReleaseVersion)
	ctx.Step(`^it is a valid and supported release$`, s.itIsAValidAndSupportedRelease)
}
