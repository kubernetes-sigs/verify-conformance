package suite

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/cucumber/godog"
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

func GetPRs() []PullRequest {
	return []PullRequest{
		{
			PullRequestQuery: PullRequestQuery{
				Title:  "Conformance results for v1.23 Cool (passing but without labels yet)",
				Number: 1,
			},
			Labels: []string{},
			ProductYAMLURLDataTypes: map[string]string{
				"vendor":            "string",
				"name":              "string",
				"version":           "string",
				"type":              "string",
				"description":       "string",
				"website_url":       "text/html",
				"repo_url":          "text/html",
				"documentation_url": "text/html",
				"product_logo_url":  "image/svg",
			},
			SupportingFiles: []*PullRequestFile{
				&PullRequestFile{
					Name:     "v1.23/cool/README.md",
					BaseName: "README.md",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.23/cke/README.md",
					Contents: `# Conformance test for Cool`,
				},
				&PullRequestFile{
					Name:     "v1.23/cool/PRODUCT.yaml",
					BaseName: "PRODUCT.yaml",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.23/cke/PRODUCT.yaml",
					Contents: `
vendor: Cool
name: cOOL - A Cool Kubernetes Engine
version: v1.23.3
website_url: https://cool.kubernetes/engine
repo_url: https://github.com/cool/kubernetes-engine
documentation_url: https://github.com/cool/kubernetes-engine
product_logo_url: https://github.com/cybozu-go/cke/blob/main/logo/cybozu_logo.svg
type: Installer
description: Cool Kubernetes Engine, a distributed service that automates Kubernetes cluster management.
`,
				},
				&PullRequestFile{
					Name:     "v1.23/cool/junit_01.xml",
					BaseName: "junit_01.xml",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.23/cke/junit_01.xml",
					Contents: `
<?xml version="1.0" encoding="UTF-8"?>
  <testsuite name="Kubernetes e2e suite" tests="311" failures="0" errors="0" time="5121.343">
      <testcase name="[sig-storage] In-tree Volumes [Driver: local][LocalVolumeType: dir-link] [Testpattern: Dynamic PV (block volmode)] multiVolume [Slow] should access to two volumes with the same volume mode and retain data across pod recreation on different node [LinuxOnly]" classname="Kubernetes e2e suite" time="0">
          <skipped></skipped>
      </testcase>
      <testcase name="[sig-auth] PodSecurityPolicy [Feature:PodSecurityPolicy] should forbid pod creation when no PSP is available" classname="Kubernetes e2e suite" time="0">
          <skipped></skipped>
      </testcase>
      <testcase name="[sig-storage] In-tree Volumes [Driver: ceph][Feature:Volumes][Serial] [Testpattern: Dynamic PV (default fs)] subPath should support existing single file [LinuxOnly]" classname="Kubernetes e2e suite" time="0">
          <skipped></skipped>
      </testcase>
  </testsuite>
`,
				},
				&PullRequestFile{
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.23/cke/e2e.log",
					BaseName: "e2e.log",
					Name:     "v1.23/cool/e2e.log",
					Contents: `
May 27 04:41:36.616: INFO: 3 / 3 pods ready in namespace 'kube-system' in daemonset 'node-dns' (0 seconds elapsed)
May 27 04:41:36.616: INFO: e2e test version: v1.23.4
May 27 04:41:36.617: INFO: kube-apiserver version: v1.23.4
May 27 04:41:36.617: INFO: >>> kubeConfig: /tmp/kubeconfig-441052555
May 27 04:41:36.620: INFO: Cluster IP family: ipv4
SSSSS
`,
				},
			},
		},
		{
			PullRequestQuery: PullRequestQuery{
				Title:  "Conformance results for v1.23 Something (Passing completely)",
				Number: 2,
			},
			Labels: []string{
				"no-failed-tests-v1.23",
				"release-documents-checked",
				"release-v1.23",
				"tests-verified-v1.23",
			},
			ProductYAMLURLDataTypes: map[string]string{
				"vendor":            "string",
				"name":              "string",
				"version":           "string",
				"type":              "string",
				"description":       "string",
				"website_url":       "text/html",
				"repo_url":          "text/html",
				"documentation_url": "text/html",
				"product_logo_url":  "application/postscript",
			},
			SupportingFiles: []*PullRequestFile{
				&PullRequestFile{
					Name:     "v1.23/cool/README.md",
					BaseName: "README.md",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.23/cke/README.md",
					Contents: `# Conformance test for Something`,
				},
				&PullRequestFile{
					Name:     "v1.23/cool/PRODUCT.yaml",
					BaseName: "PRODUCT.yaml",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.23/cke/PRODUCT.yaml",
					Contents: `
vendor: Something
name: something - A Cool Kubernetes Engine
version: v1.23.3
website_url: https://something.kubernetes/engine
repo_url: https://github.com/something/kubernetes-engine
documentation_url: https://github.com/something/kubernetes-engine
product_logo_url: https://github.com/cybozu-go/cke/blob/main/logo/cybozu_logo.svg
type: Installer
description: Something Kubernetes Engine, a distributed service that automates Kubernetes cluster management.
`,
				},
				&PullRequestFile{
					Name:     "v1.23/cool/junit_01.xml",
					BaseName: "junit_01.xml",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.23/cke/junit_01.xml",
					Contents: `
<?xml version="1.0" encoding="UTF-8"?>
  <testsuite name="Kubernetes e2e suite" tests="311" failures="0" errors="0" time="5121.343">
      <testcase name="[sig-storage] In-tree Volumes [Driver: local][LocalVolumeType: dir-link] [Testpattern: Dynamic PV (block volmode)] multiVolume [Slow] should access to two volumes with the same volume mode and retain data across pod recreation on different node [LinuxOnly]" classname="Kubernetes e2e suite" time="0">
          <skipped></skipped>
      </testcase>
      <testcase name="[sig-auth] PodSecurityPolicy [Feature:PodSecurityPolicy] should forbid pod creation when no PSP is available" classname="Kubernetes e2e suite" time="0">
          <skipped></skipped>
      </testcase>
      <testcase name="[sig-storage] In-tree Volumes [Driver: ceph][Feature:Volumes][Serial] [Testpattern: Dynamic PV (default fs)] subPath should support existing single file [LinuxOnly]" classname="Kubernetes e2e suite" time="0">
          <skipped></skipped>
      </testcase>
  </testsuite>
`,
				},
				&PullRequestFile{
					Name:     "v1.23/cool/e2e.log",
					BaseName: "e2e.log",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.23/cke/e2e.log",
					Contents: `
May 27 04:41:36.616: INFO: 3 / 3 pods ready in namespace 'kube-system' in daemonset 'node-dns' (0 seconds elapsed)
May 27 04:41:36.616: INFO: e2e test version: v1.23.4
May 27 04:41:36.617: INFO: kube-apiserver version: v1.23.4
May 27 04:41:36.617: INFO: >>> kubeConfig: /tmp/kubeconfig-441052555
May 27 04:41:36.620: INFO: Cluster IP family: ipv4
SSSSS
`,
				},
			},
		},
		{
			PullRequestQuery: PullRequestQuery{
				Title:  "Conformance results for v1.23 SomethingTheSequel (Passing but missing a no-tests-failed label)",
				Number: 2,
			},
			Labels: []string{"release-documents-checked", "release-v1.23", "tests-verified-v1.23"},
			ProductYAMLURLDataTypes: map[string]string{
				"vendor":            "string",
				"name":              "string",
				"version":           "string",
				"type":              "string",
				"description":       "string",
				"website_url":       "text/html",
				"repo_url":          "text/html",
				"documentation_url": "text/html",
				"product_logo_url":  "image/svg",
			},
			SupportingFiles: []*PullRequestFile{
				&PullRequestFile{
					Name:     "v1.23/cool/README.md",
					BaseName: "README.md",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.23/cke/README.md",
					Contents: `# Conformance test for Something`,
				},
				&PullRequestFile{
					Name:     "v1.23/cool/PRODUCT.yaml",
					BaseName: "PRODUCT.yaml",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.23/cke/PRODUCT.yaml",
					Contents: `
vendor: Something
name: something - A Cool Kubernetes Engine
version: v1.23.3
website_url: https://something.kubernetes/engine
repo_url: https://github.com/something/kubernetes-engine
documentation_url: https://github.com/something/kubernetes-engine
product_logo_url: https://github.com/cybozu-go/cke/blob/main/logo/cybozu_logo.svg
type: Installer
description: Something Kubernetes Engine, a distributed service that automates Kubernetes cluster management.
`,
				},
				&PullRequestFile{
					Name:     "v1.23/cool/junit_01.xml",
					BaseName: "junit_01.xml",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.23/cke/junit_01.xml",
					Contents: `
<?xml version="1.0" encoding="UTF-8"?>
  <testsuite name="Kubernetes e2e suite" tests="311" failures="0" errors="0" time="5121.343">
      <testcase name="[sig-storage] In-tree Volumes [Driver: local][LocalVolumeType: dir-link] [Testpattern: Dynamic PV (block volmode)] multiVolume [Slow] should access to two volumes with the same volume mode and retain data across pod recreation on different node [LinuxOnly]" classname="Kubernetes e2e suite" time="0">
          <skipped></skipped>
      </testcase>
      <testcase name="[sig-auth] PodSecurityPolicy [Feature:PodSecurityPolicy] should forbid pod creation when no PSP is available" classname="Kubernetes e2e suite" time="0">
          <skipped></skipped>
      </testcase>
      <testcase name="[sig-storage] In-tree Volumes [Driver: ceph][Feature:Volumes][Serial] [Testpattern: Dynamic PV (default fs)] subPath should support existing single file [LinuxOnly]" classname="Kubernetes e2e suite" time="0">
          <skipped></skipped>
      </testcase>
  </testsuite>
`,
				},
				&PullRequestFile{
					Name:     "v1.23/cool/e2e.log",
					BaseName: "e2e.log",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.23/cke/e2e.log",
					Contents: `
May 27 04:41:36.616: INFO: 3 / 3 pods ready in namespace 'kube-system' in daemonset 'node-dns' (0 seconds elapsed)
May 27 04:41:36.616: INFO: e2e test version: v1.23.4
May 27 04:41:36.617: INFO: kube-apiserver version: v1.23.4
May 27 04:41:36.617: INFO: >>> kubeConfig: /tmp/kubeconfig-441052555
May 27 04:41:36.620: INFO: Cluster IP family: ipv4
SSSSS
`,
				},
			},
		},
		{
			PullRequestQuery: PullRequestQuery{
				Title:  "Conformance results for Failurernetes (Failing at pretty much everything)",
				Number: 3,
			},
			Labels: []string{"release-documents-checked", "release-v1.23", "required-tests-missing"},
			ProductYAMLURLDataTypes: map[string]string{
				"vendor":            "string",
				"name":              "string",
				"version":           "string",
				"type":              "string",
				"description":       "string",
				"website_url":       "",
				"repo_url":          "",
				"documentation_url": "application/json",
				"product_logo_url":  "image/gif",
			},
			SupportingFiles: []*PullRequestFile{
				&PullRequestFile{
					Name:     "v1.23/cool-metal/PRODUCT.yaml",
					BaseName: "PRODUCT.yaml",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.23/cke/PRODUCT.yaml",
					Contents: `
vendor: Something
name: something - A Cool Kubernetes Engine
version: v1.23.3
website_url: https://something.kubernetes/engine
repo_url: https://github.com/something/kubernetes-engine
product_logo_url: https://github.com/cybozu-go/cke/blob/main/logo/cybozu_logo.svg
type: Installer
description: Something Kubernetes Engine, a distributed service that automates Kubernetes cluster management.
`,
				},
				&PullRequestFile{
					Name:     "v1.23/cool-metal/junit_01.xml",
					BaseName: "junit_01.xml",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.23/cke/junit_01.xml",
					Contents: ``,
				},
				&PullRequestFile{
					Name:     "v1.23/cool-metal/e2e.log",
					BaseName: "e2e.log",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.23/cke/e2e.log",
					Contents: `
May 27 04:41:36.616: INFO: 3 / 3 pods ready in namespace 'kube-system' in daemonset 'node-dns' (0 seconds elapsed)
May 27 04:41:36.616: INFO: e2e test version: v2
May 27 04:41:36.617: INFO: kube-apiserver version: v2
May 27 04:41:36.617: INFO: >>> kubeConfig: /tmp/kubeconfig-441052555
May 27 04:41:36.620: INFO: Cluster IP family: ipv4
SSSSS
`,
				},
				&PullRequestFile{
					Name:     "v1.23/cool/PRODUCT.yaml",
					BaseName: "PRODUCT.yaml",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.23/cke/PRODUCT.yaml",
					Contents: `
vendor: Something
name: something - A Cool Kubernetes Engine
version: v1.23.3
website_url: https://something.kubernetes/engine
repo_url: https://github.com/something/kubernetes-engine
product_logo_url: https://github.com/cybozu-go/cke/blob/main/logo/cybozu_logo.svg
type: Installer
description: Something Kubernetes Engine, a distributed service that automates Kubernetes cluster management.
`,
				},
				&PullRequestFile{
					Name:     "v1.23/cool/junit_01.xml",
					BaseName: "junit_01.xml",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.23/cke/junit_01.xml",
					Contents: ``,
				},
				&PullRequestFile{
					Name:     "v1.23/cool/e2e.log",
					BaseName: "e2e.log",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.23/cke/e2e.log",
					Contents: `
May 27 04:41:36.616: INFO: 3 / 3 pods ready in namespace 'kube-system' in daemonset 'node-dns' (0 seconds elapsed)
May 27 04:41:36.616: INFO: e2e test version: v2
May 27 04:41:36.617: INFO: kube-apiserver version: v2
May 27 04:41:36.617: INFO: >>> kubeConfig: /tmp/kubeconfig-441052555
May 27 04:41:36.620: INFO: Cluster IP family: ipv4
SSSSS
`,
				},
				&PullRequestFile{
					Name:     "recipe.org",
					BaseName: "recipe.org",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/recipe.org",
					Contents: `
* How to cook pasta

1. bring water to a boil
2. add some oil and salt to the water
3. add pasta to the pot
4. once pasta is aldente, remove from heat
5. drain water from the pot`,
				},
			},
		},
	}
}

