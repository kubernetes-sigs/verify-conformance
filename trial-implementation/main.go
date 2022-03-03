package main

import (
	"fmt"
	"strings"

	"cncf.io/infra/verify-conformance-release/pkg/suite"
)

var latest = "v1.23.4"

func GetPRs() []suite.PullRequest {
	return []suite.PullRequest{
		{
			PullRequestQuery: suite.PullRequestQuery{
				Title:  "Update docs",
				Number: 0,
			},
		},
		{
			PullRequestQuery: suite.PullRequestQuery{
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
			SupportingFiles: []*suite.PullRequestFile{
				&suite.PullRequestFile{
					Name:     "v1.23/cool/README.md",
					BaseName: "README.md",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.23/cke/README.md",
					Contents: `# Conformance test for Cool`,
				},
				&suite.PullRequestFile{
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
				&suite.PullRequestFile{
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
				&suite.PullRequestFile{
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
			PullRequestQuery: suite.PullRequestQuery{
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
			SupportingFiles: []*suite.PullRequestFile{
				&suite.PullRequestFile{
					Name:     "v1.23/cool/README.md",
					BaseName: "README.md",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.23/cke/README.md",
					Contents: `# Conformance test for Something`,
				},
				&suite.PullRequestFile{
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
				&suite.PullRequestFile{
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
				&suite.PullRequestFile{
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
			PullRequestQuery: suite.PullRequestQuery{
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
			SupportingFiles: []*suite.PullRequestFile{
				&suite.PullRequestFile{
					Name:     "v1.23/cool/README.MD",
					BaseName: "README.MD",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.23/cke/README.md",
					Contents: `# Conformance test for Something`,
				},
				&suite.PullRequestFile{
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
				&suite.PullRequestFile{
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
				&suite.PullRequestFile{
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
			PullRequestQuery: suite.PullRequestQuery{
				Title:  "Conformance results for Failurernetes (Failing at pretty much everything)",
				Number: 3,
			},
			Labels: []string{"release-documents-checked", "release-v1.18", "required-tests-missing"},
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
			SupportingFiles: []*suite.PullRequestFile{
				&suite.PullRequestFile{
					Name:     "v1.19/cool-metal/PRODUCT.yaml",
					BaseName: "PRODUCT.yaml",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.19/cke/PRODUCT.yaml",
					Contents: `
vendor: Something
name: something - A Cool Kubernetes Engine
version: v1.19.3
website_url: https://something.kubernetes/engine
repo_url: https://github.com/something/kubernetes-engine
product_logo_url: https://github.com/cybozu-go/cke/blob/main/logo/cybozu_logo.svg
type: Installer
description: Something Kubernetes Engine, a distributed service that automates Kubernetes cluster management.
`,
				},
				&suite.PullRequestFile{
					Name:     "v1.19/cool-metal/junit_01.xml",
					BaseName: "junit_01.xml",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.19/cke/junit_01.xml",
					Contents: ``,
				},
				&suite.PullRequestFile{
					Name:     "v1.19/cool-metal/e2e.log",
					BaseName: "e2e.log",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.19/cke/e2e.log",
					Contents: `
May 27 04:41:36.616: INFO: 3 / 3 pods ready in namespace 'kube-system' in daemonset 'node-dns' (0 seconds elapsed)
May 27 04:41:36.616: INFO: e2e test version: v2
May 27 04:41:36.617: INFO: kube-apiserver version: v2
May 27 04:41:36.617: INFO: >>> kubeConfig: /tmp/kubeconfig-441052555
May 27 04:41:36.620: INFO: Cluster IP family: ipv4
SSSSS
`,
				},
				&suite.PullRequestFile{
					Name:     "v1.19/cool/PRODUCT.yaml",
					BaseName: "PRODUCT.yaml",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.19/cke/PRODUCT.yaml",
					Contents: `
vendor: Something
name: something - A Cool Kubernetes Engine
version: v1.19.3
website_url: https://something.kubernetes/engine
repo_url: https://github.com/something/kubernetes-engine
product_logo_url: https://github.com/cybozu-go/cke/blob/main/logo/cybozu_logo.svg
type: Installer
description: Something Kubernetes Engine, a distributed service that automates Kubernetes cluster management.
`,
				},
				&suite.PullRequestFile{
					Name:     "v1.19/cool/junit_01.xml",
					BaseName: "junit_01.xml",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.19/cke/junit_01.xml",
					Contents: ``,
				},
				&suite.PullRequestFile{
					Name:     "v1.19/cool/e2e.log",
					BaseName: "e2e.log",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.19/cke/e2e.log",
					Contents: `
May 27 04:41:36.616: INFO: 3 / 3 pods ready in namespace 'kube-system' in daemonset 'node-dns' (0 seconds elapsed)
May 27 04:41:36.616: INFO: e2e test version: v2
May 27 04:41:36.617: INFO: kube-apiserver version: v2
May 27 04:41:36.617: INFO: >>> kubeConfig: /tmp/kubeconfig-441052555
May 27 04:41:36.620: INFO: Cluster IP family: ipv4
SSSSS
`,
				},
				&suite.PullRequestFile{
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
		{
			PullRequestQuery: suite.PullRequestQuery{
				Title:  "Conformance results for v1.18 Something (invalid version)",
				Number: 2,
			},
			Labels: []string{
				"release-v1.18",
				"not-verifiable",
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
			SupportingFiles: []*suite.PullRequestFile{
				&suite.PullRequestFile{
					Name:     "v1.18/cool/README.md",
					BaseName: "README.md",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.18/cke/README.md",
					Contents: `# Conformance test for Something`,
				},
				&suite.PullRequestFile{
					Name:     "v1.18/cool/PRODUCT.yaml",
					BaseName: "PRODUCT.yaml",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.18/cke/PRODUCT.yaml",
					Contents: `
vendor: Something
name: something - A Cool Kubernetes Engine
version: v1.18.3
website_url: https://something.kubernetes/engine
repo_url: https://github.com/something/kubernetes-engine
documentation_url: https://github.com/something/kubernetes-engine
product_logo_url: https://github.com/cybozu-go/cke/blob/main/logo/cybozu_logo.svg
type: Installer
description: Something Kubernetes Engine, a distributed service that automates Kubernetes cluster management.
`,
				},
				&suite.PullRequestFile{
					Name:     "v1.18/cool/junit_01.xml",
					BaseName: "junit_01.xml",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.18/cke/junit_01.xml",
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
				&suite.PullRequestFile{
					Name:     "v1.18/cool/e2e.log",
					BaseName: "e2e.log",
					BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.18/cke/e2e.log",
					Contents: `
May 27 04:41:36.616: INFO: 3 / 3 pods ready in namespace 'kube-system' in daemonset 'node-dns' (0 seconds elapsed)
May 27 04:41:36.616: INFO: e2e test version: v1.18.4
May 27 04:41:36.617: INFO: kube-apiserver version: v1.18.4
May 27 04:41:36.617: INFO: >>> kubeConfig: /tmp/kubeconfig-441052555
May 27 04:41:36.620: INFO: Cluster IP family: ipv4
SSSSS
`,
				},
			},
		},
		{
			PullRequestQuery: suite.PullRequestQuery{
				Title:  "Conformance results for v1.23 Something (no files found)",
				Number: 2,
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
		},
	}
}

func main() {
	prs := GetPRs()
	for _, pr := range prs {
		prSuite := suite.NewPRSuite(&pr).
			SetSubmissionMetadatafromFolderStructure()
		prSuite.KubernetesReleaseVersionLatest = latest
		prSuite.NewTestSuite(suite.PRSuiteOptions{Paths: []string{"../kodata/features"}}).Run()

		finalComment, labels, err := prSuite.GetLabelsAndCommentsFromSuiteResultsBuffer()
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println("PR title:", prSuite.PR.Title)
		fmt.Println("Release Version:", prSuite.KubernetesReleaseVersion)
		fmt.Println("Labels:", strings.Join(labels, ", "))
		fmt.Println(finalComment)
	}
}
