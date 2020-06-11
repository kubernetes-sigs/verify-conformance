module cncf.io/infra/verify-conformance-tests

go 1.14

replace (
	cloud.google.com/go => cloud.google.com/go v0.44.3
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v12.2.0+incompatible
	github.com/golang/lint => golang.org/x/lint v0.0.0-20190301231843-5614ed5bae6f
	golang.org/x/lint => golang.org/x/lint v0.0.0-20190409202823-959b441ac422
	k8s.io/api => k8s.io/api v0.17.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.3
	k8s.io/client-go => k8s.io/client-go v0.17.3
	k8s.io/code-generator => k8s.io/code-generator v0.17.3
)

require (
	github.com/shurcooL/githubv4 v0.0.0-20200414012201-bbc966b061dd
	github.com/sirupsen/logrus v1.6.0
	k8s.io/test-infra v0.0.0-20200611174856-d80abdbad3dc
)
