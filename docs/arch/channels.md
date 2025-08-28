# Channel Settings Guide

- [Channel Settings Guide](#channel-settings-guide)
  - [Overview](#overview)
  - [Purpose and Rationale](#purpose-and-rationale)
    - [Why Channels Matter](#why-channels-matter)
    - [Key Benefits](#key-benefits)
  - [Channel Configuration](#channel-configuration)
    - [Basic Settings](#basic-settings)
      - [**Channel Information**](#channel-information)
      - [**Authentication \& Connection**](#authentication--connection)
      - [**Access Control \& Models**](#access-control--models)
      - [**Advanced Configuration**](#advanced-configuration)
  - [Model Configs: Unified Pricing System](#model-configs-unified-pricing-system)
    - [What is Model Configs?](#what-is-model-configs)
    - [Key Features](#key-features)
      - [**🎯 Unified Format**](#-unified-format)
      - [**📊 Comprehensive Settings**](#-comprehensive-settings)
      - [**🔧 Advanced Configuration Options**](#-advanced-configuration-options)
      - [**💰 Pricing Calculation Examples**](#-pricing-calculation-examples)
    - [Migration and Compatibility](#migration-and-compatibility)
      - [**🔄 Automatic Migration**](#-automatic-migration)
      - [**✅ Migration Status**](#-migration-status)
    - [Using Model Configs](#using-model-configs)
      - [**1. 📝 JSON Editing**](#1--json-editing)
      - [**2. 🛠️ Helper Tools**](#2-️-helper-tools)
      - [**3. 🎨 Visual Feedback**](#3--visual-feedback)
  - [Channel Testing and Monitoring](#channel-testing-and-monitoring)
    - [Health Checks](#health-checks)
    - [Channel Status Indicators](#channel-status-indicators)
    - [Performance Metrics](#performance-metrics)
    - [Error Handling and Troubleshooting](#error-handling-and-troubleshooting)
    - [Channel Optimization](#channel-optimization)
  - [Debug Panel: Troubleshooting Made Easy](#debug-panel-troubleshooting-made-easy)
    - [What is the Debug Panel?](#what-is-the-debug-panel)
    - [Accessing the Debug Panel](#accessing-the-debug-panel)
    - [Debug Panel Features](#debug-panel-features)
      - [**📊 Migration Status Overview**](#-migration-status-overview)
      - [**🔍 Diagnostic Information**](#-diagnostic-information)
      - [**🛠️ One-Click Actions**](#️-one-click-actions)
    - [Common Issues and Solutions](#common-issues-and-solutions)
      - [**🔧 Mixed Model Data**](#-mixed-model-data)
      - [**⚠️ Migration Issues**](#️-migration-issues)
      - [**❌ Invalid JSON**](#-invalid-json)
  - [Best Practices](#best-practices)
    - [Channel Naming and Organization](#channel-naming-and-organization)
    - [Security Best Practices](#security-best-practices)
    - [Performance Optimization](#performance-optimization)
    - [Model Configs Best Practices](#model-configs-best-practices)
    - [Monitoring and Maintenance](#monitoring-and-maintenance)
    - [Disaster Recovery and Backup](#disaster-recovery-and-backup)
    - [Troubleshooting Workflow](#troubleshooting-workflow)
  - [Advanced Features](#advanced-features)
    - [Batch Operations](#batch-operations)
    - [API Integration](#api-integration)
    - [Security Features](#security-features)
  - [Support and Resources](#support-and-resources)
    - [Getting Help](#getting-help)
    - [Additional Resources](#additional-resources)
  - [Quick Reference](#quick-reference)
    - [Essential Configuration Checklist](#essential-configuration-checklist)
    - [Common Configuration Patterns](#common-configuration-patterns)
    - [Troubleshooting Quick Guide](#troubleshooting-quick-guide)
    - [Support Escalation Path](#support-escalation-path)

## Overview

The Channel Settings page is the central hub for configuring and managing API channels in One API. Channels represent connections to different AI service providers (OpenAI, Claude, etc.) and define how requests are routed, authenticated, and billed.

## Purpose and Rationale

### Why Channels Matter

- **Service Integration**: Connect to multiple AI providers through a unified interface
- **Load Balancing**: Distribute requests across multiple channels for better performance
- **Cost Management**: Configure pricing and billing for different models and providers
- **Access Control**: Manage which users and groups can access specific channels

### Key Benefits

- **Centralized Management**: Configure all AI providers from one interface
- **Flexible Pricing**: Set custom pricing for different models and usage patterns
- **Real-time Monitoring**: Track channel status, usage, and performance
- **Easy Troubleshooting**: Built-in debugging tools for quick issue resolution

## Channel Configuration

### Basic Settings

#### **Channel Information**

**Name**

- **Purpose**: Human-readable identifier for the channel
- **Format**: Any descriptive text (e.g., "OpenAI GPT-4 Production", "Claude Development")
- **Best Practice**: Use clear, descriptive names that indicate provider, model type, and environment

**Type**

- **Purpose**: Specifies the AI service provider and determines available configuration options
- **Available Types**:
  - `1` - OpenAI (GPT models, DALL-E, Whisper, TTS)
  - `8` - Claude (Anthropic models)
  - `3` - Azure OpenAI Service
  - `14` - Gemini (Google AI)
  - `16` - Cohere
  - `18` - Baidu (文心一言)
  - `19` - Zhipu (智谱 AI)
  - `23` - Tencent (腾讯混元)
  - And many more...
- **Impact**: Determines available models, authentication methods, and API endpoints

**Status**

- **Purpose**: Controls whether the channel is active and can receive requests
- **Options**:
  - `1` - Enabled (channel accepts requests)
  - `2` - Disabled (channel is bypassed)
- **Use Cases**: Disable for maintenance, testing, or when API keys are invalid

**Priority**

- **Purpose**: Determines channel selection order when multiple channels support the same model
- **Range**: Integer values (lower numbers = higher priority)
- **Behavior**: System tries channels in ascending priority order
- **Example**: Priority 1 channel is tried before Priority 5 channel

#### **Authentication & Connection**

**Base URL**

- **Purpose**: API endpoint for the service provider
- **Format**: Complete URL including protocol (https://)
- **Examples**:
  - OpenAI: `https://api.openai.com`
  - Azure: `https://your-resource.openai.azure.com`
  - Custom proxy: `https://your-proxy.example.com`
- **Notes**: Leave empty to use provider defaults

**API Key**

- **Purpose**: Authentication credentials for the service provider
- **Format**: Provider-specific key format
- **Security**: Stored encrypted in the database
- **Examples**:
  - OpenAI: `sk-...` (starts with sk-)
  - Claude: Anthropic API key
  - Azure: Resource-specific key

**Organization**

- **Purpose**: Organization identifier for providers that support it
- **Providers**: Primarily OpenAI and Azure
- **Format**: Organization ID string
- **Optional**: Leave empty if not using organization-scoped access

**Headers**

- **Purpose**: Custom HTTP headers sent with each API request
- **Format**: JSON object with header name-value pairs
- **Example**:
  ```json
  {
    "User-Agent": "MyApp/1.0",
    "X-Custom-Header": "custom-value"
  }
  ```
- **Use Cases**: Custom authentication, request tracking, proxy configuration

#### **Access Control & Models**

**Groups**

- **Purpose**: Restricts channel access to specific user groups
- **Format**: Comma-separated list of group names
- **Behavior**:
  - Empty = all users can access
  - Specified groups = only those group members can access
- **Example**: `admin,premium,developers`

**Models**

- **Purpose**: Defines which AI models are available through this channel
- **Format**: Comma-separated list of model names
- **Behavior**:
  - Empty = all provider models available
  - Specified models = only listed models available
- **Examples**:
  - OpenAI: `gpt-4,gpt-3.5-turbo,text-embedding-ada-002`
  - Claude: `claude-3-opus-20240229,claude-3-sonnet-20240229`

**Model Mapping**

- **Purpose**: Maps external model names to internal provider model names
- **Format**: JSON object with mapping pairs
- **Use Cases**:
  - Provide consistent model names across providers
  - Support legacy model names
  - Custom model aliases
- **Example**:
  ```json
  {
    "gpt-4": "gpt-4-0613",
    "gpt-3.5": "gpt-3.5-turbo-0613",
    "claude": "claude-3-sonnet-20240229"
  }
  ```

#### **Advanced Configuration**

**Config**

- **Purpose**: Provider-specific advanced configuration options
- **Format**: JSON object with provider-specific settings
- **Common Options**:
  - `region`: Geographic region for the service
  - `version`: API version to use
  - `timeout`: Request timeout in seconds
  - `max_retries`: Maximum retry attempts
- **Example**:
  ```json
  {
    "region": "us-east-1",
    "version": "2023-05-15",
    "timeout": 30,
    "max_retries": 3
  }
  ```

**Test Model**

- **Purpose**: Specific model used for channel health checks
- **Format**: Single model name
- **Behavior**: System uses this model to test channel connectivity
- **Default**: Uses the first available model if not specified

**Weight**

- **Purpose**: Load balancing weight for channels with same priority
- **Range**: Integer values (higher = more requests)
- **Behavior**: Distributes requests proportionally among same-priority channels
- **Example**: Weight 3 channel gets 3x more requests than weight 1 channel

**Rate Limit**

- **Purpose**: Controls the maximum number of requests per token per channel within a 3-minute window
- **Format**: Integer value (requests per 3 minutes)
- **Default**: 0 (unlimited)
- **Behavior**: When limit is reached, requests are rejected with HTTP 429 (Too Many Requests)
- **Use Cases**:
  - Prevent API abuse and excessive usage
  - Manage costs by limiting request frequency
  - Ensure fair usage across multiple users
  - Comply with upstream provider rate limits
- **Implementation**: Uses token-based rate limiting with SHA-256 hashed keys for privacy
- **Example**: Setting to 100 allows maximum 100 requests per token per channel every 3 minutes

##### System Prompt

- **Purpose**: Optional text prepended as a system message to every request routed through this channel
- **Behavior**: Injected at the start of the conversation; clients may still provide their own system messages which are merged according to adapter rules
- **Use Cases**: Enforce guardrails, tone/style guidance, safety or compliance instructions
- **Notes**: Keep concise; overly long prompts increase token usage and costs

##### AWS Bedrock Inference Profile ARN Map

- **Purpose**: Route specific model names to AWS Bedrock Inference Profiles
- **Format**: JSON object mapping model name → Inference Profile ARN
- **Example**:

  ```json
  {
    "claude-3-5-sonnet-20241022": "arn:aws:bedrock:us-east-1:123456789012:inference-profile/us.anthropic.claude-3-5-sonnet-20241022-v2:0",
    "claude-3-haiku-20240307": "arn:aws:bedrock:us-east-1:123456789012:inference-profile/us.anthropic.claude-3-haiku-20240307-v1:0"
  }
  ```

- **Notes**: Only applicable to AWS Bedrock channels; values must be valid ARNs in the configured region/account

## Model Configs: Unified Pricing System

### What is Model Configs?

Model Configs is a **unified configuration system** that consolidates all model-specific settings into a single, comprehensive format. It replaces the previous separate "Model Pricing" and "Completion Pricing" fields with a streamlined approach.

### Key Features

#### **🎯 Unified Format**

All model settings are now configured in one place using JSON format:

```json
{
  "gpt-3.5-turbo": {
    "ratio": 0.0015,
    "completion_ratio": 2.0,
    "max_tokens": 16385
  },
  "gpt-4": {
    "ratio": 0.03,
    "completion_ratio": 2.0,
    "max_tokens": 8192
  }
}
```

#### **📊 Comprehensive Settings**

**ratio**

- **Purpose**: Base pricing multiplier for input tokens (prompt tokens)
- **Format**: Decimal number (e.g., 0.0015, 0.03)
- **Calculation**: `final_cost = base_cost × ratio`
- **Examples**:
  - `0.001` = 1x the base cost
  - `0.5` = 0.5x the base cost (50% discount)
  - `2.0` = 2x the base cost (100% markup)
- **Use Cases**: Adjust pricing based on model cost, profit margins, or user tiers

**completion_ratio**

- **Purpose**: Pricing multiplier for output tokens (completion tokens)
- **Format**: Decimal number, typically higher than input ratio
- **Relationship**: Applied in addition to the base ratio
- **Calculation**: `completion_cost = base_cost × ratio × completion_ratio`
- **Rationale**: Output tokens are typically more expensive than input tokens
- **Examples**:
  - `2.0` = Output tokens cost 2x input tokens
  - `3.0` = Output tokens cost 3x input tokens
  - `1.0` = Same cost for input and output tokens

**max_tokens**

- **Purpose**: Maximum token limit for the model in a single request
- **Format**: Integer value
- **Enforcement**: System rejects requests exceeding this limit
- **Provider Limits**: Should not exceed actual model capabilities
- **Examples**:
  - GPT-3.5-turbo: `16385`
  - GPT-4: `8192`
  - Claude-3: `200000`
- **Use Cases**: Prevent excessive usage, control costs, enforce quotas

#### **🔧 Advanced Configuration Options**

You can add additional fields for specific use cases:

```json
{
  "gpt-4": {
    "ratio": 0.03,
    "completion_ratio": 2.0,
    "max_tokens": 8192,
    "temperature_limit": 1.0,
    "allowed_functions": true,
    "rate_limit_rpm": 3500,
    "rate_limit_tpm": 90000
  }
}
```

**temperature_limit**

- **Purpose**: Maximum allowed temperature parameter
- **Range**: 0.0 to 2.0
- **Default**: No limit if not specified

**allowed_functions**

- **Purpose**: Whether function calling is permitted
- **Values**: `true` or `false`
- **Impact**: Blocks function calls if set to `false`

**rate_limit_rpm**

- **Purpose**: Requests per minute limit for this model
- **Format**: Integer value
- **Enforcement**: System throttles requests exceeding this limit

**rate_limit_tpm**

- **Purpose**: Tokens per minute limit for this model
- **Format**: Integer value
- **Calculation**: Includes both input and output tokens

#### **💰 Pricing Calculation Examples**

**Example 1: Basic Calculation**

```
Model: gpt-3.5-turbo
Input tokens: 1000
Output tokens: 500
Base cost per 1M tokens: $1.0

Configuration:
{
  "ratio": 1.5,
  "completion_ratio": 2.0
}

Calculation:
Input cost = (1000/1000000) × $1.0 × 1.5 = $0.0015
Output cost = (500/1000000) × $1.0 × 1.5 × 2.0 = $0.0015
Total cost = $0.0015 + $0.0015 = $0.003
```

**Example 2: Different Model Pricing**

```json
{
  "gpt-3.5-turbo": {
    "ratio": 1.0,
    "completion_ratio": 2.0,
    "max_tokens": 16385
  },
  "gpt-4": {
    "ratio": 20.0,
    "completion_ratio": 2.0,
    "max_tokens": 8192
  },
  "text-embedding-ada-002": {
    "ratio": 0.1,
    "completion_ratio": 1.0,
    "max_tokens": 8191
  }
}
```

This configuration:

- Charges standard rates for GPT-3.5-turbo
- Charges 20x rates for GPT-4 (reflecting higher costs)
- Charges 0.1x rates for embeddings (cheaper model)

### Migration and Compatibility

#### **🔄 Automatic Migration**

- **Seamless Upgrade**: Existing channels are automatically migrated to the new format
- **Data Preservation**: All existing pricing data is preserved during migration
- **Backward Compatibility**: Legacy data remains accessible for reference

#### **✅ Migration Status**

The system provides clear indicators of migration status:

- **🟢 Migrated**: Fully converted to unified format
- **🟡 Partial**: Migrated with legacy data present
- **🔴 Needs Migration**: Requires manual intervention
- **⚪ Empty**: No pricing data configured

### Using Model Configs

#### **1. 📝 JSON Editing**

- **Auto-formatting**: JSON is automatically formatted for readability
- **Real-time validation**: Visual indicators show JSON validity
- **Syntax highlighting**: Monospace font for better code editing

#### **2. 🛠️ Helper Tools**

- **Load Default**: Automatically populate with provider-appropriate defaults
- **Format JSON**: Manually trigger pretty-printing
- **Validation**: Real-time feedback on JSON syntax
- **Interactive Tooltips**: Hover over question mark icons (❓) next to field labels for detailed explanations
  - **Rate Limit**: Explains the 3-minute window and usage scenarios
  - **Model Mapping**: Shows JSON format examples and use cases
  - **Model Configs**: Details pricing structure and field meanings
  - **System Prompt**: Describes forced prompt behavior and setup requirements

#### **3. 🎨 Visual Feedback**

- **✅ Valid JSON**: Green border and success indicator
- **❌ Invalid JSON**: Red border and error message
- **Auto-format on blur**: Formatting applied when you finish editing

## Channel Testing and Monitoring

### Health Checks

**Automatic Testing**

- **Purpose**: System automatically tests channel connectivity and functionality
- **Frequency**: Configurable interval (default: every 5 minutes)
- **Test Method**: Sends a simple request using the configured test model
- **Status Updates**: Channel status is updated based on test results

**Manual Testing**

- **Test Button**: Available in channel edit interface
- **Test Request**: Sends a predefined request to verify connectivity
- **Response Validation**: Checks for proper API response format
- **Error Reporting**: Displays detailed error messages for troubleshooting

**Test Model Configuration**

- **Purpose**: Specifies which model to use for health checks
- **Selection**: Should be a reliable, low-cost model
- **Fallback**: Uses first available model if test model is not specified
- **Examples**:
  - OpenAI: `gpt-3.5-turbo`
  - Claude: `claude-3-haiku-20240307`

### Channel Status Indicators

**Status Values**

- **1 (Enabled)**: Channel is active and healthy
- **2 (Disabled)**: Channel is manually disabled
- **3 (Auto-Disabled)**: Channel failed health checks and was automatically disabled
- **4 (Exhausted)**: Channel has exceeded quota or rate limits

**Status Monitoring**

- **Real-time Updates**: Status changes are reflected immediately in the interface
- **Alert Notifications**: System can send alerts when channels go offline
- **Historical Tracking**: Channel uptime and downtime are logged for analysis

### Performance Metrics

**Response Time Monitoring**

- **Average Response Time**: Tracked per channel and model
- **Percentile Metrics**: P50, P95, P99 response times
- **Trend Analysis**: Historical performance trends
- **Alerting**: Notifications when response times exceed thresholds

**Usage Statistics**

- **Request Count**: Total requests processed by the channel
- **Token Usage**: Input and output tokens consumed
- **Error Rate**: Percentage of failed requests
- **Cost Tracking**: Total costs incurred through the channel

**Load Balancing Metrics**

- **Request Distribution**: How requests are distributed among channels
- **Channel Utilization**: Percentage of capacity used per channel
- **Failover Events**: When and why requests failed over to backup channels

### Error Handling and Troubleshooting

**Common Error Types**

**Authentication Errors**

- **Symptoms**: 401 Unauthorized responses
- **Causes**: Invalid API key, expired credentials, wrong organization
- **Solutions**:
  - Verify API key format and validity
  - Check organization settings
  - Regenerate API key if necessary

**Rate Limiting Errors**

- **Symptoms**: 429 Too Many Requests responses
- **Causes**: Exceeding provider rate limits
- **Solutions**:
  - Reduce request frequency
  - Implement request queuing
  - Upgrade to higher tier plan

**Model Availability Errors**

- **Symptoms**: 404 Not Found or model-specific errors
- **Causes**: Model not available, incorrect model name, region restrictions
- **Solutions**:
  - Verify model name spelling
  - Check model availability in your region
  - Update to supported model versions

**Network Connectivity Errors**

- **Symptoms**: Timeout errors, connection refused
- **Causes**: Network issues, firewall blocking, DNS problems
- **Solutions**:
  - Check network connectivity
  - Verify firewall rules
  - Test DNS resolution

### Channel Optimization

**Performance Tuning**

**Priority Optimization**

- **Strategy**: Set priorities based on cost, performance, and reliability
- **Best Practice**: Use fastest/cheapest channels with highest priority
- **Monitoring**: Track performance metrics to adjust priorities

**Weight Distribution**

- **Purpose**: Balance load among channels with same priority
- **Calculation**: Distribute requests proportionally to weights
- **Optimization**: Adjust weights based on channel capacity and performance

**Model Selection**

- **Cost Optimization**: Use appropriate models for different use cases
- **Performance Optimization**: Select models based on response time requirements
- **Quality Optimization**: Choose models based on output quality needs

**Capacity Planning**

**Scaling Strategies**

- **Horizontal Scaling**: Add more channels for the same provider
- **Vertical Scaling**: Upgrade to higher-tier plans with better limits
- **Geographic Distribution**: Use channels in different regions for better performance

**Quota Management**

- **Monitoring**: Track usage against quotas and limits
- **Alerting**: Set up alerts before reaching quota limits
- **Planning**: Forecast usage growth and plan capacity accordingly

## Debug Panel: Troubleshooting Made Easy

### What is the Debug Panel?

The Debug Panel is a **built-in diagnostic tool** that helps administrators quickly identify and resolve channel configuration issues, particularly related to model pricing and data migration.

### Accessing the Debug Panel

1. **Navigate** to any channel edit page
2. **Click** the "Debug" button in the page header
3. **View** comprehensive diagnostic information
4. **Take action** with one-click fixes

### Debug Panel Features

#### **📊 Migration Status Overview**

- **Channel Information**: ID, name, and type
- **Data Status**: Unified vs legacy format indicators
- **Model Lists**: Shows which models are configured
- **Status Indicators**: Color-coded migration status

#### **🔍 Diagnostic Information**

- **Model Configs**: Lists models in unified format
- **Legacy Data**: Shows any remaining old-format data
- **Data Conflicts**: Identifies mixed or inconsistent data
- **Validation Results**: Reports configuration errors

#### **🛠️ One-Click Actions**

- **Fix Channel**: Automatically resolve common issues
- **Refresh Status**: Update diagnostic information
- **Log Debug Info**: Generate detailed logs for support
- **Clean Mixed Data**: Remove conflicting configurations

### Common Issues and Solutions

#### **🔧 Mixed Model Data**

**Problem**: Channel shows models from wrong provider (e.g., Claude models in OpenAI channel)

**Solution**:

1. Open Debug Panel
2. Check migration status
3. Click "Fix Channel" to restore correct models
4. Refresh page to verify fix

#### **⚠️ Migration Issues**

**Problem**: Channel shows "Needs Migration" status

**Solution**:

1. Use Debug Panel to identify specific issues
2. Click "Fix Channel" for automatic resolution
3. Manually verify model configurations
4. Contact support if issues persist

#### **❌ Invalid JSON**

**Problem**: Model Configs field shows validation errors

**Solution**:

1. Check JSON syntax using built-in validator
2. Use "Format JSON" button to fix formatting
3. Use "Load Default" to restore working configuration
4. Refer to documentation for correct format

## Best Practices

### Channel Naming and Organization

**Naming Conventions**

- **Provider-Environment-Purpose**: `OpenAI-Prod-GPT4`, `Claude-Dev-Testing`
- **Include Model Type**: `Azure-GPT35-Customer-Support`, `OpenAI-GPT4-Content-Generation`
- **Environment Indicators**: Use suffixes like `-prod`, `-staging`, `-dev`
- **Avoid Generic Names**: Instead of "Channel 1", use "OpenAI-Production-Primary"

**Organization Strategies**

- **Group by Environment**: Separate production, staging, and development channels
- **Group by Use Case**: Different channels for different applications or teams
- **Group by Performance**: High-priority channels for critical applications
- **Group by Cost**: Separate channels for different billing or cost centers

### Security Best Practices

**API Key Management**

- **Rotation Schedule**: Regularly rotate API keys (monthly or quarterly)
- **Principle of Least Privilege**: Use organization-scoped keys when possible
- **Secure Storage**: Never store API keys in code or configuration files
- **Access Logging**: Monitor API key usage for suspicious activity

**Access Control**

- **Group-Based Access**: Use groups to control channel access
- **Regular Audits**: Review group memberships and access permissions
- **Temporary Access**: Use time-limited access for contractors or temporary users
- **Documentation**: Maintain records of who has access to which channels

### Performance Optimization

**Priority Configuration**

```
High Priority (1-10): Critical production workloads
Medium Priority (11-50): Standard applications
Low Priority (51-100): Development and testing
```

**Load Balancing Strategy**

- **Primary-Secondary**: Use weight 10 for primary, weight 1 for backup
- **Equal Distribution**: Use equal weights for channels with similar performance
- **Cost-Based**: Higher weights for cheaper channels
- **Performance-Based**: Higher weights for faster channels

**Model Selection Guidelines**

- **Cost-Sensitive Applications**: Use GPT-3.5-turbo or Claude Haiku
- **Quality-Critical Applications**: Use GPT-4 or Claude Opus
- **High-Volume Applications**: Use embedding models or smaller language models
- **Real-Time Applications**: Prioritize channels with lowest latency

### Model Configs Best Practices

**Pricing Strategy**

**Tiered Pricing Example**

```json
{
  "gpt-3.5-turbo": {
    "ratio": 1.0,
    "completion_ratio": 2.0,
    "max_tokens": 16385
  },
  "gpt-4": {
    "ratio": 15.0,
    "completion_ratio": 2.0,
    "max_tokens": 8192
  },
  "gpt-4-turbo": {
    "ratio": 8.0,
    "completion_ratio": 2.0,
    "max_tokens": 128000
  }
}
```

**Cost Control Measures**

- **Token Limits**: Set conservative max_tokens for cost control
- **Rate Limiting**: Use rate_limit_rpm and rate_limit_tpm to prevent abuse
- **Model Restrictions**: Only enable necessary models for each channel
- **Usage Monitoring**: Regularly review usage patterns and costs

**Configuration Validation**

- **JSON Syntax**: Always validate JSON before saving
- **Reasonable Values**: Ensure ratios and limits make business sense
- **Model Compatibility**: Verify max_tokens don't exceed model capabilities
- **Testing**: Test configurations with small requests before full deployment

### Monitoring and Maintenance

**Regular Health Checks**

- **Daily**: Review channel status and error rates
- **Weekly**: Analyze performance trends and usage patterns
- **Monthly**: Review costs and optimize pricing configurations
- **Quarterly**: Audit access permissions and security settings

**Proactive Monitoring**

**Key Metrics to Track**

- **Availability**: Channel uptime percentage
- **Performance**: Average response times and error rates
- **Usage**: Request volume and token consumption
- **Costs**: Spending trends and budget adherence

**Alert Configuration**

- **Channel Down**: Immediate alerts when channels go offline
- **High Error Rate**: Alerts when error rate exceeds 5%
- **Cost Threshold**: Alerts when spending exceeds budget
- **Performance Degradation**: Alerts when response times increase significantly

### Disaster Recovery and Backup

**Redundancy Planning**

- **Multiple Providers**: Don't rely on a single AI provider
- **Geographic Distribution**: Use channels in different regions
- **Capacity Planning**: Ensure backup channels can handle full load
- **Failover Testing**: Regularly test failover scenarios

**Configuration Backup**

- **Export Configurations**: Regularly backup channel configurations
- **Version Control**: Track configuration changes over time
- **Documentation**: Maintain documentation of configuration decisions
- **Recovery Procedures**: Document steps to restore channels after failures

### Troubleshooting Workflow

**Systematic Approach**

1. **Check Debug Panel**: Use built-in diagnostics first
2. **Review Recent Changes**: Check if issues started after configuration changes
3. **Test Connectivity**: Verify network connectivity and API endpoints
4. **Check Provider Status**: Verify if the AI provider is experiencing issues
5. **Review Logs**: Examine detailed error logs for specific error messages
6. **Escalate if Needed**: Contact support with comprehensive diagnostic information

**Common Resolution Steps**

- **Authentication Issues**: Regenerate and update API keys
- **Rate Limiting**: Reduce request frequency or upgrade provider plan
- **Model Errors**: Verify model names and availability
- **Network Issues**: Check firewall rules and DNS resolution
- **Configuration Errors**: Use Debug Panel's "Fix Channel" feature

## Advanced Features

### Batch Operations

- **Multiple channel management** through the channels list
- **Bulk configuration updates** for similar channels
- **Mass migration tools** for system upgrades

### API Integration

- **RESTful endpoints** for programmatic management
- **Webhook support** for real-time notifications
- **Monitoring APIs** for system integration

### Security Features

- **Encrypted credential storage** for API keys
- **Access logging** for audit trails
- **Permission-based access** to sensitive settings

## Support and Resources

### Getting Help

- **Debug Panel**: First line of troubleshooting
- **Application logs**: Detailed error information
- **Documentation**: Comprehensive guides and examples
- **Community support**: Forums and discussion groups

### Additional Resources

- **API Documentation**: Technical reference for developers
- **Migration Guide**: Detailed upgrade instructions
- **Best Practices**: Recommended configurations
- **Troubleshooting FAQ**: Common issues and solutions

## User Interface Enhancements

### Interactive Help System

The channel configuration interface includes comprehensive tooltips to help users understand each field:

**Tooltip Features**

- **Question Mark Icons**: Hover over ❓ icons next to field labels for detailed explanations
- **Context-Sensitive Help**: Each tooltip provides specific guidance for that field
- **Multi-Language Support**: Available in both English and Chinese
- **Consistent Design**: Uniform tooltip styling across all three frontend templates

**Available Tooltips**

| Field | Description |
|-------|-------------|
| **Rate Limit** | Explains the 3-minute window, usage scenarios, and abuse prevention |
| **Model Mapping** | Shows JSON format examples and model redirection use cases |
| **Model Configs** | Details pricing structure, field meanings, and configuration options |
| **System Prompt** | Describes forced prompt behavior and setup requirements |

**Implementation Details**

- **Default Template**: Uses Semantic UI React `Popup` component
- **Berry Template**: Uses Material-UI `Tooltip` component
- **Air Template**: Uses Semi Design `Tooltip` component
- **Accessibility**: All tooltips include proper ARIA attributes and keyboard navigation

### Form Validation and Feedback

**Real-Time Validation**

- **JSON Syntax**: Immediate feedback for JSON fields with color-coded borders
- **Required Fields**: Clear indication of mandatory vs optional fields
- **Format Validation**: Specific validation for different field types

**Visual Indicators**

- **Success States**: Green borders and checkmarks for valid input
- **Error States**: Red borders and error messages for invalid input
- **Loading States**: Progress indicators during form submission

## Quick Reference

### Essential Configuration Checklist

**New Channel Setup**

1. ✅ **Basic Information**: Set descriptive name and correct provider type
2. ✅ **Authentication**: Configure valid API key and base URL
3. ✅ **Models**: Specify available models or leave empty for all
4. ✅ **Model Configs**: Use "Load Default" then customize pricing
5. ✅ **Access Control**: Set appropriate groups if needed
6. ✅ **Testing**: Use test button to verify connectivity
7. ✅ **Priority**: Set appropriate priority for load balancing

**Regular Maintenance**

- 🔄 **Weekly**: Check channel status and performance metrics
- 🔄 **Monthly**: Review and optimize Model Configs pricing
- 🔄 **Quarterly**: Rotate API keys and audit access permissions
- 🔄 **As Needed**: Use Debug Panel for troubleshooting

### Common Configuration Patterns

**Production Setup**

```json
{
  "name": "OpenAI-Production-Primary",
  "type": 1,
  "priority": 1,
  "weight": 10,
  "model_configs": {
    "gpt-3.5-turbo": {
      "ratio": 1.0,
      "completion_ratio": 2.0,
      "max_tokens": 16385
    },
    "gpt-4": {
      "ratio": 15.0,
      "completion_ratio": 2.0,
      "max_tokens": 8192
    }
  }
}
```

**Development Setup**

```json
{
  "name": "OpenAI-Development",
  "type": 1,
  "priority": 50,
  "weight": 1,
  "groups": "developers,qa",
  "model_configs": {
    "gpt-3.5-turbo": {
      "ratio": 0.5,
      "completion_ratio": 2.0,
      "max_tokens": 4096
    }
  }
}
```

### Troubleshooting Quick Guide

| Issue                    | Symptoms             | Quick Fix                     |
| ------------------------ | -------------------- | ----------------------------- |
| **Authentication Error** | 401 responses        | Check API key validity        |
| **Rate Limiting**        | 429 responses        | Reduce request frequency      |
| **Model Not Found**      | 404 responses        | Verify model name spelling    |
| **Mixed Model Data**     | Wrong models showing | Use Debug Panel → Fix Channel |
| **Invalid JSON**         | Validation errors    | Use "Format JSON" button      |
| **Channel Offline**      | No responses         | Check provider status page    |

### Support Escalation Path

1. **Self-Service**: Use Debug Panel for immediate diagnostics
2. **Documentation**: Consult this guide and API documentation
3. **Logs**: Check application logs for detailed error information
4. **Community**: Search forums and community discussions
5. **Support**: Contact technical support with Debug Panel output

---

_For technical support or questions about channel configuration, please refer to the Debug Panel first, then consult the application logs or contact your system administrator._
