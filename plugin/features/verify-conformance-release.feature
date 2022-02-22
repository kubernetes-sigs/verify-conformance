# Behaviour

Feature: Verify Conformance Product Submission

  Scenario: Files must exist in correct folders
    Given a conformance product submission PR
    When a PR is submitted
    Then file folder structure must match regex "(v1.[0-9]{2})/(.*)"
    # $1 is the release version of Kubernetes
    # $2 is the product name

  Scenario: Check for required files
    Given a conformance product submission PR
    When the PR is submitted
    Then it contain the required files
    And each <file> must not be empty when required via it's BlobURL

    Examples:
      | file         |
      | README.md    |
      | PRODUCT.yaml |
      | e2e.log      |
      | junit_01.xml |

  Scenario: PRODUCT.yaml must contain required fields
    Given a conformance product submission PR
    When the PRODUCT.yaml is found
    Then the PRODUCT.yaml must contain the following required <field>
    But the required <field> must not be empty
    And if <type> is "url", the content of the url must exist and match it's <dataType>

    Examples:
      | field             | type | dataType                         |
      | vendor            | info | string                           |
      | name              | info | string                           |
      | version           | info | string                           |
      | type              | info | string                           |
      | description       | info | string                           |
      | website_url       | url  | text/html                        |
      | repo_url          | url  | text/html                        |
      | documentation_url | url  | text/html                        |
      | product_logo_url  | url  | image/svg|application/postscript |

  Scenario: Check product name is in PR title
    Given a conformance product submission PR
    When the PR is submitted
    Then the title of the PR must match "(.*) (v1.[0-9]{2})[ /](.*)"
    # $1 is the string for conformance results for
    # $2 is the version of Kubernetes
    # $3 is the product name

  Scenario: Check e2e.log for Kubernetes release version
    Given a conformance product submission PR
    When the e2e.log is found
    Then search for line that matches "e2e test version: (v1.[0-9]{2}(.[0-9]{2})?)"
    # $1 is the release version of Kubernetes
    # $2 is the (optional) point release version of Kubernetes
