Feature: CCC.Core.C13.TR01 - Minimize Certificate Lifetime - Valid Unexpired Certificates
  As a security administrator
  I want all certificate-based encryption to use valid, unexpired certificates from trusted authorities
  So that encryption remains secure and up-to-date

  Background:
    Given a service with ports exposed using certificate-based encryption
    And certificate authorities are configured

  @positive
  Scenario: Service uses valid unexpired certificate
    Given a port is exposed that uses certificate-based encryption
    When the certificate is examined
    Then the certificate should be valid and unexpired
    And the certificate should be issued by a trusted certificate authority
    And the certificate should be properly configured

  @positive
  Scenario: Service uses certificate from trusted CA
    Given a port is exposed that uses certificate-based encryption
    When the certificate authority is verified
    Then the certificate should be issued by a trusted CA
    And the CA should be in the approved certificate authority list
    And the CA certificate should be valid

  @positive
  Scenario: Certificate management tools track expiration
    Given certificate management tools are in use
    When certificates are monitored
    Then expiration dates should be tracked automatically
    And alerts should be configured for upcoming expirations
    And renewal processes should be automated where possible

  @negative
  Scenario: Service rejects expired certificate
    Given a port is exposed that uses certificate-based encryption
    When an expired certificate is presented
    Then the connection should be rejected
    And an appropriate error message should be returned
    And no encrypted communication should occur

  @negative
  Scenario: Service rejects certificate from untrusted CA
    Given a port is exposed that uses certificate-based encryption
    When a certificate from an untrusted authority is presented
    Then the connection should be rejected
    And certificate authority validation should fail
    And an appropriate error message should be returned

  @negative
  Scenario: Service rejects self-signed certificate
    Given a port is exposed that uses certificate-based encryption
    When a self-signed certificate is presented
    Then the connection should be rejected
    And certificate authority validation should fail
    And an appropriate error message should be returned

  @negative
  Scenario: Service rejects revoked certificate
    Given a port is exposed that uses certificate-based encryption
    When a revoked certificate is presented
    Then the connection should be rejected
    And certificate revocation should be detected
    And an appropriate error message should be returned

  @monitoring
  Scenario: Automated certificate renewal is configured
    Given certificate management tools are deployed
    Then certificate renewal should be automated where possible
    And renewal should occur before expiration
    And the process should be monitored and logged

  @validation
  Scenario: Certificate validation includes all security checks
    Given a certificate is being validated
    Then expiration date should be checked
    And certificate authority should be verified
    And certificate revocation status should be checked
    And certificate chain should be validated

  @compliance
  Scenario: Only certificates from approved authorities are deployed
    Given certificate deployment processes
    Then only certificates from trusted authorities should be allowed
    And certificate authority approval lists should be maintained
    And unauthorized certificates should be rejected
