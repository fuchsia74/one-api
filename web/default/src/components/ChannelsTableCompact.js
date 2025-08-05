import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Button,
  Form,
  Label,
  Pagination,
  Popup,
  Table,
  Icon,
} from 'semantic-ui-react';
import { Link } from 'react-router-dom';
import {
  API,
  showError,
  showSuccess,
  timestamp2string,
} from '../helpers';
import { CHANNEL_OPTIONS } from '../constants';
import { renderGroup, renderNumber } from '../helpers/render';
import { cleanDisplay } from './shared/tableUtils';

import { ITEMS_PER_PAGE } from '../constants';

function renderTimestamp(timestamp) {
  return <>{timestamp2string(timestamp)}</>;
}

function getChannelTypeConfig(type) {
  const config = CHANNEL_OPTIONS.find(option => option.value === type) || {
    value: 0,
    text: 'Unknown',
    color: 'grey'
  };

  const iconMap = {
    1: 'openai', 2: 'microsoft', 3: 'google', 4: 'anthropic', 5: 'openai',
    6: 'amazon', 7: 'baidu', 8: 'setting', 9: 'zhipu', 10: 'robot',
    11: 'alibaba', 12: 'tencent', 13: 'bytedance', 14: 'meta',
    15: 'stability', 16: 'cohere', 17: 'mistral', 18: 'perplexity',
    19: 'groq', 20: 'openrouter', 21: 'together'
  };

  return {
    ...config,
    icon: iconMap[type] || 'server'
  };
}

function renderChannelType(type, t) {
  const config = getChannelTypeConfig(type);

  return (
    <Label basic color={config.color}>
      <Icon name={config.icon} />
      {config.text}
    </Label>
  );
}

function renderChannelStatus(status, priority) {
  let color = 'green';
  let icon = 'check circle';
  let text = 'Active';

  if (status === 2) {
    color = 'red';
    icon = 'times circle';
    text = 'Disabled';
  } else if (priority < 0) {
    color = 'orange';
    icon = 'pause circle';
    text = 'Paused';
  }

  return (
    <Label basic color={color}>
      <Icon name={icon} />
      {text}
    </Label>
  );
}

function renderBalance(type, balance, t) {
  if (balance === null || balance === undefined) {
    return '';
  }

  const formatters = {
    1: (bal) => `$${bal.toFixed(2)}`, // OpenAI
    4: (bal) => `¥${bal.toFixed(2)}`, // CloseAI
    5: (bal) => `¥${(bal / 10000).toFixed(2)}`, // OpenAI-SB
    8: (bal) => `$${bal.toFixed(2)}`, // Custom
    10: (bal) => renderNumber(bal), // AI Proxy
    12: (bal) => `¥${bal.toFixed(2)}`, // API2GPT
    13: (bal) => renderNumber(bal), // AIGC2D
    20: (bal) => `$${bal.toFixed(2)}`, // OpenRouter
    36: (bal) => `¥${bal.toFixed(2)}`, // DeepSeek
    44: (bal) => `¥${bal.toFixed(2)}`, // SiliconFlow
  };

  const formatter = formatters[type];

  if (!formatter) {
    return '';
  }

  const formattedBalance = formatter(balance);
  let color = 'green';

  // Determine color based on balance level
  if (balance <= 0) {
    color = 'red';
  } else if (balance < 10) {
    color = 'orange';
  } else if (balance < 100) {
    color = 'yellow';
  }

  return (
    <Label color={color}>
      <Icon name="dollar sign" />
      {formattedBalance}
    </Label>
  );
}

function renderResponseTime(responseTime) {
  if (!responseTime) return '';

  let color = 'green';
  if (responseTime > 5000) color = 'red';
  else if (responseTime > 2000) color = 'orange';
  else if (responseTime > 1000) color = 'yellow';

  return (
    <Label size="mini" color={color}>
      <Icon name="clock" />
      {responseTime}ms
    </Label>
  );
}

