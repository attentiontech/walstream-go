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
result, created, err := c.Apply(ctx, walstream.PipelineSpec{
    Name: "my-postgres",
    Source: walstream.SourceConfig{
        Connection: "my-postgres",
        Tables: []walstream.Table{
            {Schema: "public", Name: "users"},
            {Schema: "public", Name: "orders"},
        },
    },
    Destination: walstream.DestinationConfig{
        Connection: "my-kafka",
        Kafka: walstream.KafkaDestinationConfig{
            TopicPrefix: "myapp_",
            Initial: walstream.KafkaTopicInitial{
                Partitions:    3,
                CleanupPolicy: walstream.CleanupPolicyDelete,
            },
        },
    },
    DesiredStatus: walstream.DesiredStatusRunning,
})
```

## Core Types

- **PipelineSpec**: The persistent definition of a streaming pipeline
- **PipelineState**: Runtime state including status, errors, and statistics
- **SourceConfig**: PostgreSQL connection and table specifications
- **DestinationConfig**: Kafka configuration and topic settings
- **KafkaTopicOverride**: Per-table Kafka topic customization

## API Endpoints

The library provides types for API responses:

- **ApplyResponse**: Result of applying/creating a pipeline configuration
- **DestroyResponse**: Result of destroying a pipeline
- **Response**: Base response envelope with optional messages

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

### Message Levels

- `MessageLevelInfo`: Informational message
- `MessageLevelWarning`: Warning message

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
