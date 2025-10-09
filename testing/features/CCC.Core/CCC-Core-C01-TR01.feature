@PerPort
Feature: CCC.Core.C01.TR01 - Encrypt Data for Transmission - TLS 1.3 for Non-SSH Traffic
  As a security administrator
  I want to ensure all non-SSH network traffic uses TLS 1.3 or higher
  So that data integrity and confidentiality are protected during transmission

  Scenario: Service accepts TLS 1.3 encrypted traffic
    Given a client connects using "tls1_3" for protocol "{protocol}" on port "{portNumber}"
    Then "{result}" is not nil
    And "{result}" is not an error

  Scenario: Service rejects TLS 1.2 traffic
    Given a client connects using "tls1_2" for protocol "{protocol}" on port "{portNumber}"
    Then "{result}" is an error

  Scenario: Service rejects TLS 1.1 traffic
    Given a client connects using "tls1_1" for protocol "{protocol}" on port "{portNumber}"
    Then "{result}" is an error

  Scenario: Service rejects TLS 1.0 traffic
    Given a client connects using "tls1" for protocol "{protocol}" on port "{portNumber}"
    Then "{result}" is an error

  Scenario: Verify SSL/TLS protocol support
    Given "report" contains details of SSL Support type "protocols" for "{hostName}" on port "{portNumber}"
    Then "{report}" is a slice of objects which doesn't contain any of
      | id     | finding |
      | SSLv2  | offered |
      | SSLv3  | offered |
      | TLS1   | offered |
      | TLS1_1 | offered |
      | TLS1_2 | offered |
    And "{report}" is a slice of objects with at least the following contents
      | id     | finding |
      | TLS1_3 | offered |

  Scenario: Verify no known SSL/TLS vulnerabilities
    Given "report" contains details of SSL Support type "vulnerable" for "{hostName}" on port "{portNumber}"
    Then "{report}" is a slice of objects with at least the following contents
      | id            | finding                        |
      | heartbleed    | not vulnerable                 |
      | CCS           | not vulnerable                 |
      | ticketbleed   | not vulnerable                 |
      | ROBOT         | not vulnerable                 |
      | secure_renego | secure renegotiation supported |

  Scenario: Verify strong server defaults
    Given "report" contains details of SSL Support type "server-defaults" for "{hostName}" on port "{portNumber}"
    Then "{report}" is a slice of objects with at least the following contents
      | id                    | finding      |
      | TLS_server_preference | cipher order |
