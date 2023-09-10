Feature: verify conformance product submission PR

  Scenario: PR title is not empty
    it seems that there is no title set

    Given a PR title
    Then the PR title is not empty

  Scenario: submission contains all required files
    there seems to be some required files missing (https://github.com/cncf/k8s-conformance/blob/master/instructions.md#contents-of-the-pr)

    Given <file> is included in its file list
    Then <file> is not empty
    And <file> is valid <type>

    Examples:
      | file           | type       |
      | "README.md"    | "markdown" |
      | "PRODUCT.yaml" | "yaml"     |
      | "e2e.log"      | "text"     |
      | "junit_01.xml" | "xml"      |

  Scenario: submission only contains required files
    Given the files in the PR
    Then the files included in the PR are only: README.md, PRODUCT.yaml, e2e.log, junit_01.xml

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

  Scenario: the PRODUCT.yaml metadata contains all required fields
    it appears that the PRODUCT.yaml file does not contain all the required fields (https://github.com/cncf/k8s-conformance/blob/master/instructions.md#productyaml)

    Given a "PRODUCT.yaml" file
    Then the yaml file "PRODUCT.yaml" contains the required and non-empty <field>

    Examples:
      | field                   |
      | "vendor"                |
      | "name"                  |
      | "version"               |
      | "type"                  |
      | "description"           |
      | "website_url"           |
      | "documentation_url"     |
      | "contact_email_address" |

  Scenario: the URL and email fields in the PRODUCT.yaml are valid
    it appears that field(s) in the PRODUCT.yaml aren't correctly formatted

    Given a "PRODUCT.yaml" file
    Then the content of the <type> in the value of <field> is a valid <type>

    Examples:
      | field                   | type        |
      | "website_url"           | "URL"       |
      | "repo_url"              | "URL"       |
      | "documentation_url"     | "URL"       |
      | "product_logo_url"      | "URL"       |
      | "contact_email_address" | "email"     |

  Scenario: the URL fields in the PRODUCT.yaml resolve to their specified data types
    it appears that URL(s) in the PRODUCT.yaml don't resolve to the correct data type

    Given a "PRODUCT.yaml" file
    Then the content of the url in the value of <field> matches it's <dataType>

    Examples:
      | field               | dataType                           |
      | "website_url"       | "text/html"                        |
      | "repo_url"          | "text/html"                        |
      | "documentation_url" | "text/html"                        |
      | "product_logo_url"  | "image/svg application/postscript" |

  Scenario: title of product submission contains Kubernetes release version and product name
    the submission title is missing either a Kubernetes release version (v1.xx) or product name

    Given the title of the PR
    Then the title of the PR matches "(.*) (v1.[0-9]{2})[ /](.*)"
    # $1 is the string for conformance results for
    # $2 is the version of Kubernetes
    # $3 is the product name
    # example: Conformance test for v1.23 Cool Engine

  Scenario: the submission release version is a supported version of Kubernetes
    the Kubernetes release version in this pull request does not qualify for conformance submission anymore (https://github.com/cncf/k8s-conformance/blob/master/terms-conditions/Certified_Kubernetes_Terms.md#qualifying-offerings-and-self-testing)

    Given the release version
    And the files in the PR
    Then it is a valid and supported release

  Scenario: all required conformance tests in the junit_01.xml are present
    it appears that some tests are missing from the product submission

    Given a "junit_01.xml" file
    Then all required tests in junit_01.xml are present

  Scenario: all tests pass
    it appears that some tests failed in the product submission

    Given an "junit_01.xml" file
    Then the tests pass and are successful
    And all required tests are present

  Scenario: there is only one commit
    it appears that there is not exactly one commit. Please rebase and squash with `git rebase -i HEAD` (https://git-scm.com/docs/git-rebase)

    Given a list of commits
    Then there is only one commit
