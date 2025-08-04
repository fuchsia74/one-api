import { Label, Icon } from 'semantic-ui-react';
import { timestamp2string, renderQuota } from '../../helpers';

// Common rendering utilities for tables

// Clean display helper - returns empty string for null/undefined/empty values
export const cleanDisplay = (value, fallback = '') => {
  if (value === null || value === undefined || value === '') {
    return fallback;
  }
  return value;
};

// Clean display for numbers - returns empty string for 0, null, undefined
export const cleanDisplayNumber = (value, fallback = '') => {
  if (value === null || value === undefined || value === 0 || value === '') {
    return fallback;
  }
  return value;
};

export const renderTimestamp = (timestamp, compact = false, onCopy = null, title = null) => {
  const fullTimestamp = timestamp2string(timestamp);
  // Only slice if it's a valid timestamp (not N/A or Invalid Date)
  const displayTimestamp = compact && fullTimestamp.length > 10 && fullTimestamp.includes('-')
    ? fullTimestamp.slice(5)
    : fullTimestamp;

  return (
    <code
      onClick={onCopy}
      className={onCopy ? "cursor-pointer" : ""}
      title={title || (onCopy ? 'Click to copy' : fullTimestamp)}
      style={{
        fontSize: '0.85em',
        padding: '2px 4px',
        background: 'rgba(0,0,0,0.05)',
        borderRadius: '3px',
        border: '1px solid rgba(0,0,0,0.1)'
      }}
    >
      {displayTimestamp}
    </code>
  );
};

export const renderColorLabel = (text, color = null) => {
  // Generate a consistent color based on text hash if no color provided
  if (!color) {
    const hash = text.split('').reduce((a, b) => {
      a = ((a << 5) - a) + b.charCodeAt(0);
      return a & a;
    }, 0);
    const colors = ['blue', 'green', 'orange', 'purple', 'pink', 'teal', 'violet'];
    color = colors[Math.abs(hash) % colors.length];
  }

  return (
    <Label
      size="small"
      color={color}
      style={{
        margin: '1px',
        fontSize: '0.8em'
      }}
    >
      {text}
    </Label>
  );
};

export const renderStatusLabel = (status, statusMap, defaultColor = 'grey', t = null) => {
  const statusInfo = statusMap[status] || {
    text: t ? t('common.unknown') : 'Unknown',
    color: defaultColor
  };

  return (
    <Label basic color={statusInfo.color} size="small">
      {statusInfo.icon && <Icon name={statusInfo.icon} />}
      {t ? t(statusInfo.text) : statusInfo.text}
    </Label>
  );
};

export const renderModelTags = (models, maxDisplay = 3) => {
  if (!models || models.length === 0) {
    return <Label size="mini" basic>None</Label>;
  }

  const modelArray = Array.isArray(models) ? models :
    (typeof models === 'string' ? JSON.parse(models || '[]') : []);

  const displayModels = modelArray.slice(0, maxDisplay);
  const remainingCount = modelArray.length - maxDisplay;

  return (
    <div style={{ display: 'flex', flexWrap: 'wrap', gap: '2px' }}>
      {displayModels.map((model, index) => renderColorLabel(model))}
      {remainingCount > 0 && (
        <Label size="mini" basic>
          +{remainingCount} more
        </Label>
      )}
    </div>
  );
};

export const renderBalance = (type, balance, currency = '$', t = null) => {
  if (balance === undefined || balance === null) {
    return <span>{t ? t('common.unknown') : 'Unknown'}</span>;
  }

  if (balance === 0) {
    return <span>{t ? t('balance_not_supported') : 'Not supported'}</span>;
  }

  // Format based on type or use default currency
  const formatMap = {
    1: (b) => `$${b.toFixed(2)}`, // OpenAI - USD
    4: (b) => `¥${b.toFixed(2)}`, // CloseAI - CNY
    5: (b) => `¥${(b / 10000).toFixed(2)}`, // OpenAI-SB - CNY (scaled)
    8: (b) => `$${b.toFixed(2)}`, // Custom - USD
    10: (b) => b.toLocaleString(), // AI Proxy - number
    12: (b) => `¥${b.toFixed(2)}`, // API2GPT - CNY
    13: (b) => b.toLocaleString(), // AIGC2D - number
    20: (b) => `$${b.toFixed(2)}`, // OpenRouter - USD
    36: (b) => `¥${b.toFixed(2)}`, // DeepSeek - CNY
    44: (b) => `¥${b.toFixed(2)}`, // SiliconFlow - CNY
  };

  const formatter = formatMap[type];
  if (formatter) {
    return <span>{formatter(balance)}</span>;
  }

  // Default formatting
  return <span>{currency}{balance.toFixed(2)}</span>;
};

