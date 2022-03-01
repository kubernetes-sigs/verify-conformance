package main

import (
	"fmt"
	"strings"

	"cncf.io/infra/verify-conformance-release/pkg/suite"
)

func main() {
	prs := suite.GetPRs()
	for _, pr := range prs {
		prSuite := suite.NewPRSuite(&pr).
			SetSubmissionMetadatafromFolderStructure()
		prSuite.NewTestSuite().Run()

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
