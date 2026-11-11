# Security Policy

## Reporting a vulnerability

Please report suspected vulnerabilities privately, using the **Report a vulnerability** button on the Security tab of the repository concerned. If that button is not there, or you would rather not go through GitHub, email david.anderson@excelano.com instead. I aim to respond within seven days.

Please do not open public issues for security problems.

## Supported versions

The latest release receives security fixes. Older versions are not supported. A fix ships in a new release; update the dependency to take it.

## What quickxorhash can access

Nothing beyond the bytes you hand it. quickxorhash is a pure-function hashing library with no dependencies beyond the Go standard library. It performs no I/O, opens no network connections and reads no environment. It computes a 160-bit digest and returns it. The practical security surface is correctness of the digest, which is covered by known-answer tests against values captured from Microsoft Graph.

## What quickxorhash stores

Nothing.