const ChannelsTableCompact = () => {
  const { t } = useTranslation();
  const [channels, setChannels] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [searching, setSearching] = useState(false);

  const loadChannels = async (page = 0) => {
    setLoading(true);
    const res = await API.get(`/api/channel/?p=${page}`);
    const { success, message, data, total } = res.data;
    if (success) {
      const processedChannels = (data || []).map(processChannelData);
      setChannels(processedChannels);
      setTotalPages(Math.ceil(total / ITEMS_PER_PAGE));
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const processChannelData = (channel) => {
    if (channel.models === '') {
      channel.models = [];
      channel.test_model = '';
    } else {
      try {
        if (typeof channel.models === 'string') {
          channel.models = JSON.parse(channel.models);
        }
      } catch (e) {
        channel.models = [];
      }
    }

    if (!channel.models || !Array.isArray(channel.models)) {
      channel.models = [];
    }

    return channel;
  };

  const onPaginationChange = (e, { activePage }) => {
    setActivePage(activePage);
    loadChannels(activePage - 1);
  };

  useEffect(() => {
    loadChannels(0)
      .then()
      .catch((reason) => {
        showError(reason);
      });
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const manageChannel = async (id, action, idx) => {
    let res;
    switch (action) {
      case 'delete':
        res = await API.delete(`/api/channel/${id}`);
        break;
      case 'enable':
        res = await API.put('/api/channel/', { id: id, status: 1 });
        break;
      case 'disable':
        res = await API.put('/api/channel/', { id: id, status: 2 });
        break;
      case 'test':
        res = await API.get(`/api/channel/test/${id}`);
        const { success: testSuccess, message: testMessage, time } = res.data;
        if (testSuccess) {
          showSuccess(`Test successful!${time ? ` Response time: ${time}ms` : ''}`);
          // Update the channel's response time
          let newChannels = [...channels];
          newChannels[idx].response_time = time;
          setChannels(newChannels);
        } else {
          showError(testMessage);
        }
        return;
      default:
        return;
    }
    const { success, message } = res.data;
    if (success) {
      showSuccess('Operation successful');
      let newChannels = [...channels];
      if (action === 'delete') {
        newChannels[idx].deleted = true;
      } else if (action === 'enable') {
        newChannels[idx].status = 1;
      } else if (action === 'disable') {
        newChannels[idx].status = 2;
      }
      setChannels(newChannels);
    } else {
      showError(message);
    }
  };

  const searchChannels = async () => {
    if (searchKeyword === '') {
      // if keyword is blank, load channels instead.
      await loadChannels(0);
      setActivePage(1);
      return;
    }
    setSearching(true);
    const res = await API.get(
      `/api/channel/search?keyword=${searchKeyword}`
    );
    const { success, message, data } = res.data;
    if (success) {
      const processedChannels = (data || []).map(processChannelData);
      setChannels(processedChannels);
      setActivePage(1);
      setTotalPages(Math.ceil(data.length / ITEMS_PER_PAGE));
    } else {
      showError(message);
    }
    setSearching(false);
  };

  const handleKeywordChange = async (e, { value }) => {
    setSearchKeyword(value.trim());
  };

  const sortChannel = (key) => {
    if (channels.length === 0) return;
    setLoading(true);
    let sortedChannels = [...channels];
    sortedChannels.sort((a, b) => {
      if (!isNaN(a[key])) {
        // If the value is numeric, subtract to sort
        return a[key] - b[key];
      } else {
        // If the value is not numeric, sort as strings
        return ('' + a[key]).localeCompare(b[key]);
      }
    });
    if (sortedChannels[0].id === channels[0].id) {
      sortedChannels.reverse();
    }
    setChannels(sortedChannels);
    setLoading(false);
  };

  const refresh = async () => {
    setLoading(true);
    await loadChannels(0);
    setActivePage(1);
  };

  return (
    <>
      <Form onSubmit={searchChannels}>
        <Form.Input
          icon='search'
          fluid
          iconPosition='left'
          placeholder={t('channel.search_placeholder')}
          value={searchKeyword}
          loading={searching}
          onChange={handleKeywordChange}
        />
      </Form>

      <Table basic={'very'} compact size='small'>
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortChannel('id');
              }}
            >
              ID
            </Table.HeaderCell>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortChannel('name');
              }}
            >
              {t('name')}
            </Table.HeaderCell>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortChannel('type');
              }}
            >
              {t('type')}
            </Table.HeaderCell>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortChannel('status');
              }}
            >
              {t('status')}
            </Table.HeaderCell>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortChannel('response_time');
              }}
            >
              {t('response_time')}
            </Table.HeaderCell>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortChannel('created_time');
              }}
            >
              {t('created_time')}
            </Table.HeaderCell>
            <Table.HeaderCell>{t('actions')}</Table.HeaderCell>
          </Table.Row>
        </Table.Header>

        <Table.Body>
          {channels.map((channel, idx) => {
              if (channel.deleted) return <></>;
              return (
                <Table.Row key={channel.id}>
                  <Table.Cell>
                    <Label circular>
                      {channel.id}
                    </Label>
                  </Table.Cell>
                  <Table.Cell>
                    <div>
                      <strong>{cleanDisplay(channel.name)}</strong>
                      {channel.group && (
                        <div style={{ fontSize: '0.9em', color: '#666' }}>
                          {renderGroup(channel.group)}
                        </div>
                      )}
                    </div>
                  </Table.Cell>
                  <Table.Cell>{renderChannelType(channel.type, t)}</Table.Cell>
                  <Table.Cell>{renderChannelStatus(channel.status, channel.priority)}</Table.Cell>
                  <Table.Cell>{renderResponseTime(channel.response_time)}</Table.Cell>
                  <Table.Cell>
                    {renderTimestamp(channel.created_time)}
                  </Table.Cell>
                  <Table.Cell>
                    <div>
                      <Button
                        size={'tiny'}
                        color="blue"
                        as={Link}
                        to={`/channel/edit/${channel.id}`}
                      >
                        {t('common.edit')}
                      </Button>
                      <Button
                        size={'tiny'}
                        color="orange"
                        onClick={() => manageChannel(channel.id, 'test', idx)}
                      >
                        {t('channel.test')}
                      </Button>
                      <Button
                        size={'tiny'}
                        onClick={() => {
                          manageChannel(
                            channel.id,
                            channel.status === 1 ? 'disable' : 'enable',
                            idx
                          );
                        }}
                      >
                        {channel.status === 1
                          ? t('channel.disable')
                          : t('channel.enable')}
                      </Button>
                      <Popup
                        trigger={
                          <Button size='tiny' negative>
                            {t('common.delete')}
                          </Button>
                        }
                        on='click'
                        flowing
                        hoverable
                      >
                        <Button
                          negative
                          onClick={() => {
                            manageChannel(channel.id, 'delete', idx);
                          }}
                        >
                          {t('channel.delete_confirm')}
                        </Button>
                      </Popup>
                    </div>
                  </Table.Cell>
                </Table.Row>
              );
            })}
        </Table.Body>

        <Table.Footer>
          <Table.Row>
            <Table.HeaderCell colSpan='7'>
              <Button
                size='small'
                as={Link}
                to='/channel/add'
                loading={loading}
              >
                Add
              </Button>
              <Button size='small' onClick={refresh} loading={loading}>
                Refresh
              </Button>
              <Pagination
                floated='right'
                activePage={activePage}
                onPageChange={onPaginationChange}
                size='small'
                siblingRange={1}
                totalPages={totalPages}
              />
            </Table.HeaderCell>
          </Table.Row>
        </Table.Footer>
      </Table>
    </>
  );
};

export default ChannelsTableCompact;
