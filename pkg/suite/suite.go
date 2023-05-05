package suite

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/mail"
	"net/url"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
	semver "github.com/hashicorp/go-version"
	githubql "github.com/shurcooL/githubv4"
	sonobuoyresults "github.com/vmware-tanzu/sonobuoy/pkg/client/results"
	"sigs.k8s.io/yaml"

	"cncf.io/infra/verify-conformance-release/internal/types"
	"cncf.io/infra/verify-conformance-release/pkg/common"
)

// TODO ensure file checking

var (
	lastSupportingVersions = 2
)

type ResultPrepare struct {
	Name  string
	Hints []string
}

type PullRequestQuery struct {
	Number     githubql.Int
	HeadRefOID githubql.String
	Author     struct {
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
				Oid    githubql.String
				Status struct {
					Contexts []struct {
						Context githubql.String
						State   githubql.String
					}
				}
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

type E2eLogTestPass struct {
	Message   string `json:"msg"`
	Total     int    `json:"total"`
	Completed int    `json:"completed"`
	Skipped   int    `json:"skipped"`
	Failed    int    `json:"failed"`
}

type JunitTestCase struct {
	XMLName xml.Name  `xml:"testcase"`
	Name    string    `xml:"name,attr"`
	Skipped *struct{} `xml:"skipped"`
}

type JunitTestSuite struct {
	TestSuite []JunitTestCase `xml:"testcase"`
}

type JunitTestSuitev125 struct {
	Name      string          `xml:"name,attr"`
	Package   string          `xml:"package,attr"`
	Tests     int             `xml:"tests,attr"`
	Disabled  int             `xml:"xml,attr"`
	Errors    int             `xml:"errors,attr"`
	Failures  int             `xml:"failures,attr"`
	Time      string          `xml:"time,attr"`
	Timestamp string          `xml:"timestamp,attr"`
	TestCase  []JunitTestCase `xml:"testcase"`
}

type JunitTestSuitesv125 struct {
	XMLName   xml.Name           `xml:"testsuites"`
	Tests     int                `xml:"tests,attr"`
	Disabled  int                `xml:"xml,attr"`
	Errors    int                `xml:"errors,attr"`
	Failures  int                `xml:"failures,attr"`
	Time      float64            `xml:"time,attr"`
	TestSuite JunitTestSuitev125 `xml:"testsuite"`
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
			// TODO: add tags filtering
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
		return common.SafeError(fmt.Errorf("title is empty"))
	}
	return nil
}

func (s *PRSuite) isIncludedInItsFileList(fileName string) error {
	foundFile := false
	for _, f := range s.PR.SupportingFiles {
		if strings.EqualFold(f.BaseName, fileName) {
			foundFile = true
			break
		}
	}
	if !foundFile {
		s.Labels = append(s.Labels, "missing-file-"+fileName)
		s.MissingFiles = append(s.MissingFiles, fileName)
		return common.SafeError(fmt.Errorf("missing file '%v'", fileName))
	}
	return nil
}

func (s *PRSuite) fileFolderStructureMatchesRegex(match string) error {
	pattern := regexp.MustCompile(match)
	failureError := fmt.Errorf("your product submission PR must be in folders structured like [KubernetesReleaseVersion]/[ProductName], e.g: v1.23/averycooldistro")
	for _, file := range s.PR.SupportingFiles {
		if matches := pattern.MatchString(path.Dir(file.Name)); !matches {
			return common.SafeError(fmt.Errorf("file '%v' not allowed. %v", file.Name, failureError))
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
		if !foundInPaths {
			paths = append(paths, filePath)
		}
	}
	if len(paths) != 1 {
		return common.SafeError(fmt.Errorf("there should be a single set of products in the submission. We found %v product submissions: %v", len(paths), strings.Join(paths, ", ")))
	}

	return nil
}

func (s *PRSuite) theTitleOfThePR() error {
	if s.PR.Title == "" {
		return common.SafeError(fmt.Errorf("title is empty"))
	}
	return nil
}

func (s *PRSuite) theTitleOfThePRMatches(match string) error {
	pattern := regexp.MustCompile(match)
	if !pattern.MatchString(string(s.PR.Title)) {
		return common.SafeError(fmt.Errorf("title must be formatted like 'Conformance results for [KubernetesReleaseVersion]/[ProductName]' (e.g: Conformance results for v1.23/CoolKubernetes)"))
	}
	return nil
}

func (s *PRSuite) theFilesInThePR() error {
	if len(s.PR.SupportingFiles) == 0 {
		return common.SafeError(fmt.Errorf("there were no files found in the submission"))
	}
	return nil
}

func (s *PRSuite) aFile(fileName string) error {
	file := s.GetFileByFileName(fileName)
	if file == nil {
		return common.SafeError(fmt.Errorf("missing required file '%v'", fileName))
	}
	return nil
}

func (s *PRSuite) GetFileByFileName(fileName string) *PullRequestFile {
	for _, f := range s.PR.SupportingFiles {
		if strings.EqualFold(f.BaseName, fileName) {
			return f
		}
	}
	return nil
}

func (s *PRSuite) theYamlFileContainsTheRequiredAndNonEmptyField(fileName, fieldName string) error {
	var parsedContent map[string]*interface{}
	file := s.GetFileByFileName(fileName)
	if file == nil {
		return common.SafeError(fmt.Errorf("missing required file '%v'", fileName))
	}
	err := yaml.Unmarshal([]byte(file.Contents), &parsedContent)
	if err != nil {
		return common.SafeError(fmt.Errorf("unable to read file '%v'", fileName))
	}
	if parsedContent[fieldName] == nil {
		return common.SafeError(fmt.Errorf("missing or empty field '%v' in file '%v'", fieldName, fileName))
	}
	return nil
}

func (s *PRSuite) isNotEmpty(fileName string) error {
	file := s.GetFileByFileName(fileName)
	if file == nil {
		return common.SafeError(fmt.Errorf("unable to find file '%v'", fileName))
	}
	if file.Contents == "" {
		return common.SafeError(fmt.Errorf("file '%v' is empty", fileName))
	}
	return nil
}

func (s *PRSuite) aLineOfTheFileMatches(fileName, match string) error {
	pattern := regexp.MustCompile(match)
	file := s.GetFileByFileName(fileName)
	if file == nil {
		return common.SafeError(fmt.Errorf("unable to find file '%v'", fileName))
	}
	lines := strings.Split(file.Contents, "\n")
	var matchingLine string
lineLoop:
	for _, line := range lines {
		if !pattern.MatchString(line) {
			matchingLine = line
			break lineLoop
		}
	}
	if matchingLine == "" {
		return common.SafeError(fmt.Errorf("the file '%v' does not contain a release version of Kubernetes in it", fileName))
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

func (s *PRSuite) aListOfCommits() error {
	if len(s.PR.Commits.Nodes) == 0 {
		return common.SafeError(fmt.Errorf("no commits were found"))
	}
	return nil
}

func (s *PRSuite) thereIsOnlyOneCommit() error {
	if len(s.PR.Commits.Nodes) > 1 {
		return common.SafeError(fmt.Errorf("more than one commit was found; only one commit is allowed."))
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
	if !(e2elogVersionSegments[0] == releaseVersionSegements[0] &&
		e2elogVersionSegments[1] == releaseVersionSegements[1]) {
		return common.SafeError(fmt.Errorf("the Kubernetes release version in file 'e2e.log' (%v) doesn't match the same version in the folder structure (%v)", s.E2eLogKubernetesReleaseVersion, s.KubernetesReleaseVersion))
	}
	return nil
}

func (s *PRSuite) aListOfLabelsInThePR() error {
	if len(s.PR.Labels) == 0 {
		return common.SafeError(fmt.Errorf("there are no labels found"))
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
	if !foundLabel {
		return common.SafeError(fmt.Errorf("required label '%v' not found", labelWithReleaseAttached))
	}
	return nil
}

func (s *PRSuite) theContentOfTheInTheValueOfIsAValid(fieldType string, field string) error {
	fileName := "PRODUCT.yaml"
	var parsedContent map[string]string
	file := s.GetFileByFileName(fileName)
	if file == nil {
		return common.SafeError(fmt.Errorf("missing required file '%v'", fileName))
	}
	err := yaml.Unmarshal([]byte(file.Contents), &parsedContent)
	if err != nil {
		return common.SafeError(fmt.Errorf("unable to read file '%v'", fileName))
	}
	if parsedContent[field] == "" {
		return nil
	}
	switch fieldType {
	case "URL":
		_, err = url.ParseRequestURI(parsedContent[field])
		if err != nil {
			return common.SafeError(fmt.Errorf("URL for field '%v' in PRODUCT.yaml is not a valid URL, %v", field, err))
		}
	case "email":
		_, err = mail.ParseAddress(parsedContent[field])
		if err != nil {
			return common.SafeError(fmt.Errorf("Email field '%v' in PRODUCT.yaml is not a valid address, %v", field, err))
		}
	}
	return nil
}

func (s *PRSuite) theContentOfTheUrlInTheValueOfMatches(field, dataType string) error {
	if s.PR.ProductYAMLURLDataTypes[field] == "" {
		return nil
	}
	foundDataType := false
	for _, dt := range strings.Split(dataType, " ") {
		foundDataType = strings.Contains(s.PR.ProductYAMLURLDataTypes[field], dt)
		if foundDataType {
			break
		}
	}
	if !foundDataType {
		return common.SafeError(fmt.Errorf("URL field '%v' in PRODUCT.yaml resolving content type '%v' must be (%v)", field, s.PR.ProductYAMLURLDataTypes[field], strings.Join(strings.Split(dataType, " "), ", or ")))
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
		return common.SafeError(fmt.Errorf("the Kubernetes release version in the title (%v) and folder structure (%v) don't match", titleReleaseVersion, s.KubernetesReleaseVersion))
	}
	return nil
}

func (s *PRSuite) theReleaseVersion() error {
	if s.KubernetesReleaseVersion == "" {
		return common.SafeError(fmt.Errorf("unable to find a Kubernetes release version in the title"))
	}
	return nil
}

func (s *PRSuite) itIsAValidAndSupportedRelease() error {
	latestVersion, err := semver.NewSemver(s.KubernetesReleaseVersionLatest)
	if err != nil {
		fmt.Printf("error with go-version parsing latestVersion '%v': %v\n", s.KubernetesReleaseVersionLatest, err)
		return common.SafeError(fmt.Errorf("unable to parse latest release version"))
	}
	currentVersion, err := semver.NewSemver(s.KubernetesReleaseVersion)
	if err != nil {
		fmt.Printf("error with go-version parsing currentVersion '%v': %v\n", currentVersion, err)
		return common.SafeError(fmt.Errorf("unable to parse latest release version"))
	}
	latestVersionSegments := latestVersion.Segments()
	latestVersionSegments[1] -= lastSupportingVersions
	oldestVersion := fmt.Sprintf("v%v.%v", latestVersionSegments[0], latestVersionSegments[1])
	oldestSupportedVersion, err := semver.NewSemver(oldestVersion)
	if err != nil {
		fmt.Printf("error with go-version parsing oldest release version '%v': %v\n", latestVersionSegments, err)
		return common.SafeError(fmt.Errorf("unable to parse oldest supported release version"))
	}

	if currentVersion.GreaterThan(latestVersion) {
		return common.SafeError(fmt.Errorf("unable to use version '%v' because it is newer than the current supported release (%v)", s.KubernetesReleaseVersion, s.KubernetesReleaseVersionLatest))
	} else if currentVersion.LessThan(oldestSupportedVersion) {
		return common.SafeError(fmt.Errorf("unable to use version '%v' because it is older than the last currently supported release (%v)", s.KubernetesReleaseVersion, oldestVersion))
	}
	return nil
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
			if versionSemver.GreaterThanOrEqual(testVersionSemver) {
				foundInTestVersions = true
			}
			if foundInTestVersions {
				break testSupportedVersions
			}
		}
		if !foundInTestVersions {
			continue
		}
		tests[test.Codename] = false
	}
	return tests, nil
}

func (s *PRSuite) getJunitSubmittedConformanceTests() (tests []sonobuoyresults.JUnitTestCase, err error) {
	file := s.GetFileByFileName("junit_01.xml")
	if file == nil {
		return []sonobuoyresults.JUnitTestCase{}, fmt.Errorf("unable to find file junit_01.xml")
	}
	version, err := semver.NewVersion(s.E2eLogKubernetesReleaseVersion)
	if err != nil {
		fmt.Printf("semver error: %#v", err)
		return []sonobuoyresults.JUnitTestCase{}, fmt.Errorf("unable to find target version for this submission")
	}
	constraint, _ := semver.NewConstraint(">=v1.25.0")
	if constraint.Check(version) {
		junit := sonobuoyresults.JUnitTestSuites{}
		if err := xml.Unmarshal([]byte(file.Contents), &junit); err != nil {
			return []sonobuoyresults.JUnitTestCase{}, common.SafeError(fmt.Errorf("unable to parse junit_01.xml file, %v", err))
		}
		for _, suite := range junit.Suites {
			for _, testcase := range suite.TestCases {
				if testcase.SkipMessage != nil {
					continue
				}
				if !strings.Contains(testcase.Name, "[Conformance]") {
					continue
				}
				testcase.Name = strings.Replace(testcase.Name, "&#39;", "'", -1)
				testcase.Name = strings.Replace(testcase.Name, "&#34;", "\"", -1)
				testcase.Name = strings.Replace(testcase.Name, "&gt;", ">", -1)
				testcase.Name = strings.Replace(testcase.Name, "'cat /tmp/health'", "\"cat /tmp/health\"", -1)
				tests = append(tests, testcase)
			}
		}
	} else {
		testSuite := JunitTestSuite{}
		if err := xml.Unmarshal([]byte(file.Contents), &testSuite); err != nil {
			return []sonobuoyresults.JUnitTestCase{}, common.SafeError(fmt.Errorf("unable to parse junit_01.xml file, %v", err))
		}
		for _, testcase := range testSuite.TestSuite {
			if testcase.Skipped != nil {
				continue
			}
			if !strings.Contains(testcase.Name, "[Conformance]") {
				continue
			}
			testcase.Name = strings.Replace(testcase.Name, "&#39;", "'", -1)
			testcase.Name = strings.Replace(testcase.Name, "&#34;", "\"", -1)
			testcase.Name = strings.Replace(testcase.Name, "&gt;", ">", -1)
			testcase.Name = strings.Replace(testcase.Name, "'cat /tmp/health'", "\"cat /tmp/health\"", -1)
			tests = append(tests, sonobuoyresults.JUnitTestCase{
				Name:    testcase.Name,
				XMLName: testcase.XMLName,
			})
		}
	}
	return tests, nil
}

func (s *PRSuite) GetJunitSubmittedConformanceTests() (tests []string, err error) {
	collectedTests, err := s.getJunitSubmittedConformanceTests()
	if err != nil {
		return []string{}, err
	}
	for _, t := range collectedTests {
		tests = append(tests, t.Name)
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
		submittedTest = strings.TrimPrefix(submittedTest, "[It] ")
		if _, found := requiredTests[submittedTest]; !found {
			continue
		}
		requiredTests[submittedTest] = true
	}
	for test, found := range requiredTests {
		if found {
			continue
		}
		missingTests = append(missingTests, test)
	}

	return missingTests, nil
}

func (s *PRSuite) determineSuccessfulTestsBelowv125() (success bool, passed int, err error) {
	file := s.GetFileByFileName("e2e.log")
	if file == nil {
		return false, 0, fmt.Errorf("unable to find file e2e.log")
	}
	fileLines := strings.Split(file.Contents, "\n")
	lastLinesAmount := len(fileLines) - 100
	if lastLinesAmount < 0 {
		lastLinesAmount = len(fileLines)
	}
	fileLast100Lines := fileLines[lastLinesAmount:]
	var pattern *regexp.Regexp
	patternComplete := regexp.MustCompile(`^(SUCCESS|FAIL)! -- ([1-9][0-9]+) Passed \| ([0-9]+) Failed \| ([0-9]+) Pending \| ([0-9]+) Skipped$`)
	patternCompleteWithFlaked := regexp.MustCompile(`^(SUCCESS|FAIL)! -- ([1-9][0-9]+) Passed \| ([0-9]+) Failed \| ([0-9]+) Flaked \| ([0-9]+) Pending \| ([0-9]+) Skipped$`)
	matchingLine := ""
	for _, line := range fileLast100Lines {
		if patternComplete.MatchString(line) {
			matchingLine = line
			pattern = patternComplete
		} else if patternCompleteWithFlaked.MatchString(line) {
			matchingLine = line
			pattern = patternCompleteWithFlaked
		}
	}
	if matchingLine == "" {
		return false, 0, fmt.Errorf("unable to determine test results (passed, failed, flaked, pending, skipped) from e2e.log")
	}
	allIndexes := pattern.FindAllSubmatchIndex([]byte(matchingLine), -1)
	for _, loc := range allIndexes {
		passed, err = strconv.Atoi(matchingLine[loc[4]:loc[5]])
		if err != nil {
			return false, 0, fmt.Errorf("failed to parse successful tests")
		}
		// failed := string(file.Name[loc[4]:loc[5]])
		// pending := string(file.Name[loc[6]:loc[7]])
		// skipped := string(file.Name[loc[8]:loc[9]])
	}
	return true, passed, nil
}

func (s *PRSuite) determineSuccessfulTestsv125AndAbove() (success bool, passed int, tests []string, err error) {
	junitTests, err := s.getJunitSubmittedConformanceTests()
	if err != nil {
		return false, 0, []string{}, err
	}
	hasFailure := false
	for _, t := range junitTests {
		if t.ErrorMessage != nil || t.Failure != nil {
			hasFailure = true
			continue
		}
		passed += 1
		testName := strings.TrimPrefix(t.Name, "[It] ")
		tests = append(tests, testName)
	}
	if hasFailure {
		return false, passed, tests, nil
	}
	return true, passed, tests, nil
}

func (s *PRSuite) DetermineSuccessfulTests() (success bool, passed int, tests []string, err error) {
	success, passed, tests, err = s.determineSuccessfulTestsv125AndAbove()
	if err != nil {
		return false, 0, []string{}, err
	}
	return success, passed, tests, nil
}

func (s *PRSuite) allRequiredTestsInJunitXmlArePresent() error {
	missingTests, err := s.GetMissingJunitTestsFromPRSuite()
	if err != nil {
		return err
	}
	if len(missingTests) > 0 {
		s.Labels = append(s.Labels, "required-tests-missing")
		sort.Strings(missingTests)
		return common.SafeError(fmt.Errorf("the following test(s) are missing: \n    - %v", strings.Join(missingTests, "\n    - ")))
	}
	s.Labels = append(s.Labels, "tests-verified-"+s.KubernetesReleaseVersion)
	return nil
}

func (s *PRSuite) collectPassedTestsFromE2elog() (tests []string, err error) {
	file := s.GetFileByFileName("e2e.log")
	if file == nil {
		return []string{}, fmt.Errorf("unable to find file e2e.log")
	}
	fileLines := strings.Split(file.Contents, "\n")
	for _, line := range fileLines {
		if !strings.Contains(line, "msg") {
			continue
		}
		line = strings.ReplaceAll(line, "•", "")
		var e2eLogTestPass E2eLogTestPass
		err = json.Unmarshal([]byte(line), &e2eLogTestPass)
		if err != nil {
			continue
		}
		if !(strings.Contains(e2eLogTestPass.Message, "PASSED") ||
			strings.Contains(e2eLogTestPass.Message, "[Conformance]")) {
			continue
		}
		tests = append(tests, strings.ReplaceAll(e2eLogTestPass.Message, "PASSED ", ""))
	}
	return tests, nil
}

func (s *PRSuite) theTestsPassAndAreSuccessful() error {
	success, _, _, err := s.DetermineSuccessfulTests()
	if err != nil {
		return err
	}
	if !success {
		s.Labels = append(s.Labels, "evidence-missing")
		return common.SafeError(fmt.Errorf("it appears that there are failures in some tests"))
	}
	s.Labels = append(s.Labels, "no-failed-tests-"+s.KubernetesReleaseVersion)
	return nil
}

func (s *PRSuite) allRequiredTestsInArePresent() error {
	var tests []string
	_, _, tests, err := s.DetermineSuccessfulTests()
	if err != nil {
		return err
	}
	requiredTests, err := s.GetRequiredTests()
	if err != nil {
		return err
	}

	for _, submittedTest := range tests {
		if _, found := requiredTests[submittedTest]; !found {
			continue
		}
		requiredTests[submittedTest] = true
	}
	missingTests := []string{}
	for test, found := range requiredTests {
		if found {
			continue
		}
		missingTests = append(missingTests, test)
	}
	if len(missingTests) > 0 {
		sort.Strings(missingTests)
		return common.SafeError(fmt.Errorf("there appears to be %v tests missing: \n    - %v", len(missingTests), strings.Join(missingTests, "\n    - ")))
	}
	return nil
}

func IsValidYaml(input []byte) error {
	var content map[string]interface{}
	err := yaml.Unmarshal(input, &content)
	if err != nil {
		return err
	}
	return nil
}

func (s *PRSuite) IsValid(fileName, fileType string) error {
	file := s.GetFileByFileName(fileName)
	if file == nil {
		return common.SafeError(fmt.Errorf("unable to find file '%v'", fileName))
	}
	if file.Contents == "" {
		return common.SafeError(fmt.Errorf("file '%v' is empty", fileName))
	}
	switch fileType {
	case "yaml":
		if err := IsValidYaml([]byte(file.Contents)); err != nil {
			return common.SafeError(fmt.Errorf("failed to parse (%v) YAML, %v", fileName, err))
		}
		// TODO: add xml parsing
	}
	return nil
}

func aPRTitle() error {
	return nil
}

func (s *PRSuite) GetLabelsAndCommentsFromSuiteResultsBuffer() (comment string, labels []string, state string, err error) {
	cukeFeatures := []types.CukeFeatureJSON{}
	err = json.Unmarshal(s.buffer.Bytes(), &cukeFeatures)
	if err != nil {
		return "", []string{}, "", err
	}
	releaseVersion, err := semver.NewSemver(s.KubernetesReleaseVersion)
	if err != nil {
		return "", []string{}, "", err
	}
	releaseVersionLatest, err := semver.NewSemver(s.KubernetesReleaseVersionLatest)
	if err != nil {
		return "", []string{}, "", err
	}
	if releaseVersion.GreaterThanOrEqual(releaseVersionLatest) {
		_, err = common.ReadFile(path.Join(s.MetadataFolder, s.KubernetesReleaseVersion, "conformance.yaml"))
		if err != nil {
			return fmt.Sprintf("The release version %v is unable to be processed at this time; Please wait as this version may become available soon.", s.KubernetesReleaseVersion), append(labels, "conformance-product-submission", "unable-to-process"), "pending", nil
		}
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
			if !foundNameInStepsRun {
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
				if !foundExistingResultTitle {
					resultPrepare.Hints = append(resultPrepare.Hints, hint)
				}
			}
			if hasFails && !foundExistingResultTitle {
				resultPrepare.Name = strings.TrimSpace(e.Description)
				resultPrepares = append(resultPrepares, resultPrepare)
			}
		}
	}

	finalComment := fmt.Sprintf("All requirements (%v) have passed for the submission!", len(uniquelyNamedStepsRun))
	state = "success"
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
		finalComment += "\n\n for a full list of requirements, please refer to these sections of the docs: [_content of the PR_](https://github.com/cncf/k8s-conformance/blob/master/instructions.md#contents-of-the-pr), and [_requirements_](https://github.com/cncf/k8s-conformance/blob/master/instructions.md#requirements)."
		s.Labels = append(s.Labels, "not-verifiable")
		state = "failure"
	} else {
		s.Labels = append(s.Labels, "release-documents-checked")
	}
	finalComment += "\n"

	return finalComment, s.Labels, state, nil
}

func (s *PRSuite) InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^the PR title is not empty$`, s.thePRTitleIsNotEmpty)
	ctx.Step(`^"([^"]*)" is included in its file list$`, s.isIncludedInItsFileList)
	ctx.Step(`^the files in the PR`, s.theFilesInThePR)
	ctx.Step(`^file folder structure matches "([^"]*)"$`, s.fileFolderStructureMatchesRegex)
	ctx.Step(`^the title of the PR$`, s.theTitleOfThePR)
	ctx.Step(`^the title of the PR matches "([^"]*)"$`, s.theTitleOfThePRMatches)
	ctx.Step(`^a[n]? "([^"]*)" file$`, s.aFile)
	ctx.Step(`^"([^"]*)" is not empty$`, s.isNotEmpty)
	ctx.Step(`^a line of the file "([^"]*)" matches "([^"]*)"$`, s.aLineOfTheFileMatches)
	ctx.Step(`^a list of labels in the PR$`, s.aListOfLabelsInThePR)
	ctx.Step(`^the label prefixed with "([^"]*)" and ending with Kubernetes release version should be present$`, s.theLabelPrefixedWithAndEndingWithKubernetesReleaseVersionShouldBePresent)
	ctx.Step(`^the yaml file "([^"]*)" contains the required and non-empty "([^"]*)"$`, s.theYamlFileContainsTheRequiredAndNonEmptyField)
	ctx.Step(`^the content of the "([^"]*)" in the value of "([^"]*)" is a valid .*$`, s.theContentOfTheInTheValueOfIsAValid)
	ctx.Step(`^the content of the url in the value of "([^"]*)" matches it\'s "([^"]*)"$`, s.theContentOfTheUrlInTheValueOfMatches)
	ctx.Step(`^there is only one path of folders$`, s.thereIsOnlyOnePathOfFolders)
	ctx.Step(`^the release version matches the release version in the title$`, s.theReleaseVersionMatchesTheReleaseVersionInTheTitle)
	ctx.Step(`^the release version$`, s.theReleaseVersion)
	ctx.Step(`^it is a valid and supported release$`, s.itIsAValidAndSupportedRelease)
	ctx.Step(`^the tests pass and are successful$`, s.theTestsPassAndAreSuccessful)
	ctx.Step(`^that version matches the same Kubernetes release version as in the folder structure$`, s.thatVersionMatchesTheSameKubernetesReleaseVersionAsInTheFolderStructure)
	ctx.Step(`^all required tests in junit_01.xml are present$`, s.allRequiredTestsInJunitXmlArePresent)
	ctx.Step(`^all required tests are present$`, s.allRequiredTestsInArePresent)
	ctx.Step(`^a PR title$`, aPRTitle)
	ctx.Step(`^"([^"]*)" is valid "([^"]*)"`, s.IsValid)
	ctx.Step(`^a list of commits$`, s.aListOfCommits)
	ctx.Step(`^there is only one commit$`, s.thereIsOnlyOneCommit)
}
