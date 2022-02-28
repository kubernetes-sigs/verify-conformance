Feature: A cool

  Background:
    Given a conformance product submission PR

  # Scenario: PR title is not empty
  #   Given a PR title
  #   Then the PR title is not empty

  # Scenario: Is true
  #   Given a <value>
  #   Then the value is "true"

  #   Examples:
  #     | value |
  #     | true  |
  #     | true  |
  #     | false |

  Scenario: submission contains all required files
    Given a conformance product submission PR
    Then <file> is included in its file list
    And <file> is not empty

    Examples:
      | file           |
      | "README.md"    |
      | "PRODUCT.yaml" |
      | "e2e.log"      |
      | "junit_01.xml" |

  Scenario: submission must have files in structure of releaseversion/productname/
    Given the files in the PR
    Then file folder structure must match "(v1.[0-9]{2})/(.*)/.*"
    # $1 is the release version of Kubernetes
    # $2 is the product name
    # example: v1.23/coolthing/some.file

  Scenario: the PRODUCT.yaml metadata must contain required fields
    Given a "PRODUCT.yaml" file
    Then the yaml file "PRODUCT.yaml" must contain the required and non-empty <field>
    # And if <type> is "url", the content of the url in the <field>'s value must match it's <dataType>

    Examples:
      | field               | contentType | dataType                           |
      | "vendor"            | "info"      | "string"                           |
      | "name"              | "info"      | "string"                           |
      | "version"           | "info"      | "string"                           |
      | "type"              | "info"      | "string"                           |
      | "description"       | "info"      | "string"                           |
      | "website_url"       | "url"       | "text/html"                        |
      | "repo_url"          | "url"       | "text/html"                        |
      | "documentation_url" | "url"       | "text/html"                        |
      | "product_logo_url"  | "url"       | "image/svg application/postscript" |

  Scenario: title of product submission must contain Kubernetes release version and product name
    Given the title of the PR
    Then the title of the PR must match "(.*) (v1.[0-9]{2})[ /](.*)"
    # $1 is the string for conformance results for
    # $2 is the version of Kubernetes
    # $3 is the product name
    # example: Conformance test for v1.23 Cool Engine

  Scenario: the e2e.log output must contain the Kubernetes release version
    Given a "e2e.log" file
    Then a line of the file "e2e.log" must match "^.*e2e test version: (v1.[0-9]{2}(.[0-9]{1,2})?)$"
    # $1 is the release version of Kubernetes
    # $2 is the (optional) point release version of Kubernetes
    # example: Feb 25 10:20:32.383: INFO: e2e test version: v1.23.0

  Scenario: there are labels for tests succeeding
    Given a list of labels in the PR
    Then the label prefixed with <label> and ending with Kubernetes release version should be present

    Examples:
      | label              |
      | "no-failed-tests-" |
      | "tests-verified-"  |
    # example: no-failed-tests-v1.23