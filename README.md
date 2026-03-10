# walstream-go

Go client library for the walstream server, which streams PostgreSQL database changes to Apache Kafka topics.

## Overview

## Installation

```bash
go get github.com/attentiontech/walstream-go
```

## Quick Start

### Initialize the Client

```go
import (
    "github.com/attentiontech/walstream-go/client"
)

// Uses WALSTREAM_URL and WALSTREAM_TOKEN env vars by default
c := client.New()

// Or configure explicitly
c := client.New(
    client.WithServerURL("http://localhost:9795"),
    client.WithBearerToken("your-token"),
)
```

### Apply a Pipeline

Connection names reference server-side connection credentials managed separately from pipeline specs. The source connection refers to a named PostgreSQL connection, and the destination connection refers to a named Kafka connection.

```go
result, created, err := c.Pipelines.Apply(ctx, types.PipelineSpec{
    Name: "user_changes",
    Source: types.SourceConfig{
        Connection: "my-postgres",
        Tables: []types.Table{
            {Schema: "public", Name: "users"},
            {Schema: "public", Name: "orders"},
        },
    },
    Destination: types.DestinationConfig{
        Connection: "my-kafka",
        Kafka: types.KafkaDestinationConfig{
            TopicPrefix: "myapp_",
            Initial: types.KafkaTopicInitial{
                Partitions:    3,
                CleanupPolicy: types.CleanupPolicyDelete,
            },
        },
    },
    DesiredStatus: types.DesiredStatusRunning,
})
```

### List Pipelines

```go
pipelines, err := c.Pipelines.List(ctx)
for _, p := range pipelines {
    fmt.Printf("%s: %s\n", p.Name, p.Status)
}
```

### Get a Pipeline

```go
state, err := c.Pipelines.Get(ctx, "user_changes")
fmt.Printf("status: %s\n", state.Status)
if state.LastError != nil {
    fmt.Printf("last error: %s\n", *state.LastError)
}
```

### Check Pipeline Health

```go
status, err := c.Pipelines.Healthz(ctx, "user_changes")
fmt.Printf("health: %s\n", status)
```

### Stop a Pipeline

```go
state, err := c.Pipelines.Get(ctx, "user_changes")
state.DesiredStatus = types.DesiredStatusStopped
_, _, err = c.Pipelines.Apply(ctx, state.PipelineSpec)
```

### Destroy a Pipeline

```go
result, err := c.Pipelines.Destroy(ctx, "user_changes")
fmt.Printf("status: %s\n", result.Status)
```

## Core Types

- **PipelineSpec**: The persistent definition of a streaming pipeline
- **PipelineState**: Runtime state including status, errors, and statistics
- **SourceConfig**: PostgreSQL connection and table specifications
- **DestinationConfig**: Kafka configuration and topic settings
- **KafkaTopicOverride**: Per-table Kafka topic customization

## Status Constants

### Desired Status

- `DesiredStatusRunning`: Pipeline should be active
- `DesiredStatusStopped`: Pipeline should be inactive

### Effective Status

- `EffectiveStatusRunning`: Pipeline is actively streaming
- `EffectiveStatusFailing`: Pipeline encountered an error
- `EffectiveStatusRestarting`: Pipeline is restarting after failure
- `EffectiveStatusStopped`: Pipeline is stopped

## Configuration

### Cleanup Policies

- `CleanupPolicyDelete`: Delete old log segments after retention period
- `CleanupPolicyCompact`: Keep the latest value for each key

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
