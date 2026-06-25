# Security Policy

## Reporting a vulnerability

Please report suspected vulnerabilities privately through GitHub Security Advisories at https://github.com/excelano/quickxorhash/security/advisories/new. If you would rather not use GitHub, email david.anderson@excelano.com instead. I aim to respond within seven days.

Please do not open public issues for security problems.

## Supported versions

The latest v1.x release receives any needed fixes. Older versions are not supported.

## Scope

This package is a pure-function hashing library with no dependencies beyond the Go standard library. It performs no I/O, opens no network connections, reads no environment, and stores nothing. It computes a 160-bit digest from bytes you hand it and returns it. The practical security surface is correctness of the digest, which is covered by known-answer tests against values captured from Microsoft Graph.
