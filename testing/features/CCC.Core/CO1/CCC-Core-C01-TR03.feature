@PerPort @tlp-green @tlp-amber @tlp-red
Feature: CCC.Core.C01.TR03 - Encrypt Data for Transmission - Block or Redirect Unencrypted Traffic
  As a security administrator
  I want unencrypted traffic to be blocked or redirected to secure equivalents
  So that no data is transmitted in plaintext

  @http
  Scenario: HTTP redirects to HTTPS
    If HTTP is accessible, it should immediately redirect to HTTPS (301/302 status codes).
    This ensures that all web traffic is encrypted.

    Given an openssl s_client request to "{portNumber}" on "{hostName}" protocol "http" as "connection"
    Then "{result}" is not an error
    And "{result}" contains "301"

  @ftp
  Scenario: FTP traffic is blocked or not exposed
    Unencrypted FTP should not be accessible. The service should either refuse connections
    or not expose FTP on standard ports (21).

    Given an openssl s_client request to "21" on "{hostName}" protocol "ftp" as "connection"
    Then "{result}" is an error

  @telnet
  Scenario: Telnet traffic is blocked or not exposed
    Telnet transmits credentials in plaintext and should be completely disabled.
    SSH should be used instead for remote shell access.

    Given an openssl s_client request to "{portNumber}" on "{hostName}" protocol "telnet" as "connection"
    Then "{result}" is an error

  Scenario: Only secure protocols are exposed
    Verify that the service only exposes encrypted protocols by checking that
    all exposed ports use TLS/SSL or other encryption.

    Given "report" contains details of SSL Support type "protocols" for "{hostName}" on port "{portNumber}"
    Then "{report}" is a slice of objects with at least the following contents
      | id     | finding            |
      | TLS1_2 | offered            |
      | id     | finding            |
      | TLS1_3 | offered with final |
