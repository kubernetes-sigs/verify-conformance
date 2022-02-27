package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cucumber/godog"
	githubql "github.com/shurcooL/githubv4"
	"sigs.k8s.io/yaml"
	// "k8s.io/test-infra/prow/github"
)

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
	Contents string
}

type PullRequest struct {
	PullRequestQuery

	Labels          []string
	SupportingFiles map[string]*PullRequestFile
}

func GetPRs() []PullRequest {
	return []PullRequest{
		{
			PullRequestQuery: PullRequestQuery{
				Title:  "Conformance results for v1.23 Cool (passing)",
				Number: 1,
			},
			Labels: []string{},
			SupportingFiles: map[string]*PullRequestFile{
				"README.md": &PullRequestFile{
					Name:     "v1.23/cool/README.md",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.23/cke/README.md",
					Contents: `# Conformance test for Cool`,
				},
				"PRODUCT.yaml": &PullRequestFile{
					Name:    "v1.23/cool/PRODUCT.yaml",
					BlobURL: "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.23/cke/PRODUCT.yaml",
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
				"junit_01.xml": &PullRequestFile{
					Name:    "v1.23/cool/junit_01.xml",
					BlobURL: "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.23/cke/junit_01.xml",
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
				"e2e.log": &PullRequestFile{
					BlobURL: "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.23/cke/e2e.log",
					Name:    "v1.23/cool/e2e.log",
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
				Title:  "Conformance results for v1.23 Something (Passed)",
				Number: 2,
			},
			Labels: []string{"no-failed-tests-v1.23", "release-documents-checked", "release-v1.23", "tests-verified-v1.23"},
			SupportingFiles: map[string]*PullRequestFile{
				"README.md": &PullRequestFile{
					Name:     "v1.23/cool/README.md",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.23/cke/README.md",
					Contents: `# Conformance test for Something`,
				},
				"PRODUCT.yaml": &PullRequestFile{
					Name:    "v1.23/cool/PRODUCT.yaml",
					BlobURL: "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.23/cke/PRODUCT.yaml",
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
				"junit_01.xml": &PullRequestFile{
					Name:    "v1.23/cool/junit_01.xml",
					BlobURL: "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.23/cke/junit_01.xml",
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
				"e2e.log": &PullRequestFile{
					Name:    "v1.23/cool/e2e.log",
					BlobURL: "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.23/cke/e2e.log",
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
				Title:  "Conformance results for Something (Failing)",
				Number: 3,
			},
			Labels: []string{"release-documents-checked", "release-v1.23", "required-tests-missing"},
			SupportingFiles: map[string]*PullRequestFile{
				"PRODUCT.yaml": &PullRequestFile{
					Name:    "v1.23/cool/PRODUCT.yaml",
					BlobURL: "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.23/cke/PRODUCT.yaml",
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
				"junit_01.xml": &PullRequestFile{
					Name:     "v1.23/cool/junit_01.xml",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.23/cke/junit_01.xml",
					Contents: ``,
				},
				"e2e.log": &PullRequestFile{
					Name:    "v1.23/cool/e2e.log",
					BlobURL: "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.23/cke/e2e.log",
					Contents: `
May 27 04:41:36.616: INFO: 3 / 3 pods ready in namespace 'kube-system' in daemonset 'node-dns' (0 seconds elapsed)
May 27 04:41:36.616: INFO: e2e test version: v2
May 27 04:41:36.617: INFO: kube-apiserver version: v2
May 27 04:41:36.617: INFO: >>> kubeConfig: /tmp/kubeconfig-441052555
May 27 04:41:36.620: INFO: Cluster IP family: ipv4
SSSSS
`,
				},
				"recipe.org": &PullRequestFile{
					Name:    "recipe.org",
					BlobURL: "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/recipe.org",
					Contents: `
* How to cook pasta

1. bring water to a boil
2. add some oil and salt to the water
3. add pasta to the pot
4. once pasta is aldente, remove from heat
5. drain water from the pot
`,
				},
			},
		},
	}
}

type PRSuite struct {
	PR *PullRequest

	Suite godog.TestSuite
}

func (s *PRSuite) aConformanceProductSubmissionPR() error {
	if s.PR == nil {
		return fmt.Errorf("PR doesn't exist")
	}
	return nil
}

func (s *PRSuite) thePRTitleIsNotEmpty() error {
	if len(s.PR.Title) == 0 {
		return fmt.Errorf("Title (%v) length is too short!", s.PR.Title)
	}
	return nil
}

func (s *PRSuite) isIncludedInItsFileList(file string) error {
	for fileName := range s.PR.SupportingFiles {
		if fileName == file {
			return nil
		}
	}
	return fmt.Errorf("Could not find %v in file list", file)
}

func (s *PRSuite) fileFolderStructureMustMatchRegex(match string) error {
	pattern := regexp.MustCompile(match)

	for _, file := range s.PR.SupportingFiles {
		allIndexes := pattern.FindAllSubmatchIndex([]byte(file.Name), -1)
		for _, loc := range allIndexes {
			baseFolder := string(file.Name[loc[2]:loc[3]])
			distroName := string(file.Name[loc[4]:loc[5]])

			if baseFolder == "" || distroName == "" {
				return fmt.Errorf("The content structure of your product submission PR must match '%v' (KubernetesReleaseVersion/ProductName, e.g: v1.23/averycooldistro)", match)
			}
		}
	}
	return nil
}

func (s *PRSuite) theTitleOfThePR() error {
	if s.PR.Title == "" {
		return fmt.Errorf("unable to use product submission PR, as it appears to not have a title")
	}
	return nil
}

func (s *PRSuite) theTitleOfThePRMustMatch(match string) error {
	pattern := regexp.MustCompile(match)
	if pattern.MatchString(string(s.PR.Title)) != true {
		return fmt.Errorf("unable to use product submission PR, as the title doesn't appear to match what's required, please use something like 'Conformance results for $KubernetesReleaseVersion $ProductName' (e.g: Conformance results for v1.23 CoolKubernetes)")
	}
	return nil
}

func (s *PRSuite) theFilesInThePR() error {
	return nil
}

func (s *PRSuite) aFile(fileName string) error {
	if s.PR.SupportingFiles[fileName] == nil {
		return fmt.Errorf("unable to find file")
	}
	return nil
}

func (s *PRSuite) theYamlFileMustContainTheRequiredAndNonEmptyField(fileName, fieldName string) error {
	var parsedContent map[string]*interface{}
	err := yaml.Unmarshal([]byte(s.PR.SupportingFiles[fileName].Contents), &parsedContent)
	if err != nil {
		return fmt.Errorf("Unable to read '%v'", fileName)
	}
	if parsedContent[fieldName] == nil {
		return fmt.Errorf("missing or empty field '%v' in file '%v'", fieldName, fileName)
	}
	return nil
}

func (s *PRSuite) isNotEmpty(fileName string) error {
	if s.PR.SupportingFiles[fileName].Contents == "" {
		return fmt.Errorf("file '%v' is empty", fileName)
	}
	return nil
}

func (s *PRSuite) aLineOfTheFileMustMatch(fileName, match string) error {
	pattern := regexp.MustCompile(match)
	lines := strings.Split(s.PR.SupportingFiles[fileName].Contents, "\n")
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

func (s *PRSuite) InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^a conformance product submission PR$`, s.aConformanceProductSubmissionPR)
	ctx.Step(`^the PR title is not empty$`, s.thePRTitleIsNotEmpty)
	ctx.Step(`^"([^"]*)" is included in its file list$`, s.isIncludedInItsFileList)
	ctx.Step(`^the files in the PR`, s.theFilesInThePR)
	ctx.Step(`^file folder structure must match "([^"]*)"$`, s.fileFolderStructureMustMatchRegex)
	ctx.Step(`^the title of the PR$`, s.theTitleOfThePR)
	ctx.Step(`^the title of the PR must match "([^"]*)"$`, s.theTitleOfThePRMustMatch)
	ctx.Step(`^the yaml file "([^"]*)" must contain the required and non-empty "([^"]*)"$`, s.theYamlFileMustContainTheRequiredAndNonEmptyField)
	ctx.Step(`^a "([^"]*)" file$`, s.aFile)
	ctx.Step(`^"([^"]*)" is not empty$`, s.isNotEmpty)
	ctx.Step(`^a line of the file "([^"]*)" must match "([^"]*)"$`, s.aLineOfTheFileMustMatch)
}

func main() {
	prs := GetPRs()
	for _, pr := range prs {
		suite := PRSuite{
			PR: &pr,
		}
		status := godog.TestSuite{
			Name:                "how-are-the-prs",
			ScenarioInitializer: suite.InitializeScenario,
		}.Run()
		fmt.Println("status: ", status)
	}
}
