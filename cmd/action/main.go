package main

import (
	"encoding/json"
	"os"

	"github.com/sirupsen/logrus"
	prowflagutil "k8s.io/test-infra/prow/flagutil"
	"k8s.io/test-infra/prow/github"

	"cncf.io/infra/verify-conformance-release/pkg/plugin"
)

func main() {
	log := logrus.StandardLogger().WithField("plugin", "verify-conformance-release")
	githubToken, ok := os.LookupEnv("GITHUB_TOKEN")
	if !ok {
		log.Fatalf("error: unable to find environment variable: GITHUB_TOKEN")
	}
	eventFilePath, ok := os.LookupEnv("GITHUB_EVENT_PATH")
	if !ok {
		log.Fatalf("error: unable to find environment variable: GITHUB_EVENT_PATH")
	}
	tokenFile, err := os.CreateTemp("", "ghtoken-*")
	if err != nil {
		log.Fatalf("error: failed to create new temp file for token: %v", err)
	}
	defer func() {
		if _, err := os.ReadFile(tokenFile.Name()); err != nil {
			log.Fatalf("error: failed to remove temp github token file: %v", err)
		}
	}()
	if _, err := tokenFile.Write([]byte(githubToken)); err != nil {
		log.Fatalf("error: failed to write to temp token file")
	}

	githubClientOptions := prowflagutil.GitHubOptions{
		TokenPath: tokenFile.Name(),
	}
	ghc, err := githubClientOptions.GitHubClient(false)
	if err != nil {
		log.Fatalf("error: creating new GitHub client: %v", err)
	}
	var pullRequestEvent *github.PullRequestEvent
	eventFileBytes, err := os.ReadFile(eventFilePath)
	if err != nil {
		log.Fatalf("error: unable to read event file at path '%v': %v", eventFilePath, err)
	}
	if err := json.Unmarshal(eventFileBytes, &pullRequestEvent); err != nil {
		log.Fatalf("error: failed to parse event file into pull request: %v", err)
	}

	if err := plugin.HandlePullRequestEvent(log, ghc, pullRequestEvent); err != nil {
		log.Fatalf("error: failed to handle pull request event: %v", err)
	}
}
