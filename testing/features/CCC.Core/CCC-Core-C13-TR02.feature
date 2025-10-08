Feature: CCC.Core.C13.TR02 - Minimize Certificate Lifetime - 180 Day Rotation for TLP Amber
  As a security administrator
  I want certificates to be rotated within 180 days for TLP Amber environments
  So that certificate compromise risk is minimized through regular rotation

  Background:
    Given a service with ports exposed using certificate-based encryption
    And the service handles TLP Amber classified data
    And certificate management tools are configured

  @positive
  Scenario: Certificate is rotated within 180 days
    Given a port is exposed that uses certificate-based encryption
    And the certificate was issued 170 days ago
    When the certificate rotation policy is checked
    Then the certificate should be scheduled for rotation
    And rotation should occur within the next 10 days
    And the new certificate should be from a trusted authority

  @positive
  Scenario: Automated certificate rotation occurs at 150 days
    Given a port is exposed that uses certificate-based encryption
    And automated rotation is configured for 150 days
    When 150 days have passed since issuance
    Then the certificate should be automatically rotated
    And the new certificate should be deployed successfully
    And the old certificate should be properly revoked

  @positive
  Scenario: Certificate rotation is tracked and logged
    Given certificate management tools are in use
    When a certificate is rotated
    Then the rotation should be logged with timestamp
    And the old and new certificate details should be recorded
    And rotation compliance should be tracked

  @negative
  Scenario: Certificate exceeds 180 day lifetime
    Given a port is exposed that uses certificate-based encryption
    When a certificate has been active for more than 180 days
    Then this should be detected as non-compliant
    And an alert should be generated
    And immediate rotation should be required

  @negative
  Scenario: Certificate rotation fails
    Given a certificate is scheduled for rotation
    When the rotation process fails
    Then the failure should be detected and logged
    And alerts should be generated for manual intervention
    And the service should continue with the current valid certificate

  @negative
  Scenario: New certificate is invalid after rotation
    Given a certificate rotation is attempted
    When the new certificate is invalid or from untrusted CA
    Then the rotation should be rolled back
    And the previous valid certificate should remain active
    And an alert should be generated for investigation

  @automation
  Scenario: Certificate rotation is fully automated
    Given certificate management tools are configured
    Then rotation should be automated for certificates approaching 180 days
    And the process should not require manual intervention
    And rotation should be scheduled during maintenance windows

  @monitoring
  Scenario: Certificate age is continuously monitored
    Given certificates are deployed
    Then certificate age should be continuously tracked
    And alerts should be generated at 150 days
    And warnings should be generated at 160 days
    And critical alerts should be generated at 175 days

  @compliance
  Scenario: All certificates comply with 180 day rotation policy
    Given all certificates in the TLP Amber environment
    When certificate ages are audited
    Then no certificate should be older than 180 days
    And rotation schedules should be documented
    And compliance reports should be generated regularly
