# Usage

## Version information

Print the version, commit and build date with either command:

```sh
frontmatter-validator version
frontmatter-validator --version
```

Both commands write the same output to stdout and exit successfully without
loading configuration or scanning files:

```text
frontmatter-validator version 1.2.3 (commit: abc123, built: 2026-11-20T12:00:00Z)
```

Release builds use the metadata stamped by the build pipeline. Other builds fall
back to Go module and VCS build information. When no release version is available,
the version is `dev`; unavailable commit and date values are `unknown`. The VCS
date is the commit time, rather than the compilation time.
