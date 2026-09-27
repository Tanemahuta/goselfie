# goselfie

[![build](https://github.com/Tanemahuta/goselfie/actions/workflows/verify.yml/badge.svg)](https://github.com/Tanemahuta/goselfie/actions/workflows/verify.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/tanemahuta/goselfie.svg)](https://pkg.go.dev/github.com/tanemahuta/goselfie)

`goselfie` is a Ginkgo and Gomega disk-snapshot matcher inspired by [Selfie](https://selfie.dev/). It uses Selfie's update workflow: compare snapshots by default, mark one expectation with `_TODO` to replace it, use `selfieonce` for a temporary wider update, or use `SELFIEWRITE` to keep updating. See [Selfie's quickstart](https://selfie.dev/jvm/get-started) for the upstream workflow.

## Usage

Pass a typed lens that projects the value to bytes. `lens.ToYAML()`, for example, serializes values as YAML:

```go
import (
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    "github.com/tanemahuta/goselfie"
    "github.com/tanemahuta/goselfie/lens"
)

var _ = Describe("resource builder", func() {
    It("builds a resource", func() {
        resource := buildResource()
        Expect(resource).To(goselfie.MatchSnapshot(lens.ToYAML()))
    })
})
```

`MatchSnapshot` returns a Gomega matcher builder. `WithName` replaces the generated counter for that expectation:

```go
matcher := goselfie.MatchSnapshot(lens.ToYAML()).WithName("created resource")
Expect(resource).To(matcher)
```

Lenses run before comparison and storage. For example, YAML transformers can remove or mask unstable fields, and `lens/yaml/k8s` provides a default Kubernetes resource transformer.

## Matching and updates

`MatchSnapshot` compares the projected bytes with the stored snapshot. A missing snapshot is written automatically. A mismatch fails the expectation and prints the expected and actual values.

Use an update control when a snapshot should change:

| Control | Effect | After a passing `It` |
| --- | --- | --- |
| `MatchSnapshot_TODO(...)` | Creates or replaces this expectation only. | Rewrites the call to `MatchSnapshot(...)`. |
| `// selfieonce` | Allows missing and mismatched snapshots in its scope to update once. | Removes `selfieonce` comments from the test source. |
| `// SELFIEWRITE` | Allows snapshots in its scope to update on every run. | Leaves the comment in place. |
| `UPDATE_SNAPSHOTS=true` | Allows all encountered snapshots to update. | No source marker is added. |

In goselfie, a file-level directive must appear before the `package` declaration. A directive inside a Ginkgo container applies to that container's specs; one inside a matcher assertion applies to that expectation. For example:

```go
// selfieonce
package example_test
```

```go
var _ = Describe("resource builder", func() {
    // SELFIEWRITE
    It("builds a resource", func() {
        Expect(buildResource()).To(goselfie.MatchSnapshot(lens.ToYAML()))
    })
})
```

To update just one assertion temporarily, put the directive in its matcher expression:

```go
Expect(buildResource()).To(
    // selfieonce
    goselfie.MatchSnapshot(lens.ToYAML()),
)
```

`CI=true` takes precedence and disables snapshot updates. Missing or mismatched snapshots fail. Unlike Selfie's runner, goselfie does not separately fail just because an update marker is present; a passing spec still runs the usual source cleanup for `selfieonce` and successful `_TODO` matchers. Environment settings are read once per process.

## Names and files

An unnamed snapshot name contains the nested Ginkgo `Describe`/`Context`/`When` and leaf-node text, followed by an expectation number. For example, two expectations in `Describe("resources")` and `It("builds a resource")` are named:

```text
resources/builds a resource/#1
resources/builds a resource/#2
```

`WithName("created")` uses `resources/builds a resource/created` for that expectation. The counter still advances, so later unnamed expectations keep their number.

Snapshots for `resources_test.go` live in `__snapshots__/resources_test.ss` next to the test file. The album format stores text and YAML verbatim and binary data as wrapped hexadecimal. Each snapshot records its content type; YAML lenses write YAML snapshots.

Lenses receive the input content type in `Apply` and return the output content type with the projected value. Snapshot-taking resolves an unknown output type through the prioritized matchers in `snapshot/content` (YAML, text, and binary are registered by default, in that order). The same content registry resolves string encoders: YAML and text are written verbatim, while the default encoder represents bytes as hexadecimal.

`goselfie/gomega/failuremessage` registers text, YAML, and binary providers by default. Text and YAML use Gomega's `format.MessageWithDiff`; the registry's default provider renders bytes as spaced hexadecimal and uses the same Gomega string diff. Add a provider to `failuremessage.Providers` for a custom content type by implementing `FailureMessage(actual, expected []byte)` and `NegateFailureMessage(actual, expected []byte)`.

## Ginkgo lifecycle

The album belongs to the `_test.go` file: all expectations from that file use the same `.ss` archive. Ginkgo also has individual `It` nodes, and goselfie resolves the active file and container hierarchy when a matcher is constructed. At the end of each successful `It`, its cleanup hook writes the snapshots encountered by that `It` together with untouched snapshots from sibling `It` nodes. It removes stale entries only from the container path that ran. Matching uses the full nested container path and expectation name, so a sibling `It` with the same expectation name is not used as the expected value. The hook also cleans applicable source markers and removes temporary snapshot data. A failed `It` discards its temporary data without writing the album or editing the source file.

Writes use temporary files and atomic replacement, but concurrent Ginkgo processes writing snapshots for the same test file are not merged.

The matcher requires an active Ginkgo spec at construction time. For custom integrations and isolated tests, `goselfie/gomega` also exposes `NewMatcherBuilderWithContext`.

## Scope

Goselfie focuses on Selfie's disk-snapshot matching and update-marker workflow, using Ginkgo, Gomega, and typed lenses. Inline literal snapshots are outside this package's disk-snapshot API.

- `goselfie`: public matcher constructors.
- `goselfie/album`: snapshot file storage and streaming serialization.
- `goselfie/ginkgo`: test context, update modes, and source cleanup.
- `goselfie/gomega`: typed matcher builder.
- `goselfie/gomega/failuremessage`: content-type-specific Gomega failure providers.
- `goselfie/lens`: typed projections and YAML serialization.
- `goselfie/lens/yaml`: YAML value transformers.
- `goselfie/lens/yaml/k8s`: default normalization for Kubernetes resources.
- `goselfie/snapshot`: snapshot data and update-mode types.
- `goselfie/snapshot/content`: content type enum and prioritized matchers.
- `goselfie/utils/path`: compiled path matchers used by YAML transforms.

## Automation

Every push to `main` runs the Go verification workflow. After it passes, semantic-release creates a GitHub release and version tag from Conventional Commit messages: `!` or `BREAKING CHANGE:` produces a major release, `feat:` produces a minor release, and `fix:` or `perf:` produces a patch release. The aliases `breaking:` and `feature:` are also supported. Other commit types do not trigger a release.

Pull requests opened or updated by the repository owner are approved and set to merge automatically. Dependabot patch and minor updates, plus major development-dependency updates, follow the same path; other Dependabot updates receive a `dependabot:manual-review` label. Enable GitHub Actions approval permissions and auto-merge, and configure `go` as a required branch check so changes cannot merge before verification passes. The auto-merge workflow uses `pull_request_target` without checking out or running pull request code. For public repositories, allow this event in the repository or organization Actions policy; GitHub plans to enforce its default block on November 2, 2026.

Set the `AUTO_RELEASE_TOKEN` repository secret to a token with contents and pull-request write access. It lets auto-merged PRs trigger the verification and release workflows. Without it, the workflow falls back to `GITHUB_TOKEN`, whose generated merge does not start those follow-up workflows.
