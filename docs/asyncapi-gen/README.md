## Generate File in Version 2.4.0
```
cd asyncapi-gen
go generate ./...
```

## Convert to Version 3.0.0
copy/paste to https://studio.asyncapi.com
a dialog to convert should pop up. if not, you must clear your cookies/local-storage.

alternatively, you could use the `asyncapi convert` cli described in https://www.asyncapi.com/docs/migration/migrating-to-v3

## Additional Annotations
the resulting asyncapi.json should be modified
- channels.Protocol-Topic.address = "<Protocol-Topic>"
- channels.Protocol-Topic.description = "topic is Protocol.Handler related to the device that should receive a command"
- channels.camunda_incident.description = "deprecated"