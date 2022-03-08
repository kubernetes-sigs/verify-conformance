package suite

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
	semver "github.com/hashicorp/go-version"
	githubql "github.com/shurcooL/githubv4"
	"sigs.k8s.io/yaml"

	"cncf.io/infra/verify-conformance-release/internal/types"
	"cncf.io/infra/verify-conformance-release/pkg/common"
)

// TODO ensure file checking

var (
	lastSupportingVersions = 3
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

type ConformanceTestMetadata struct {
	Testname    string `yaml:"testname"`
	Codename    string `yaml:"codename"`
	Description string `yaml:"description"`
	Release     string `yaml:"release"`
	File        string `yaml:"file"`
}

type JunitTestCase struct {
	XMLName xml.Name  `xml:"testcase"`
	Name    string    `xml:"name,attr"`
	Skipped *struct{} `xml:"skipped"`
}

type JunitTestSuite struct {
	TestSuite []JunitTestCase `xml:"testcase"`
}

type PRSuiteOptions struct {
	Paths []string
}

type PRSuite struct {
	PR                             *PullRequest
	KubernetesReleaseVersion       string
	KubernetesReleaseVersionLatest string
	ProductName                    string
	MissingFiles                   []string
	E2eLogKubernetesReleaseVersion string
	Labels                         []string

	MetadataFolder string
	Suite          godog.TestSuite
	buffer         bytes.Buffer
}

func (s *PRSuite) GetRequiredTests() (tests map[string]bool, err error) {
	versionSemver, err := semver.NewSemver(s.KubernetesReleaseVersion)
	if err != nil {
		return map[string]bool{}, err
	}
	var conformanceMetadata []ConformanceTestMetadata
	content, err := common.ReadFile(path.Join(s.MetadataFolder, s.KubernetesReleaseVersion, "conformance.yaml"))
	if err != nil {
		return map[string]bool{}, err
	}
	err = yaml.Unmarshal([]byte(content), &conformanceMetadata)
	if err != nil {
		return map[string]bool{}, err
	}
	tests = map[string]bool{}
	for _, test := range conformanceMetadata {
		foundInTestVersions := false
	testSupportedVersions:
		for _, r := range strings.Split(test.Release, ",") {
			testVersionSemver, err := semver.NewSemver(r)
			if err != nil {
				return map[string]bool{}, err
			}
			if versionSemver.GreaterThanOrEqual(testVersionSemver) == true {
				foundInTestVersions = true
			}
			if foundInTestVersions == true {
				break testSupportedVersions
			}
		}
		if foundInTestVersions != true {
			continue
		}
		tests[test.Codename] = false
	}
	return tests, nil
}

func (s *PRSuite) GetJunitSubmittedConformanceTests() (tests []string, err error) {
	file := s.GetFileByFileName("junit_01.xml")
	if file == nil {
		return []string{}, fmt.Errorf("unable to find file junit_01.xml")
	}
	testSuite := JunitTestSuite{}
	if err := xml.Unmarshal([]byte(file.Contents), &testSuite); err != nil {
		return []string{}, fmt.Errorf("unable to parse junit_01.xml file, %v", err)
	}
	for _, testcase := range testSuite.TestSuite {
		if testcase.Skipped != nil {
			continue
		}
		if strings.Contains(testcase.Name, "[Conformance]") == false {
			continue
		}
		testcase.Name = strings.Replace(testcase.Name, "&#39;", "'", -1)
		testcase.Name = strings.Replace(testcase.Name, "&#34;", "\"", -1)
		testcase.Name = strings.Replace(testcase.Name, "&gt;", ">", -1)
		testcase.Name = strings.Replace(testcase.Name, "'cat /tmp/health'", "\"cat /tmp/health\"", -1)
		tests = append(tests, testcase.Name)
	}

	return tests, nil
}

func (s *PRSuite) GetMissingJunitTestsFromPRSuite() (missingTests []string, err error) {
	requiredTests, err := s.GetRequiredTests()
	if err != nil {
		return []string{}, err
	}
	submittedTests, err := s.GetJunitSubmittedConformanceTests()
	if err != nil {
		return []string{}, err
	}

	for _, submittedTest := range submittedTests {
		if _, found := requiredTests[submittedTest]; found != true {
			continue
		}
		requiredTests[submittedTest] = true
	}
	for test, found := range requiredTests {
		if found == true {
			continue
		}
		missingTests = append(missingTests, test)
	}

	return missingTests, nil
}

func (s *PRSuite) DetermineE2eLogSucessful() (success bool, passed int, err error) {
	file := s.GetFileByFileName("e2e.log")
	if file == nil {
		return false, 0, fmt.Errorf("unable to find file e2e.log")
	}
	fileLines := strings.Split(file.Contents, "\n")
	lastLinesAmount := len(fileLines) - 10
	if lastLinesAmount < 0 {
		lastLinesAmount = len(fileLines)
	}
	fileLast10Lines := fileLines[lastLinesAmount:]
	patternComplete := regexp.MustCompile(`^SUCCESS! -- ([1-9][0-9]+) Passed \| ([0-9]+) Failed \| ([0-9]+) Pending \| ([0-9]+) Skipped$`)
	matchingLine := ""
	for _, line := range fileLast10Lines {
		if patternComplete.MatchString(line) == true {
			matchingLine = line
		}
	}
	if matchingLine == "" {
		return false, 0, fmt.Errorf("unable to determine test results (passed, failed, pending, skipped) from e2e.log")
	}
	allIndexes := patternComplete.FindAllSubmatchIndex([]byte(matchingLine), -1)
	for _, loc := range allIndexes {
		passed, err = strconv.Atoi(matchingLine[loc[2]:loc[3]])
		if err != nil {
			return false, 0, fmt.Errorf("failed to parse successful tests")
		}
		// failed := string(file.Name[loc[4]:loc[5]])
		// pending := string(file.Name[loc[6]:loc[7]])
		// skipped := string(file.Name[loc[8]:loc[9]])
	}
	return true, passed, nil
}

func NewPRSuite(PR *PullRequest) *PRSuite {
	return &PRSuite{
		PR:     PR,
		Labels: []string{"conformance-product-submission"},

		MetadataFolder: path.Join(os.Getenv("KO_DATA_PATH"), "conformance-testdata"),
		buffer:         *bytes.NewBuffer(nil),
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

func (s *PRSuite) SetMetadataFolder(path string) *PRSuite {
	s.MetadataFolder = path
	return s
}

func (s *PRSuite) thePRTitleIsNotEmpty() error {
	if len(s.PR.Title) == 0 {
		return fmt.Errorf("title is empty")
	}
	return nil
}

func (s *PRSuite) isIncludedInItsFileList(fileName string) error {
	foundFile := false
	for _, f := range s.PR.SupportingFiles {
		if strings.ToLower(f.BaseName) == strings.ToLower(fileName) {
			foundFile = true
			break
		}
	}
	if foundFile == false {
		s.Labels = append(s.Labels, "missing-file-"+fileName)
		s.MissingFiles = append(s.MissingFiles, fileName)
		return fmt.Errorf("missing file '%v'", fileName)
	}
	return nil
}

func (s *PRSuite) fileFolderStructureMatchesRegex(match string) error {
	pattern := regexp.MustCompile(match)
	failureError := fmt.Errorf("your product submission PR be in folders like [KubernetesReleaseVersion]/[ProductName], e.g: v1.23/averycooldistro")
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
	paths := []string{}
	for _, file := range s.PR.SupportingFiles {
		filePath := path.Dir(file.Name)
		if filePath == "." {
			continue
		}
		foundInPaths := false
		for _, p := range paths {
			if p == filePath {
				foundInPaths = true
			}
		}
		if foundInPaths == false {
			paths = append(paths, filePath)
		}
	}
	if len(paths) != 1 {
		return fmt.Errorf("there should be a single set of products in the submission. We found %v. %v", len(paths), strings.Join(paths, ", "))
	}

	return nil
}

func (s *PRSuite) theTitleOfThePR() error {
	if s.PR.Title == "" {
		return fmt.Errorf("title is empty")
	}
	return nil
}

func (s *PRSuite) theTitleOfThePRMatches(match string) error {
	pattern := regexp.MustCompile(match)
	if pattern.MatchString(string(s.PR.Title)) != true {
		return fmt.Errorf("title must be formatted like 'Conformance results for [KubernetesReleaseVersion]/[ProductName]' (e.g: Conformance results for v1.23/CoolKubernetes)")
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
	file := s.GetFileByFileName(fileName)
	if file == nil {
		return fmt.Errorf("missing required file '%v'", fileName)
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
		return fmt.Errorf("missing required file '%v'", fileName)
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
	var matchingLine string
lineLoop:
	for _, line := range lines {
		if pattern.MatchString(line) == true {
			matchingLine = line
			break lineLoop
		}
	}
	if matchingLine == "" {
		return fmt.Errorf("the file '%v' does not contain a release version of Kubernetes in it", fileName)
	}
	allIndexes := pattern.FindAllSubmatchIndex([]byte(matchingLine), -1)
	for _, loc := range allIndexes {
		e2eLogKubernetesReleaseVersion := string(matchingLine[loc[2]:loc[3]])
		if e2eLogKubernetesReleaseVersion == "" {
			continue
		}
		s.E2eLogKubernetesReleaseVersion = e2eLogKubernetesReleaseVersion
		break
	}
	return nil
}

func (s *PRSuite) thatVersionMatchesTheSameKubernetesReleaseVersionAsInTheFolderStructure() error {
	e2elogVersion, err := semver.NewSemver(s.E2eLogKubernetesReleaseVersion)
	if err != nil {
		return err
	}
	e2elogVersionSegments := e2elogVersion.Segments()
	releaseVersion, err := semver.NewSemver(s.KubernetesReleaseVersion)
	if err != nil {
		return err
	}
	releaseVersionSegements := releaseVersion.Segments()
	fmt.Println("e2elog version", s.E2eLogKubernetesReleaseVersion, s.KubernetesReleaseVersion)
	if !(e2elogVersionSegments[0] == releaseVersionSegements[0] ||
		e2elogVersionSegments[1] == releaseVersionSegements[1]) {
		return fmt.Errorf("the Kubernetes release version in file 'e2e.log' (%v) doesn't match the same version in the folder structure (%v)", s.E2eLogKubernetesReleaseVersion, s.KubernetesReleaseVersion)
	}
	return nil
}

func (s *PRSuite) aListOfLabelsInThePR() error {
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
		return fmt.Errorf("required label '%v' not found", labelWithReleaseAttached)
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
		return fmt.Errorf("URL field '%v' in PRODUCT.yaml resolving content type '%v' must be (%v)", field, s.PR.ProductYAMLURLDataTypes[field], strings.Join(strings.Split(dataType, " "), ", or "))
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
		return fmt.Errorf("Kubernetes release version in the title (%v) and folder structure (%v) don't match", titleReleaseVersion, s.KubernetesReleaseVersion)
	}
	return nil
}

func (s *PRSuite) theReleaseVersion() error {
	if s.KubernetesReleaseVersion == "" {
		return fmt.Errorf("unable to find a Kubernetes release version in the title")
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
	latestVersionSegments[1] -= lastSupportingVersions
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

func (s *PRSuite) allRequiredTestsInJunitXmlArePresent() error {
	missingTests, err := s.GetMissingJunitTestsFromPRSuite()
	if err != nil {
		return err
	}
	if len(missingTests) > 0 {
		s.Labels = append(s.Labels, "required-tests-missing")
		sort.Strings(missingTests)
		return fmt.Errorf("the following test(s) are missing: \n    - %v", strings.Join(missingTests, "\n    - "))
	}
	s.Labels = append(s.Labels, "tests-verified-"+s.KubernetesReleaseVersion)
	return nil
}

type E2eLogTestPass struct {
	Message   string `json:"msg"`
	Total     int    `json:"total"`
	Completed int    `json:"completed"`
	Skipped   int    `json:"skipped"`
	Failed    int    `json:"failed"`
}

func (s *PRSuite) collectPassedTestsFromE2elog() (tests []string, err error) {
	file := s.GetFileByFileName("e2e.log")
	if file == nil {
		return []string{}, fmt.Errorf("unable to find file e2e.log")
	}
	fileLines := strings.Split(file.Contents, "\n")
	for _, line := range fileLines {
		if strings.Contains(line, "msg") == false {
			continue
		}
		line = strings.ReplaceAll(line, "•", "")
		var e2eLogTestPass E2eLogTestPass
		err = json.Unmarshal([]byte(line), &e2eLogTestPass)
		if err != nil {
			continue
		}
		if !(strings.Contains(e2eLogTestPass.Message, "PASSED") == true ||
			strings.Contains(e2eLogTestPass.Message, "[Conformance]") == true) {
			continue
		}
		tests = append(tests, strings.ReplaceAll(e2eLogTestPass.Message, "PASSED ", ""))
	}
	return tests, nil
}

func (s *PRSuite) theTestsPassAndAreSuccessful() error {
	success, passed, err := s.DetermineE2eLogSucessful()
	if err != nil {
		return err
	}
	if success == false {
		s.Labels = append(s.Labels, "evidence-missing")
		return fmt.Errorf("it appears that there failures in the e2e.log")
	}
	tests, err := s.collectPassedTestsFromE2elog()
	if err != nil {
		return err
	}
	if passed != len(tests) {
		return fmt.Errorf("it appears that the amount of tests in e2e.log (%v) don't match the amount of tests passed (%v)", passed, len(tests))
	}
	s.Labels = append(s.Labels, "no-failed-tests-"+s.KubernetesReleaseVersion)
	return nil
}

func (s *PRSuite) allRequiredTestsInE2eLogArePresent() error {
	e2etests, err := s.collectPassedTestsFromE2elog()
	if err != nil {
		return err
	}
	requiredTests, err := s.GetRequiredTests()
	if err != nil {
		return err
	}

	for _, submittedTest := range e2etests {
		if _, found := requiredTests[submittedTest]; found != true {
			continue
		}
		requiredTests[submittedTest] = true
	}
	missingTests := []string{}
	for test, found := range requiredTests {
		if found == true {
			continue
		}
		missingTests = append(missingTests, test)
	}
	if len(missingTests) > 0 {
		return fmt.Errorf("there appears to be %v tests missing from e2e.log: \n    - %v", len(missingTests), strings.Join(missingTests, "\n    - "))
	}
	return nil
}

func (s *PRSuite) theTestsMatch() error {
	e2eTests, err := s.collectPassedTestsFromE2elog()
	if err != nil {
		return err
	}
	junitTests, err := s.GetJunitSubmittedConformanceTests()
	if err != nil {
		return err
	}
	if len(junitTests) != len(e2eTests) {
		return fmt.Errorf("the amount of tests in e2e.log (%v) doesn't match the amount of tests in junit_01.xml (%v)", len(junitTests), len(e2eTests))
	}
	missingTests := []string{}
	for _, e2eTest := range e2eTests {
		foundInJunitTests := false
		for _, junitTest := range junitTests {
			if e2eTest == junitTest {
				foundInJunitTests = true
			}
		}
		if foundInJunitTests != true {
			missingTests = append(missingTests, e2eTest)
		}
	}
	if len(missingTests) > 0 {
		return fmt.Errorf("there appears to be %v tests in e2e.log that aren't in junit_01.xml: \n    - %v", len(missingTests), strings.Join(missingTests, "\n    - "))
	}
	return nil
}

func (s *PRSuite) GetLabelsAndCommentsFromSuiteResultsBuffer() (comment string, labels []string, err error) {
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
					if r.Name == strings.TrimSpace(e.Description) {
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
	// TODO use prSuite.Labels
	if s.KubernetesReleaseVersion != "" {
		s.Labels = append(s.Labels, "release-"+s.KubernetesReleaseVersion)
	}
	if len(resultPrepares) > 0 {
		finalComment = fmt.Sprintf("%v of %v requirements have passed. Please review the following:", len(uniquelyNamedStepsRun)-len(resultPrepares), len(uniquelyNamedStepsRun))
		for _, r := range resultPrepares {
			finalComment += "\n- [FAIL] " + r.Name
			for _, h := range r.Hints {
				finalComment += "\n  - " + h
			}
		}
		finalComment += "\n\n for a full list of requirements, please refer to the [_content of the PR_ section of the docs](https://github.com/cncf/k8s-conformance/blob/master/instructions.md#contents-of-the-pr)."
		s.Labels = append(s.Labels, "not-verifiable")
	} else {
		s.Labels = append(s.Labels, "release-documents-checked")
	}
	finalComment += "\n"

	return finalComment, s.Labels, nil
}

func (s *PRSuite) InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^the PR title is not empty$`, s.thePRTitleIsNotEmpty)
	ctx.Step(`^"([^"]*)" is included in its file list$`, s.isIncludedInItsFileList)
	ctx.Step(`^the files in the PR`, s.theFilesInThePR)
	ctx.Step(`^file folder structure matches "([^"]*)"$`, s.fileFolderStructureMatchesRegex)
	ctx.Step(`^the title of the PR$`, s.theTitleOfThePR)
	ctx.Step(`^the title of the PR matches "([^"]*)"$`, s.theTitleOfThePRMatches)
	ctx.Step(`^the yaml file "([^"]*)" contains the required and non-empty "([^"]*)"$`, s.theYamlFileContainsTheRequiredAndNonEmptyField)
	ctx.Step(`^a[n]? "([^"]*)" file$`, s.aFile)
	ctx.Step(`^"([^"]*)" is not empty$`, s.isNotEmpty)
	ctx.Step(`^a line of the file "([^"]*)" matches "([^"]*)"$`, s.aLineOfTheFileMatches)
	ctx.Step(`^a list of labels in the PR$`, s.aListOfLabelsInThePR)
	ctx.Step(`^the label prefixed with "([^"]*)" and ending with Kubernetes release version should be present$`, s.theLabelPrefixedWithAndEndingWithKubernetesReleaseVersionShouldBePresent)
	ctx.Step(`^if "([^"]*)" is set to url, the content of the url in the value of "([^"]*)" matches it\'s "([^"]*)"$`, s.ifIsSetToUrlTheContentOfTheUrlInTheValueOfMatchesIts)
	ctx.Step(`^there is only one path of folders$`, s.thereIsOnlyOnePathOfFolders)
	ctx.Step(`^the release version matches the release version in the title$`, s.theReleaseVersionMatchesTheReleaseVersionInTheTitle)
	ctx.Step(`^the release version$`, s.theReleaseVersion)
	ctx.Step(`^it is a valid and supported release$`, s.itIsAValidAndSupportedRelease)
	ctx.Step(`^the tests pass and are successful$`, s.theTestsPassAndAreSuccessful)
	ctx.Step(`^that version matches the same Kubernetes release version as in the folder structure$`, s.thatVersionMatchesTheSameKubernetesReleaseVersionAsInTheFolderStructure)
	ctx.Step(`^all required tests in junit_01.xml are present$`, s.allRequiredTestsInJunitXmlArePresent)
	ctx.Step(`^all required tests in e2e.log are present$`, s.allRequiredTestsInE2eLogArePresent)
	ctx.Step(`^the tests match$`, s.theTestsMatch)
}
