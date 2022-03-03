# Behaviour

Feature: A cool

  Background:
    Given a conformance product submission PR

  Scenario: PR title is not empty
    it seems that there is no title set

    Given a PR title
    Then the PR title is not empty

  Scenario: submission contains all required files
    there seems to be some files missing

    Given a conformance product submission PR
    Then <file> is included in its file list
    And <file> is not empty

    Examples:
      | file           |
      | "README.md"    |
      | "PRODUCT.yaml" |
      | "e2e.log"      |
      | "junit_01.xml" |

  Scenario: submission has files in structure of releaseversion/productname/
    submission file structure is not like a conformance submission

    Given the files in the PR
    Then file folder structure matches "(v1.[0-9]{2})/(.*)"
    # $1 is the release version of Kubernetes
    # $2 is the product name
    # example: v1.23/coolthing/some.file

  Scenario: submission is only one product
    it appears that you are submitting more than one product

    Given the files in the PR
    Then there is only one path of folders

  Scenario: submission release version in title matches release version in folder structure
    it seems that the release version of Kubernetes that is found in your title doesn't match the version in the file structure

    Given the files in the PR
    And the title of the PR
    Then the release version matches the release version in the title

  Scenario: the PRODUCT.yaml metadata contains required fields
    there seems to be some missing fields in the PRODUCT.yaml

    Given a "PRODUCT.yaml" file
    Then the yaml file "PRODUCT.yaml" contains the required and non-empty <field>
    And if <contentType> is set to url, the content of the url in the value of <field> matches it's <dataType>

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

  Scenario: title of product submission contains Kubernetes release version and product name
    it appears that there isn't a product name or Kubernetes release version in the title of the submission

    Given the title of the PR
    Then the title of the PR matches "(.*) (v1.[0-9]{2})[ /](.*)"
    # $1 is the string for conformance results for
    # $2 is the version of Kubernetes
    # $3 is the product name
    # example: Conformance test for v1.23 Cool Engine

  Scenario: the e2e.log output contains the Kubernetes release version
    it appears that in the e2e.log the Kubernetes release version is not found

    Given a "e2e.log" file
    Then a line of the file "e2e.log" matches "^.*e2e test version: (v1.[0-9]{2}(.[0-9]{1,2})?)$"
    # $1 is the release version of Kubernetes
    # $2 is the (optional) point release version of Kubernetes
    # example: Feb 25 10:20:32.383: INFO: e2e test version: v1.23.0

  Scenario: the submission release version is a supported version of Kubernetes
    the release version of Kubernetes in this submission is not supported for conformance

    Given the release version
    And the files in the PR
    Then it is a valid and supported release

  # Scenario: there are labels for tests succeeding


  #   Given a list of labels in the PR
  #   Then the label prefixed with <label> and ending with Kubernetes release version should be present

  #   Examples:
  #     | label              |
  #     | "no-failed-tests-" |
  #     | "tests-verified-"  |
  #   # example: no-failed-tests-v1.23

  Scenario: all required conformance tests in the junit_01 and e2e.log pass and are successful
    it appears that some tests in the product submission appear to not pass

    Given a "e2e.log" file
    And a "junit_01.xml" file
    Then the tests must pass and be successful