export const renderLatency = (elapsedTime, t = null) => {
  if (!elapsedTime) {
    return <span style={{ color: '#999' }}>{t ? t('common.unknown') : 'N/A'}</span>;
  }

  let color = '#28a745'; // green
  if (elapsedTime > 5000) color = '#dc3545'; // red
  else if (elapsedTime > 2000) color = '#ffc107'; // yellow

  return (
    <span style={{ color, fontWeight: '500' }}>
      {elapsedTime} ms
    </span>
  );
};

export const renderQuotaDisplay = (quota, precision = 6, t = null) => {
  if (!quota) {
    return <Label size="small" basic>{t ? t('common.free') : 'Free'}</Label>;
  }

  return (
    <span style={{ fontFamily: 'monospace', fontSize: '0.9em' }}>
      {renderQuota(quota, t, precision)}
    </span>
  );
};

export const renderResponseTime = (responseTime) => {
  if (!responseTime) {
    return <span style={{ color: '#999' }}>Not tested</span>;
  }

  let color = '#28a745'; // green
  if (responseTime > 3000) color = '#dc3545'; // red
  else if (responseTime > 1000) color = '#ffc107'; // yellow

  return (
    <span style={{ color, fontWeight: '500' }}>
      {responseTime} ms
    </span>
  );
};

export const renderPriority = (priority) => {
  let color = 'grey';
  if (priority > 0) color = 'green';
  else if (priority < 0) color = 'red';

  return (
    <Label
      basic
      color={color}
      size="small"
      style={{ minWidth: '45px', textAlign: 'center' }}
    >
      {priority}
    </Label>
  );
};

export const renderUsageStats = (usage) => {
  if (!usage) return null;

  return (
    <div style={{ fontSize: '0.85em', lineHeight: '1.2' }}>
      {usage.prompt_tokens && (
        <div>
          <Icon name="comment" size="small" />
          {usage.prompt_tokens}
        </div>
      )}
      {usage.completion_tokens && (
        <div>
          <Icon name="reply" size="small" />
          {usage.completion_tokens}
        </div>
      )}
    </div>
  );
};

export const renderDetailContent = (content, maxLength = 100) => {
  if (!content) return null;

  const truncated = content.length > maxLength ?
    content.substring(0, maxLength) + '...' : content;

  return (
    <div style={{
      maxWidth: '300px',
      wordBreak: 'break-word',
      fontSize: '0.9em',
      lineHeight: '1.3'
    }}>
      {truncated}
      {content.length > maxLength && (
        <div style={{ marginTop: '4px' }}>
          <Label size="mini" basic pointing="up">
            {content.length - maxLength} more chars
          </Label>
        </div>
      )}
    </div>
  );
};

// Common status maps
export const CHANNEL_STATUS_MAP = {
  1: { text: 'channel.table.status_enabled', color: 'green', icon: 'check circle' },
  2: { text: 'channel.table.status_disabled', color: 'red', icon: 'times circle' },
  3: { text: 'channel.table.status_auto_disabled', color: 'yellow', icon: 'warning' }
};

export const LOG_TYPE_MAP = {
  1: { text: 'log.type.topup', color: 'green', icon: 'plus' },
  2: { text: 'log.type.usage', color: 'blue', icon: 'minus' },
  3: { text: 'log.type.admin', color: 'orange', icon: 'cog' },
  4: { text: 'log.type.system', color: 'purple', icon: 'server' },
  5: { text: 'log.type.test', color: 'violet', icon: 'lab' }
};

export const TOKEN_STATUS_MAP = {
  1: { text: 'token.status.enabled', color: 'green', icon: 'check circle' },
  2: { text: 'token.status.disabled', color: 'red', icon: 'times circle' },
  3: { text: 'token.status.expired', color: 'grey', icon: 'clock outline' }
};

export const USER_STATUS_MAP = {
  1: { text: 'user.status.enabled', color: 'green', icon: 'check circle' },
  2: { text: 'user.status.disabled', color: 'red', icon: 'times circle' },
  3: { text: 'user.status.pending', color: 'yellow', icon: 'clock outline' }
};
