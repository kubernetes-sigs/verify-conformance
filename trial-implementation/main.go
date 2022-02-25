package main

import (
	"fmt"
	"path"

	"github.com/cucumber/godog"
	githubql "github.com/shurcooL/githubv4"
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
	Contents string
}

type PullRequest struct {
	PullRequestQuery

	Labels          []string
	SupportingFiles map[string]PullRequestFile
}

func GetPRs() []PullRequest {
	return []PullRequest{
		{
			PullRequestQuery: PullRequestQuery{
				Title: "Cool (Passing)",
			},
			Labels: []string{},
			SupportingFiles: map[string]PullRequestFile{
				"v1.13/cool/README.md": PullRequestFile{
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.20/cke/README.md",
					Contents: `# Conformance test for Cool`,
				},
				"v1.13/cool/PRODUCT.yaml": PullRequestFile{
					BlobURL: "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.20/cke/PRODUCT.yaml",
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
				"v1.13/cool/junit_01.xml": PullRequestFile{
					BlobURL: "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.20/cke/junit_01.xml",
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
				"v1.13/cool/e2e.log": PullRequestFile{
					BlobURL: "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.20/cke/e2e.log",
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
				Title: "Something (Passed)",
			},
			Labels: []string{"no-failed-tests-v1.23", "release-documents-checked", "release-v1.23", "tests-verified-v1.23"},
			SupportingFiles: map[string]PullRequestFile{
				"v1.13/cool/README.md": PullRequestFile{
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.20/cke/README.md",
					Contents: `# Conformance test for Something`,
				},
				"v1.13/cool/PRODUCT.yaml": PullRequestFile{
					BlobURL: "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.20/cke/PRODUCT.yaml",
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
				"v1.13/cool/junit_01.xml": PullRequestFile{
					BlobURL: "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.20/cke/junit_01.xml",
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
				"v1.13/cool/e2e.log": PullRequestFile{
					BlobURL: "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.20/cke/e2e.log",
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
				Title: "Something (Failing)",
			},
			Labels: []string{"release-documents-checked", "release-v1.23", "required-tests-missing"},
			SupportingFiles: map[string]PullRequestFile{
				"v1.13/cool/PRODUCT.yaml": PullRequestFile{
					BlobURL: "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.20/cke/PRODUCT.yaml",
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
				"v1.13/cool/junit_01.xml": PullRequestFile{
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.20/cke/junit_01.xml",
					Contents: ``,
				},
				"v1.13/cool/e2e.log": PullRequestFile{
					BlobURL: "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.20/cke/e2e.log",
					Contents: `
May 27 04:41:36.616: INFO: 3 / 3 pods ready in namespace 'kube-system' in daemonset 'node-dns' (0 seconds elapsed)
May 27 04:41:36.616: INFO: e2e test version: v2
May 27 04:41:36.617: INFO: kube-apiserver version: v2
May 27 04:41:36.617: INFO: >>> kubeConfig: /tmp/kubeconfig-441052555
May 27 04:41:36.620: INFO: Cluster IP family: ipv4
SSSSS
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
		if path.Base(fileName) == file {
			return nil
		}
	}
	return fmt.Errorf("Could not find %v in file list", file)
}

func (s *PRSuite) InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^a conformance product submission PR$`, s.aConformanceProductSubmissionPR)
	ctx.Step(`^the PR title is not empty$`, s.thePRTitleIsNotEmpty)
	ctx.Step(`^"([^"]*)" is included in its file list$`, s.isIncludedInItsFileList)
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
