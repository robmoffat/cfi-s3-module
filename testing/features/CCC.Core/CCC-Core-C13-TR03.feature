Feature: CCC.Core.C13.TR03 - Minimize Certificate Lifetime - 90 Day Rotation for TLP Red
  As a security administrator
  I want certificates to be rotated within 90 days for TLP Red environments
  So that the highest security standards are maintained for the most sensitive data

  Background:
    Given a service with ports exposed using certificate-based encryption
    And the service handles TLP Red classified data
    And certificate management tools are configured for high-security environments

  @positive
  Scenario: Certificate is rotated within 90 days
    Given a port is exposed that uses certificate-based encryption
    And the certificate was issued 80 days ago
    When the certificate rotation policy is checked
    Then the certificate should be scheduled for rotation
    And rotation should occur within the next 10 days
    And the new certificate should be from a trusted authority

  @positive
  Scenario: Automated certificate rotation occurs at 75 days
    Given a port is exposed that uses certificate-based encryption
    And automated rotation is configured for 75 days
    When 75 days have passed since issuance
    Then the certificate should be automatically rotated
    And the new certificate should be deployed successfully
    And the old certificate should be immediately revoked

  @positive
  Scenario: High-frequency certificate rotation is supported
    Given the high-security requirements for TLP Red
    When certificates are rotated every 60 days
    Then the system should handle frequent rotations smoothly
    And no service disruption should occur
    And all rotations should be logged and audited

  @negative
  Scenario: Certificate exceeds 90 day lifetime
    Given a port is exposed that uses certificate-based encryption
    When a certificate has been active for more than 90 days
    Then this should be detected as a critical security violation
    And immediate alerts should be generated
    And the certificate should be forcibly rotated

  @negative
  Scenario: Certificate rotation is delayed beyond 90 days
    Given a certificate rotation was scheduled
    When the rotation is delayed and exceeds 90 days
    Then this should trigger a security incident
    And emergency rotation procedures should be initiated
    And the delay should be investigated and documented

  @negative
  Scenario: Automated rotation fails in TLP Red environment
    Given a certificate rotation fails
    When the failure occurs in a TLP Red environment
    Then immediate manual intervention should be triggered
    And security teams should be notified immediately
    And backup rotation procedures should be initiated

  @security
  Scenario: Certificate revocation is immediate upon rotation
    Given a certificate is rotated in TLP Red environment
    When the new certificate is deployed
    Then the old certificate should be immediately revoked
    And revocation should be published to all relevant CRL/OCSP services
    And revocation status should be verified

  @monitoring
  Scenario: Enhanced monitoring for TLP Red certificates
    Given certificates in TLP Red environment
    Then certificate age should be monitored in real-time
    And alerts should be generated at 60 days
    And warnings should be generated at 75 days
    And critical alerts should be generated at 85 days

  @automation
  Scenario: Fully automated rotation with minimal human intervention
    Given the high-security requirements
    Then certificate rotation should be fully automated
    And human intervention should be minimized
    And all processes should be logged and auditable
    And rotation should occur during approved maintenance windows

  @compliance
  Scenario: Strict compliance with 90 day rotation policy
    Given all certificates in the TLP Red environment
    When certificate ages are audited
    Then no certificate should be older than 90 days
    And any violations should be treated as security incidents
    And compliance should be verified daily

  @incident-response
  Scenario: Security incident triggered by rotation failure
    Given a certificate rotation failure in TLP Red
    When the failure is detected
    Then a security incident should be automatically created
    And incident response procedures should be initiated
    And all relevant stakeholders should be notified immediately
