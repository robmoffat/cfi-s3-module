Feature: CCC.Core.C01.TR08 - Encrypt Data for Transmission - Mutual TLS (mTLS) Implementation
  As a security administrator
  I want all TLS connections to implement mutual TLS with client and server certificate authentication
  So that both parties are authenticated and connections are fully secure

  Background:
    Given a service that transmits data using TLS
    And the service handles sensitive data requiring mTLS

  @positive
  Scenario: Service successfully establishes mTLS connection
    Given a service configured for mTLS
    When a client connects with a valid client certificate
    Then the mTLS connection should be established successfully
    And both client and server certificates should be validated
    And the connection should be encrypted

  @positive
  Scenario: Service validates client certificate authority
    Given a service configured for mTLS
    When a client connects with a certificate from a trusted CA
    Then the client certificate should be validated successfully
    And the connection should be established
    And the certificate authority should be verified

  @positive
  Scenario: Service validates server certificate authority
    Given a service configured for mTLS
    When a client validates the server certificate
    Then the server certificate should be from a trusted CA
    And the certificate should be valid and unexpired
    And the connection should be established

  @positive
  Scenario: Automated certificate rotation works correctly
    Given a service with mTLS enabled
    When certificates are rotated automatically
    Then new certificates should be deployed successfully
    And connections should continue to work with new certificates
    And old certificates should be revoked properly

  @negative
  Scenario: Service rejects connection without client certificate
    Given a service configured for mTLS
    When a client attempts to connect without a client certificate
    Then the connection should be rejected
    And an appropriate error should be returned
    And no data should be transmitted

  @negative
  Scenario: Service rejects connection with invalid client certificate
    Given a service configured for mTLS
    When a client connects with an invalid or expired client certificate
    Then the connection should be rejected
    And certificate validation should fail
    And an appropriate error should be returned

  @negative
  Scenario: Service rejects connection with untrusted client certificate
    Given a service configured for mTLS
    When a client connects with a certificate from an untrusted CA
    Then the connection should be rejected
    And certificate authority validation should fail
    And an appropriate error should be returned

  @negative
  Scenario: Client rejects connection with invalid server certificate
    Given a service configured for mTLS
    When the server presents an invalid or expired certificate
    Then the client should reject the connection
    And certificate validation should fail
    And no data should be transmitted

  @negative
  Scenario: Service allows TLS without client authentication
    Given a service that should require mTLS
    When the service is configured to allow TLS without client certificates
    Then this configuration should be detected as non-compliant
    And the test should fail

  @configuration
  Scenario: mTLS is configured for all sensitive endpoints
    Given endpoints that process sensitive data
    Then all such endpoints should be configured for mTLS
    And client certificate authentication should be required
    And server certificate authentication should be enforced

  @certificate-management
  Scenario: Certificate authorities are properly managed
    Given mTLS configuration
    Then certificate authorities should be securely managed
    And only trusted CAs should be configured
    And CA certificates should be regularly reviewed and updated

  @monitoring
  Scenario: Certificate expiration is monitored
    Given mTLS enabled services
    Then certificate expiration dates should be tracked
    And alerts should be generated before certificates expire
    And automated renewal should be configured where possible
