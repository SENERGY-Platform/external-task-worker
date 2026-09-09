# Aspects arrive as a list

A command names a set of aspects, not one aspect. The single-valued fields that
predate the list still exist, are marked deprecated, and behave as an alias for
a list with one element.

## Scope

Applies from `models/go v0.0.0-20260902082034-9c8c8bd56d88` and
`marshaller v0.2.0` onwards, to everything this worker reads from camunda and
everything it writes to a protocol topic.

Not about the aspects of a **content variable**. Those are `aspect_ids` on the
device type and say what a variable carries; the fields here say what a command
demands. The two are named similarly and point in opposite directions.

## The fold happens once, at the camunda boundary

`GetCommandRequest` in `lib/command.go` calls `command.SetAspects()` directly
after `json.Unmarshal`. From there inward nothing reads the deprecated field:

```go
err = json.Unmarshal([]byte(payload), &command)
if err != nil {
	return command, err
}
command.SetAspects()
```

A deployment written before the lists sends `aspect`, a current one sends
`aspects`, and both reach the rest of the worker as `Aspects`. That is the only
place the fold is needed for a task, so a new consumer of `Command` does not
have to know about the deprecated field — it reads `Aspects` and is correct for
both.

## The fields and their accessors

| Type | Deprecated | List | Accessors |
|---|---|---|---|
| `messages.Command` | `aspect` | `aspects` | `GetAspects()`, `SetAspects()` |
| `messages.Metadata` | `output_aspect_node` | `output_aspect_nodes` | `GetOutputAspectNodes()`, `SetOutputAspectNodes()` |
| `marshaller.ConfigurableV2` | `aspect_node` | `aspect_nodes` | `GetAspectNodes()` |
| `messages.EventRequest` | — | `AspectNodes` | — |

`EventRequest` is in-process only and never serialized, so it carries the list
alone.

The accessors are in `lib/messages/aspects.go` and are built on the marshaller's
own helpers rather than on local copies:
`marshallermodel.AspectNodesAlias` for the fold,
`AspectMatchLevel` and `ContentVariableAspectIds` for the matching. Reusing them
is deliberate — the AND-over-queried-aspects rule with subtree coverage is the
device-repository's rule, and a second implementation of it here would drift.

## Writing keeps the deprecated field filled

`SetOutputAspectNodes` writes the list **and** the deprecated single node, so a
protocol handler that only knows the old field still gets an answer. The node it
picks is the one with the alphabetically first id, which is what the
device-repository does when it fills the deprecated field of a path option. That
is a silent narrowing for such a reader: it sees one node out of several, chosen
by how the urns sort rather than by which aspect matters.

Both output aspect fields are written only for `command.Version >= 3` and only
when an output characteristic is set (`lib/protocol.go`). A version-2 command
takes the v1 marshalling path and its protocol message carries neither field, so
a test or a handler on that path is unaffected by any of this.

## Device group filtering

`getFilteredServices` in `lib/devicegroups/devicegroups.go` takes the aspects
from `command.GetAspects()` and asks, per content variable, whether it satisfies
the whole set:

```go
if variable.FunctionId == functionId &&
	marshallermodel.AspectMatchLevel(marshallermodel.ContentVariableAspectIds(variable), aspectNodes) >= 0 {
	return true
}
```

`AspectMatchLevel` returns `-1` as soon as one queried aspect is unmatched, so
several aspects are an AND on **one** variable, and each covers its own subtree.
It reads the subtree from `ChildIds` and `DescendentIds` of the queried node, so
the nodes have to come from the device-repository — a node carrying only an id
matches nothing but itself.
