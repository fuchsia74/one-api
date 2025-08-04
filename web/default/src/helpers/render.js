import React from 'react';
import { Label, Message } from 'semantic-ui-react';
import { getChannelOption } from './helper';

export function renderText(text, limit) {
  if (text.length > limit) {
    return text.slice(0, limit - 3) + '...';
  }
  return text;
}

export function renderGroup(group) {
  if (group === '') {
    return <Label>default</Label>;
  }
  let groups = group.split(',');
  groups.sort();
  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        flexWrap: 'wrap',
        gap: '2px',
        rowGap: '6px',
      }}
    >
      {groups.map((group) => {
        if (group === 'vip' || group === 'pro') {
          return <Label color='yellow'>{group}</Label>;
        } else if (group === 'svip' || group === 'premium') {
          return <Label color='red'>{group}</Label>;
        }
        return <Label>{group}</Label>;
      })}
    </div>
  );
}

export function renderNumber(num) {
  if (num >= 1000000000) {
    return (num / 1000000000).toFixed(1) + 'B';
  } else if (num >= 1000000) {
    return (num / 1000000).toFixed(1) + 'M';
  } else if (num >= 10000) {
    return (num / 1000).toFixed(1) + 'k';
  } else {
    return num;
  }
}

/**
 * Renders a number with tooltip showing exact value when abbreviated.
 * WARNING: Returns JSX element for abbreviated numbers, plain value otherwise.
 * Only use in React components, not for plain text contexts.
 * @param {number} num - The number to render
 * @returns {JSX.Element|string|number} JSX span with tooltip for large numbers, plain value otherwise
 */
export function renderNumberWithTooltip(num) {
  const abbreviated = renderNumber(num);
  const exact = num.toLocaleString();

  // If the number was abbreviated, return JSX with tooltip
  if (num >= 10000) {
    return (
      <span title={exact} style={{ cursor: 'help', textDecoration: 'underline dotted' }}>
        {abbreviated}
      </span>
    );
  }

  // If not abbreviated, just return the number
  return abbreviated;
}

/**
 * Renders a number for chart tooltips, showing both abbreviated and exact values.
 * Always returns a string, safe for use in any context.
 * @param {number} num - The number to render
 * @returns {string} Abbreviated number with exact value in parentheses for large numbers
 */
export function renderNumberForChart(num) {
  const abbreviated = renderNumber(num);
  const exact = num.toLocaleString();

  // For charts, return the abbreviated version but include exact in a data attribute or similar
  // The chart tooltip can then show both values
  if (num >= 10000) {
    return `${abbreviated} (${exact})`;
  }

  return abbreviated;
}

export function renderQuota(quota, t, precision = 2) {
  const displayInCurrency =
    localStorage.getItem('display_in_currency') === 'true';
  const quotaPerUnit = parseFloat(
    localStorage.getItem('quota_per_unit') || '500000'
  );

  if (displayInCurrency && t) {
    const amount = (quota / quotaPerUnit).toFixed(precision);
    return t('common.quota.display_short', { amount });
  }

  return renderNumber(quota);
}

export function renderQuotaWithPrompt(quota, t) {
  const displayInCurrency =
    localStorage.getItem('display_in_currency') === 'true';
  const quotaPerUnit = parseFloat(
    localStorage.getItem('quota_per_unit') || '500000'
  );

  if (displayInCurrency) {
    const amount = (quota / quotaPerUnit).toFixed(2);
    return ` (${t('common.quota.display', { amount })})`;
  }

  return '';
}

const colors = [
  'red',
  'orange',
  'yellow',
  'olive',
  'green',
  'teal',
  'blue',
  'violet',
  'purple',
  'pink',
  'brown',
  'grey',
  'black',
];

export function renderColorLabel(text) {
  let hash = 0;
  for (let i = 0; i < text.length; i++) {
    hash = text.charCodeAt(i) + ((hash << 5) - hash);
  }
  let index = Math.abs(hash % colors.length);
  return (
    <Label basic color={colors[index]}>
      {text}
    </Label>
  );
}

export function renderChannelTip(channelId) {
  let channel = getChannelOption(channelId);
  if (channel === undefined || channel.tip === undefined) {
    return <></>;
  }
  return (
    <Message>
      <div dangerouslySetInnerHTML={{ __html: channel.tip }}></div>
    </Message>
  );
}
