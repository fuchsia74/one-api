# Alert Pusher Integration

This document describes the alert pusher integration in the One API logging system, which allows automatic notification of error-level log messages to external alert systems.

## Overview

The alert pusher integration uses the `github.com/Laisky/go-utils/v5/log` alert functionality to automatically send error-level log messages to external notification systems. This is particularly useful for monitoring production deployments and getting immediate notifications when errors occur.

## Configuration

The alert pusher is configured using three environment variables:

### Environment Variables

| Variable         | Description                             | Required | Example                                  |
| ---------------- | --------------------------------------- | -------- | ---------------------------------------- |
| `LOG_PUSH_API`   | The API endpoint URL for sending alerts | Yes      | `https://api.example.com/webhook/alerts` |
| `LOG_PUSH_TYPE`  | The type of alert system being used     | No       | `webhook`, `slack`, `discord`, etc.      |
| `LOG_PUSH_TOKEN` | Authentication token for the alert API  | No       | `your-secret-token`                      |

### Example Configuration

```bash
# Basic webhook configuration
export LOG_PUSH_API="https://your-webhook-url.com/api/alerts"
export LOG_PUSH_TYPE="webhook"
export LOG_PUSH_TOKEN="your-secret-token"

# Slack webhook configuration
export LOG_PUSH_API="https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"
export LOG_PUSH_TYPE="slack"

# Discord webhook configuration
export LOG_PUSH_API="https://discord.com/api/webhooks/YOUR/DISCORD/WEBHOOK"
export LOG_PUSH_TYPE="discord"
```

## Features

### Automatic Error Alerting

- Only error-level log messages trigger alerts (INFO, DEBUG, WARN levels are not sent)
- Rate limiting is built-in (1 alert per second maximum) to prevent spam
- Includes hostname context in all log messages for better traceability
- Non-blocking operation - alert failures don't affect application performance

### Enhanced Logging Context

All log messages now include:

- **Host**: The hostname of the server running the application
- **Structured fields**: All existing zap fields are preserved
- **Stack traces**: Error-level logs include stack trace information

### Graceful Degradation

- If alert pusher configuration is not provided, logging works normally without alerts
- If alert API is unreachable, warnings are logged but application continues normally
- Invalid configuration doesn't prevent application startup

## Usage Examples

### Basic Setup

1. Set the environment variables:

```bash
export LOG_PUSH_API="https://your-alert-system.com/api/webhook"
export LOG_PUSH_TYPE="webhook"
export LOG_PUSH_TOKEN="your-token"
```

2. Start the application:

```bash
./one-api
```

3. The application will automatically configure alert pusher and log:

```
INFO alert pusher configured {"alert_api": "https://your-alert-system.com/api/webhook", "alert_type": "webhook"}
```

### Testing Alert Functionality

You can test the alert functionality by triggering an error-level log:

```go
logger.Logger.Error("test alert message",
    zap.String("component", "test"),
    zap.String("error_type", "test_error"))
```

This will:

1. Log the error message locally
2. Send an alert to your configured alert system
3. Include hostname and structured fields in the alert

### Webhook Payload Format

When using webhook alerts, the payload sent to your endpoint will include:

```json
{
  "level": "error",
  "message": "your error message",
  "timestamp": "2025-07-31T20:14:52Z",
  "host": "your-hostname",
  "fields": {
    "component": "test",
    "error_type": "test_error"
  },
  "stack_trace": "..."
}
```

## Integration with Existing Systems

### Slack Integration

For Slack integration, create an incoming webhook in your Slack workspace and use the webhook URL as `LOG_PUSH_API`.

### Discord Integration

For Discord integration, create a webhook in your Discord server and use the webhook URL as `LOG_PUSH_API`.

### Custom Webhook Systems

The alert pusher can work with any HTTP webhook endpoint that accepts POST requests with JSON payloads.

## Monitoring and Troubleshooting

### Log Messages

- **Success**: `alert pusher configured` - Alert pusher is working
- **Warning**: `send alert mutation failed` - Alert delivery failed (check network/API)
- **Error**: `create AlertPusher` - Configuration error (check environment variables)

### Common Issues

1. **Network connectivity**: Ensure the alert API endpoint is reachable from your server
2. **Authentication**: Verify the `LOG_PUSH_TOKEN` is correct if required by your alert system
3. **Rate limiting**: The system limits to 1 alert per second to prevent spam

### Testing Configuration

Run the included tests to verify your setup:

```bash
go test ./common/logger -v
```

## Security Considerations

- Store `LOG_PUSH_TOKEN` securely and don't commit it to version control
- Use HTTPS endpoints for `LOG_PUSH_API` to ensure encrypted transmission
- Consider implementing IP whitelisting on your alert webhook endpoint
- Monitor alert volume to detect potential security issues or system problems

## Performance Impact

The alert pusher integration is designed to have minimal performance impact:

- Alerts are sent asynchronously (non-blocking)
- Rate limiting prevents excessive network requests
- Failed alert deliveries are logged but don't affect application performance
- Memory usage is minimal with built-in cleanup of old alert data
