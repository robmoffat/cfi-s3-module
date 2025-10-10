# Go Cucumber Cloud Language Patterns

## Annotations

```gherkin
@PerPort           # Indicates that the test is written for a single port
@http, @ssh        # Indicate that the test only applies to a specific port protocol (see below)
@plaintext, @tls   # Applies to only plaintext/tls ports. e.g http is plaintext whereas https is tls.
@PerService        # This test applies across the whole service
@tlp-green @tlp-amber @tlp-red  # Traffic-light protocol level of the control.
```

## Pre-Configured Variables

Where a test is `@PerPort`:

- `portNumber` e.g. 22
- `hostName` e.g. example.com
- `protocol` e.g. imap, pop3, ldap, postgres
- `serviceType`

Where a test is `@PerService`:

- `hostName`
- `serviceType`

## Example OpenSSL Protocol Connection

### HTTPS

```gherkin
Given an openssl s_client request to "{portNumber}" on "{hostName}" protocol "smtp" as "connection"
Given an openssl s_client request using "tls1_1" to "{portNumber}" on "{hostName}" protocol "smtp" as "connection"
Then I transmit "{httpRequest}" over "{connection}"
# Where httpRequest could be:
GET / HTTP/1.1
Host: example.com
Connection: close

#
```

Will return the HTTP response in `result` and connection is closed.

### SMTP

```gherkin
Given an openssl s_client request to "{portNumber}" on "{hostName}" protocol "smtp" as "connection"
```

`result` might contain:

```
220 mail.example.com ESMTP Postfix
EHLO client.example.com
250-mail.example.com Hello client.example.com
250-STARTTLS
250 AUTH LOGIN PLAIN
STARTTLS
220 Ready to start TLS
```

```gherkin
Then I transmit "{smtpRequest}" over "{connection}"

# Where smtpRequest might be:
EHLO client.example.com
250-mail.example.com Hello client.example.com
250 AUTH LOGIN PLAIN
MAIL FROM:<you@example.com>
RCPT TO:<someone@example.com>
DATA
Subject: Test from OpenSSL

Hello world
.
QUIT
```

Will return the response in `result`.

### Arguments

- TLS argument: `tls1_1`, `tls1_2`, `tls1_3` etc.

| Protocol | Port | Start-TLS flag to use |
| -------- | ---- | --------------------- |
| SMTP     | 587  | `-starttls smtp`      |
| IMAP     | 143  | `-starttls imap`      |
| POP3     | 110  | `-starttls pop3`      |
| LDAP     | 389  | `-starttls ldap`      |
| Postgres | 5432 | `-starttls postgres`  |
| XMPP     | 5222 | `-starttls xmpp`      |

### Connection State

```gherkin
Then close connection "{connection}"
And "{connection}" is closed
```

Closes the opened connection.

## SSL Support

```gherkin
Given "report" contains details of SSL Support type "X" for "{hostName}" on port "{portNumber}"
Given "report" contains details of SSL Support type "X" for "{hostName}" on port "{portNumber}" with STARTTLS
```

This uses the `testssl.sh` project to return a JSON report about the SSL details on a specific port.  
Add STARTTLS if you wish to connect to a plaintext port and use TLS over it.

### Types

| Test Type           | Flag                  | Description                                                               |
| ------------------- | --------------------- | ------------------------------------------------------------------------- |
| `each-cipher`       | `--each-cipher`       | Checks each local cipher remotely                                         |
| `cipher-per-proto`  | `--cipher-per-proto`  | Checks ciphers per protocol                                               |
| `std`               | `--std`               | Tests standard cipher categories by strength                              |
| `forward-secrecy`   | `-f`                  | Checks forward secrecy settings                                           |
| `protocols`         | `-p`                  | Checks TLS/SSL protocols, for HTTP: including QUIC/HTTP/3 and ALPN/HTTP2  |
| `grease`            | `--grease`            | Tests several server implementation bugs like GREASE and size limitations |
| `server-defaults`   | `-S`                  | Displays the server's default picks and certificate info                  |
| `server-preference` | `--server-preference` | Displays the server's picks: protocol+cipher                              |
| `vulnerable`        | `-U`                  | Tests for various vulnerabilities (e.g., heartbleed)                      |

### JSON Structure

See the `examples_of_testssl` directory for examples of how this command produces JSON.

### Output Examples

**Protocol findings**:

```gherkin
Then "{report}" is a slice of objects which doesn't contain any of
  | id     | finding |
  | SSLv2  | offered |
  | SSLv3  | offered |
  | TLS1   | offered |
  | TLS1_1 | offered |
  | TLS1_2 | offered |
And "{report}" is a slice of objects with at least the following contents
  | id     | finding            |
  | TLS1_3 | offered with final |
```

**Vulnerability findings**:

```gherkin
Then "{report}" is a slice of objects with at least the following contents
  | id            | finding                                |
  | heartbleed    | not vulnerable, no heartbeat extension |
  | CCS           | not vulnerable                         |
  | ticketbleed   | not vulnerable                         |
  | ROBOT         | not vulnerable                         |
  | secure_renego | supported                              |
```

**Server defaults findings**:

```gherkin
Then "{report}" is a slice of objects with at least the following contents
  | id                    | finding |
  | cert_expirationStatus | ok      |
  | cert_chain_of_trust   | passed. |
```

## Example Plaintext Protocol Connection

### HTTP

```gherkin
Given a client connects to "{hostName}" with protocol "http" on port "{portNumber}" as "connection"
```

This establishes a plaintext HTTP connection to verify the server is listening and responding. The `result` will contain the HTTP server response:

```
HTTP/1.1 200 OK
Server: nginx/1.18.0
Content-Type: text/html
```

The test validates that the server responds successfully:

```gherkin
Then "{result}" is not nil
And "{result}" is not an error
And "{result}" contains "HTTP/1.1"
```

Note: HTTP should generally be redirected to HTTPS in production environments to ensure encrypted communications.

### Shell

```gherkin
Given a client connects to "{hostName}" with protocol "telnet" on port "{portNumber}" as "connection"
```

This establishes a plaintext telnet connection. The `result` will contain the server response, e.g.:

```
Ubuntu 22.04.1 LTS
login:
```

Note: Telnet should generally NOT be used in production as it transmits credentials in plaintext.

### FTP

```gherkin
Given a client connects to "{hostName}" with protocol "ftp" on port "{portNumber}" as "connection"
```

This establishes a plaintext FTP connection. The `result` will contain the FTP server banner:

```
220 (vsFTPd 3.0.3)
```

Note: FTP should be replaced with SFTP or FTPS for secure file transfers.

### Generating Examples

Use the `examples_of_testssl/generate-examples.sh` script:

```bash
cd examples_of_testssl
./generate-examples.sh <hostname>:<port>
# e.g., ./generate-examples.sh robmoff.at:443
```

This will generate JSON files for all test types: `<hostname>_<port>_<test-type>.json`
