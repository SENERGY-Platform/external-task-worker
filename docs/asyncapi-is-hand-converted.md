# The committed asyncapi.json is hand-converted

`docs/asyncapi.json` is not the file the generator writes. A new message field
has to be added to it by hand, and nothing in the build or in CI notices when
that is forgotten.

## Scope

Applies to this repository's AsyncAPI spec as of 2026-09-09. It is the same
split `device-repository` has, documented there in
`docs/asyncapi-is-half-generated.md`, with one difference: there the generated
file is the committed one, here it is not.

Not about the Swagger conventions for HTTP services — this worker has no HTTP
API to document, only kafka messages.

## Two files, no mechanical relation

`docs/asyncapi-gen/` is its own Go module, with
`replace github.com/SENERGY-Platform/external-task-worker => ../../`, so
`go generate ./...` from the repository root does not reach it. Running it means
changing into that directory, and a dependency bump in the root module needs a
`go mod tidy` there before the generator builds at all.

What it produces is `docs/asyncapi-gen/asyncapi.json`, in AsyncAPI **2.4.0** and
on a single line. That file is ignored, through
`docs/asyncapi-gen/.gitignore`.

What is committed is `docs/asyncapi.json`, an indented AsyncAPI **3.0.0**
document. It was converted by hand and there is no converter in the repository.

The consequence: the two files cannot be diffed against each other, so the
drift check the Go conventions ask for cannot be built from them as they stand.
Whoever wants that check has to make the generator write the file that is
committed.

## Adding a field

Generate first, read the field off the generator's output, then patch the
committed file by hand in 3.0.0 shape. The schemas live under
`components.schemas` in both, named `<Package><Type>` — a field on
`messages.Metadata` is a property of `MessagesMetadata`, one on
`model.ContentVariable` a property of `ModelsContentVariable`.

Do not regenerate the committed file from the generator's output and reformat
it: that would drop the 3.0.0 conversion and every field that was patched in by
hand since.
