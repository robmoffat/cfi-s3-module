Feature: CCC.Core.C01.TR01 - Encrypt Data for Transmission - TLS 1.3 for Non-SSH Traffic
  As a security administrator
  I want to ensure all non-SSH network traffic uses TLS 1.3 or higher
  So that data integrity and confidentiality are protected during transmission

  Background:
    Given a service with exposed network ports
    And the service handles non-SSH network traffic

  @positive
  Scenario: Service accepts TLS 1.3 encrypted traffic
    Given a port is exposed for non-SSH network traffic
    When a client connects using TLS 1.3
    Then the connection should be established successfully
    And all traffic should be encrypted using TLS 1.3
    And a TLS handshake should be completed

  @positive
  Scenario: Service accepts higher TLS version traffic
    Given a port is exposed for non-SSH network traffic
    When a client connects using TLS 1.4 or higher
    Then the connection should be established successfully
    And all traffic should be encrypted using the higher TLS version
    And a TLS handshake should be completed

  @negative
  Scenario: Service rejects unencrypted traffic
    Given a port is exposed for non-SSH network traffic
    When a client attempts to connect without encryption
    Then the connection should be rejected
    And no data should be transmitted

  @negative
  Scenario: Service rejects TLS 1.2 traffic
    Given a port is exposed for non-SSH network traffic
    When a client connects using TLS 1.2
    Then the connection should be rejected
    And an appropriate error message should be returned

  @negative
  Scenario: Service rejects TLS 1.1 traffic
    Given a port is exposed for non-SSH network traffic
    When a client connects using TLS 1.1
    Then the connection should be rejected
    And an appropriate error message should be returned

  @negative
  Scenario: Service rejects TLS 1.0 traffic
    Given a port is exposed for non-SSH network traffic
    When a client connects using TLS 1.0
    Then the connection should be rejected
    And an appropriate error message should be returned

  @negative
  Scenario: Service rejects SSL traffic
    Given a port is exposed for non-SSH network traffic
    When a client connects using SSL
    Then the connection should be rejected
    And an appropriate error message should be returned

  @edge-case
  Scenario: Service handles TLS handshake failure gracefully
    Given a port is exposed for non-SSH network traffic
    When a client initiates a connection but the TLS handshake fails
    Then the connection should be terminated securely
    And no sensitive information should be leaked
