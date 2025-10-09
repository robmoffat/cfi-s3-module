## Annotations

```gherkin
@PerPort           # Indicates that the test is written for a single port
@http              # Indicate that the test only applies to a specific port protocol (see below)
@PerService        # This test applies across the whole service
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

## Example Protocol Connection

```gherkin
Given a client connects using "tls1_3" for protocol "{protocol}" on port "{portNumber}"
```

Will create a connection and store it in `result`.

- Protocol argument: `tls1_1`, `tls1_2`, `tls1_3` etc.

| Protocol | Port | Start-TLS flag to use |
| -------- | ---- | --------------------- |
| SMTP     | 587  | `-starttls smtp`      |
| IMAP     | 143  | `-starttls imap`      |
| POP3     | 110  | `-starttls pop3`      |
| LDAP     | 389  | `-starttls ldap`      |
| Postgres | 5432 | `-starttls postgres`  |
| XMPP     | 5222 | `-starttls xmpp`      |

```gherkin
Close connection "{result}"
```

Closes the opened connection.

## SSL Support

```gherkin
Given "report" contains details of SSL Support type "X" for "{hostName}" on port "{portNumber}"
```

This uses the `testssl.sh` project to return a JSON report about the SSL details on a specific port. See `examples_of_testssl` directory for examples of the output.

### Types

- `each-cipher` checks each local cipher remotely
- `cipher-per-proto` checks those per protocol
- `std` tests standard cipher categories by strength
- `forward-secrecy` checks forward secrecy settings
- `protocols` checks TLS/SSL protocols, for HTTP: including QUIC/HTTP/3 and ALPN/HTTP2 (and SPDY)
- `grease` tests several server implementation bugs like GREASE and size limitations
- `server-defaults` displays the server's default picks and certificate info
- `server-preference` displays the server's picks: protocol+cipher
- `vulnerable` test for various vulnerabilities (e.g. heartbleed)

### Output

Depends on the rype, but follow a pattern like this:

```
Then "{report}" is a slice of objects with at least the following contents
| id           | finding    |
| heartbleed   | not vulnerable , timed out |
```

Or:

```gherkin
Then "{report}" is a slice of objects with at least the following contents
| id            | finding  |
| early_data    | No TLS 1.3 offered |
```