type PRSuite struct {
	PR                       *PullRequest
	KubernetesReleaseVersion string
	ProductName              string

	Suite  godog.TestSuite
	buffer bytes.Buffer
}

func NewPRSuite(PR *PullRequest) *PRSuite {
	return &PRSuite{
		PR: PR,

		buffer: *bytes.NewBuffer(nil),
	}
}

func (s *PRSuite) NewTestSuite() godog.TestSuite {
	s.Suite = godog.TestSuite{
		Name: "how-are-the-prs",
		Options: &godog.Options{
			// Format: "pretty",
			Format: "cucumber",
			Output: &s.buffer,
		},
		ScenarioInitializer: s.InitializeScenario,
	}
	return s.Suite
}

func (s *PRSuite) aConformanceProductSubmissionPR() error {
	if s.PR == nil {
		return fmt.Errorf("unable to find PR from query")
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
	return fmt.Errorf("missing file '%v'", file)
}

func (s *PRSuite) fileFolderStructureMatchesRegex(match string) error {
	pattern := regexp.MustCompile(match)

	failureError := fmt.Errorf("the content structure of your product submission PR must match '%v' (KubernetesReleaseVersion/ProductName, e.g: v1.23/averycooldistro)", match)
	for _, file := range s.PR.SupportingFiles {
		if matches := pattern.MatchString(file.Name); matches != true {
			return fmt.Errorf("file '%v' not allowed. %v", file.Name, failureError)
		}
		allIndexes := pattern.FindAllSubmatchIndex([]byte(file.Name), -1)
		for _, loc := range allIndexes {
			baseFolder := string(file.Name[loc[2]:loc[3]])
			distroName := string(file.Name[loc[4]:loc[5]])

			if baseFolder == "" || distroName == "" {
				return fmt.Errorf("the content structure of your product submission PR must match '%v' (KubernetesReleaseVersion/ProductName, e.g: v1.23/averycooldistro)", match)
			}
		}
	}
	return nil
}

func (s *PRSuite) thereIsOnlyOnePathOfFolders() error {
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
		fullFilePath := path.Join(s.KubernetesReleaseVersion, s.ProductName, fileName)
		if f.Name == fullFilePath {
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
		if s.PR.ProductYAMLURLDataTypes[field] == dt {
			foundDataType = true
		}
	}
	if foundDataType == false {
		return fmt.Errorf("unable to use field '%v' in PRODUCT.yaml as the data in the resolved content doesn't match what is expected (%v)", field, dataType)
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
				resultPrepare.Name = e.Name
				resultPrepares = append(resultPrepares, resultPrepare)
			}
		}
	}

	finalComment := fmt.Sprintf("All requirements (%v) have passed for the submission!", len(uniquelyNamedStepsRun))
	labels = []string{}
	if s.KubernetesReleaseVersion != "" {
		labels = []string{"release-" + s.KubernetesReleaseVersion}
	}
	if len(resultPrepares) > 0 {
		finalComment = "Some requirements have not passed:"
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
}
