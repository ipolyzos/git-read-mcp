# Security Policy

## Supported versions

Security fixes are applied on active release lines, with priority for the latest release and `main`.

## Reporting a vulnerability

Please report vulnerabilities privately by emailing:

- `ipolyzos@users.noreply.github.com`

Include:

- affected version/commit
- reproduction steps
- impact assessment
- any suggested remediation

Please do not disclose publicly until a fix is available.

## Response expectations

- Initial triage target: 3 business days
- Status update target: 7 business days
- Fix timeline: based on severity and complexity

## Scope

In-scope areas include:

- MCP input validation and error handling
- subprocess execution boundaries and timeouts
- path traversal prevention and repository containment
- Git command allowlisting and argument safety

See `docs/security.md` for technical model details.
