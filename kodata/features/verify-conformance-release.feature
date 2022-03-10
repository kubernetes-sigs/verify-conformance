# verify-conformance-release

# The behaviour of the bot is described here, in [[https://cucumber.io/docs/gherkin/][Gherkin]].  Each scenario is a requirement a PR must meet to qualify for conformance.

# Note: the line immediately beneath the scenario is the comment posted to the PR if the requirement is not met.

Feature: verify conformance product submission PR

  Scenario: PR title is not empty
    it seems that there is no title set

    Given a PR title
    Then the PR title is not empty

  Scenario: submission contains all required files
    there seems to be some required files missing (https://github.com/cncf/k8s-conformance/blob/master/instructions.md#contents-of-the-pr)

    Given <file> is included in its file list
    Then <file> is not empty

    Examples:
      | file           |
      | "README.md"    |
      | "PRODUCT.yaml" |
      | "e2e.log"      |
      | "junit_01.xml" |

  Scenario: submission has files in structure of releaseversion/productname/
    the submission file directory does not seem to match the Kubernetes release version in the files

    Given the files in the PR
    Then file folder structure matches "(v1.[0-9]{2})/(.*)"
    # $1 is the release version of Kubernetes
    # $2 is the product name
    # example: v1.23/coolthing

  Scenario: submission is only one product
    the submission seems to contain files of multiple Kubernetes release versions or products. Each Kubernetes release version and products should be submitted in a separate PRs

    Given the files in the PR
    Then there is only one path of folders

  Scenario: submission release version in title matches release version in folder structure
    the title of the submission does not seem to contain a Kubernetes release version that matches the release version in the submitted files

    Given the files in the PR
    And the title of the PR
    Then the release version matches the release version in the title

  Scenario: the PRODUCT.yaml metadata contains required fields
    it appears that the PRODUCT.yaml file does not contain all the required fields (https://github.com/cncf/k8s-conformance/blob/master/instructions.md#productyaml)

    Given a "PRODUCT.yaml" file
    Then the yaml file "PRODUCT.yaml" contains the required and non-empty <field> where not <optional>
    And if <contentType> is set to url, the content of the url in the value of <field> matches it's <dataType> where not <optional>

    Examples:
      | field               | contentType | dataType                           | optional |
      | "vendor"            | "info"      | "string"                           | "false"  |
      | "name"              | "info"      | "string"                           | "false"  |
      | "version"           | "info"      | "string"                           | "false"  |
      | "type"              | "info"      | "string"                           | "false"  |
      | "description"       | "info"      | "string"                           | "false"  |
      | "website_url"       | "url"       | "text/html"                        | "false"  |
      | "repo_url"          | "url"       | "text/html"                        | "true"   |
      | "documentation_url" | "url"       | "text/html"                        | "false"  |
      | "product_logo_url"  | "url"       | "image/svg application/postscript" | "true"   |

  Scenario: title of product submission contains Kubernetes release version and product name
    the submission title is missing either a Kubernetes release version (v1.xx) or product name

    Given the title of the PR
    Then the title of the PR matches "(.*) (v1.[0-9]{2})[ /](.*)"
    # $1 is the string for conformance results for
    # $2 is the version of Kubernetes
    # $3 is the product name
    # example: Conformance test for v1.23 Cool Engine

  Scenario: the e2e.log output contains the Kubernetes release version
    it seems the e2e.log does not contain the Kubernetes release version that match the submission title

    Given an "e2e.log" file
    Then a line of the file "e2e.log" matches "^.*e2e test version: (v1.[0-9]{2}(.[0-9]{1,2})?)$"
    And that version matches the same Kubernetes release version as in the folder structure
    # $1 is the release version of Kubernetes
    # $2 is the (optional) point release version of Kubernetes
    # example: Feb 25 10:20:32.383: INFO: e2e test version: v1.23.0

  Scenario: the submission release version is a supported version of Kubernetes
    the Kubenetes release version in this pull request does not qualify for conformance submission anymore (https://github.com/cncf/k8s-conformance/blob/master/terms-conditions/Certified_Kubernetes_Terms.md#qualifying-offerings-and-self-testing)

    Given the release version
    And the files in the PR
    Then it is a valid and supported release

  Scenario: all required conformance tests in the junit_01.xml are present
    it appears that some tests are missing from the product submission

    Given a "junit_01.xml" file
    Then all required tests in junit_01.xml are present

  Scenario: all tests pass in e2e.log
    it appears that some tests failed in the product submission

    Given an "e2e.log" file
    Then the tests pass and are successful
    And all required tests in e2e.log are present

  Scenario: the tests in junit_01.xml and e2e.log match
    it appears that there is a mismatch of tests in junit_01.xml and e2e.log

    Given an "e2e.log" file
    And a "junit_01.xml" file
    Then the tests match
