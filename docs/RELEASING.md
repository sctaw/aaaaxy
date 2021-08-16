# Releasing

AAAAXY releases are published in the following places:

  - GitHub Releases

TODO(divVerent): add more; provide links.

# Versioning

We follow semantic versioning, but in an adapted form to suppport
speedrunning leaderboards.

Thus, the version components are incremented as follows:

  - Major version must be incremented on changes that likely break
    existing speedruns.
      - In particular, slowing down a section required for any speedrun
        categories, including 100%, requires a major version bump.
      - Exception: when some major cheese (grossly unintended skip) is
        fixed, a minor version bump is sufficient.
  - Minor version must be incremented on changes that likely add faster
    speedruns.
      - In particular, simplifying a section or making it optional
        requires a minor version bump.
  - Patch level must be incremented in any other case.
  - We use `*-alpha.*` pre-release sub-versions whenever major or minor
    have been bumped; this only gets turned into a real new
    `<major>.<minor>.0` version when actually releasing.
      - When patch level is bumped, switching to a pre-release version
        is not necessary.

Note that I say &ldquo;likely&rdquo;; if a section is not required for
any known/published speedrun, slowing it down may not require a major
version bump.

## Automation

Version is partially automated using git.

The version is built based on the closest git `v<major>.<minor>-alpha`
or `v<major>.<minor>` tag. Binary releases will be tagged
`r<major>.<minor>.<patch>` or `r<major>.<minor>.0-alpha<patch>`; any
commit on git counts as a source-only release with a patchlevel and
needs not be released as a binary.

## Conversion to Windows Version Numbers

The Windows scheme differs a little from semantic versioning by using a
`major.minor.revision.buildnumber` scheme. We convert as follows:

  - `major` maps to `major`.
  - `minor` maps to `minor`.
  - `patch` maps to `revision+N` where N is `0` for alpha, `10000` for
    beta, `20000` for rc and `30000` for finished versions.
  - `buildnumber` is always the total number of commits in the
    repository.