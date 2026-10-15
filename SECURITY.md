# Security Policy

## Supported versions

treegen is a single small binary with no server component. Only the latest
release gets fixes.

| Version | Supported |
| --- | --- |
| latest release | yes |
| anything older | no |

## Reporting a vulnerability

Please do not open a public issue.

Use [GitHub's private vulnerability reporting](https://github.com/callmegautam/treegen/security/advisories/new),
or email iamgautamsuthar@gmail.com. Include what you found, how to reproduce it,
and what an attacker gets out of it. I will acknowledge within a few days and
credit you in the release notes unless you would rather stay anonymous.

## What is worth reporting

treegen reads directories and writes empty files and folders, so the interesting
surface is small but real:

- **`treegen new` writing outside its target directory.** Entry names are reduced
  to their base name and `Apply` refuses any path that resolves outside the root.
  A tree that escapes anyway is a bug worth reporting.
- **`treegen new` clobbering existing data.** Without `--force`, an existing file
  must never be truncated.
- **The install scripts.** They download over HTTPS and verify against
  `checksums.txt`. A way to get an unverified or substituted binary installed is
  in scope.
- **Symlink handling.** The walk never follows symlinks. A case where it does,
  and escapes the target directory, is in scope.

Crashes on deliberately malformed input are usually just bugs, not
vulnerabilities. Open a normal issue for those.
