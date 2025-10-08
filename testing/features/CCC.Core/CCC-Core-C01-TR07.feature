Feature: CCC.Core.C01.TR07 - Encrypt Data for Transmission - IANA Port Assignment Compliance
  As a security administrator
  I want services to run only the officially assigned protocols on their designated ports
  So that port usage follows IANA standards and security best practices

  Background:
    Given a service with exposed network ports
    And IANA Service Name and Transport Protocol Port Number Registry standards

  @positive
  Scenario: HTTP service runs on port 80
    Given port 80 is exposed
    When the service is examined
    Then only HTTP protocol should be running on port 80
    And no other protocols should be present on this port

  @positive
  Scenario: HTTPS service runs on port 443
    Given port 443 is exposed
    When the service is examined
    Then only HTTPS protocol should be running on port 443
    And no other protocols should be present on this port

  @positive
  Scenario: SSH service runs on port 22
    Given port 22 is exposed
    When the service is examined
    Then only SSH protocol should be running on port 22
    And no other protocols should be present on this port

  @positive
  Scenario: FTP service runs on port 21
    Given port 21 is exposed
    When the service is examined
    Then only FTP protocol should be running on port 21
    And no other protocols should be present on this port

  @positive
  Scenario: SMTP service runs on port 25
    Given port 25 is exposed
    When the service is examined
    Then only SMTP protocol should be running on port 25
    And no other protocols should be present on this port

  @positive
  Scenario: DNS service runs on port 53
    Given port 53 is exposed
    When the service is examined
    Then only DNS protocol should be running on port 53
    And no other protocols should be present on this port

  @negative
  Scenario: Non-standard service running on well-known port 80
    Given port 80 is exposed
    When a non-HTTP service is running on port 80
    Then this should be detected as non-compliant
    And the test should fail

  @negative
  Scenario: HTTP service running on non-standard port 8080
    Given port 8080 is exposed
    When HTTP service is running on port 8080 instead of port 80
    Then this should be flagged for review
    And proper justification should be documented

  @negative
  Scenario: SSH service running on port 443
    Given port 443 is exposed
    When SSH service is running on port 443 instead of port 22
    Then this should be detected as non-compliant
    And the test should fail

  @negative
  Scenario: Multiple protocols running on same port
    Given a port is exposed
    When multiple different protocols are running on the same port
    Then this should be detected as non-compliant
    And the test should fail

  @negative
  Scenario: Custom service on reserved system port
    Given a system port (1-1023) is exposed
    When a custom application service is running on a reserved port
    Then this should be detected as non-compliant
    And the test should fail

  @validation
  Scenario: Validate port assignments against IANA registry
    Given all exposed ports on the service
    When each port is checked against IANA registry
    Then each port should run only its officially assigned protocol
    And any deviations should be documented and justified

  @monitoring
  Scenario: Continuous monitoring of port usage
    Given the service is operational
    Then port usage should be continuously monitored
    And any changes in protocol-to-port assignments should be detected
    And alerts should be generated for non-compliant configurations
