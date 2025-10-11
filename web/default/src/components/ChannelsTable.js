import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Button,
  Form,
  Icon,
  Label,
  Popup,
  Table,
} from 'semantic-ui-react';
import { Link } from 'react-router-dom';
import {
  API,
  showError,
  showSuccess,
  timestamp2string,
} from '../helpers';
import { CHANNEL_OPTIONS, ITEMS_PER_PAGE } from '../constants';
import { renderGroup } from '../helpers/render';
import { cleanDisplay } from './shared/tableUtils';
import BaseTable from './shared/BaseTable';

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

const ChannelsTable = () => {
  const { t } = useTranslation();
  const [channels, setChannels] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [searching, setSearching] = useState(false);
  const [sortBy, setSortBy] = useState('');
  const [sortOrder, setSortOrder] = useState('desc');

  const SORT_OPTIONS = [
    { key: '', text: t('channels.sort.default', 'Default'), value: '' },
    { key: 'id', text: t('channels.sort.id', 'ID'), value: 'id' },
    { key: 'name', text: t('channels.sort.name', 'Name'), value: 'name' },
    { key: 'type', text: t('channels.sort.type', 'Type'), value: 'type' },
    { key: 'status', text: t('channels.sort.status', 'Status'), value: 'status' },
    { key: 'response_time', text: t('channels.sort.response_time', 'Response Time'), value: 'response_time' },
    { key: 'created_time', text: t('channels.sort.created_time', 'Created Time'), value: 'created_time' },
  ];

  const loadChannels = async (page = 0, sortBy = '', sortOrder = 'desc') => {
    setLoading(true);
    let url = `/api/channel/?p=${page}&size=${ITEMS_PER_PAGE}`;
    if (sortBy) {
      url += `&sort=${sortBy}&order=${sortOrder}`;
    }
    const res = await API.get(url);
    const { success, message, data, total } = res.data;
    if (success) {
      const processedChannels = (data || []).map(processChannelData);
      setChannels(processedChannels);
      const computedTotal = typeof total === 'number' ? total : processedChannels.length;
      setTotalPages(Math.max(1, Math.ceil(computedTotal / ITEMS_PER_PAGE)));
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
    loadChannels(activePage - 1, sortBy, sortOrder);
  };

  useEffect(() => {
    loadChannels(0, sortBy, sortOrder)
      .then()
      .catch((reason) => {
        showError(reason);
      });
  }, [sortBy, sortOrder]); // eslint-disable-line react-hooks/exhaustive-deps

  const manageChannel = async (id, action, idx) => {
    let res;
    switch (action) {
      case 'delete':
        res = await API.delete(`/api/channel/${id}`);
        break;
      case 'enable':
        res = await API.put('/api/channel/?status_only=1', { id: id, status: 1 });
        break;
      case 'disable':
        res = await API.put('/api/channel/?status_only=1', { id: id, status: 2 });
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
      await loadChannels(0, sortBy, sortOrder);
      setActivePage(1);
      return;
    }
    setSearching(true);
    let url = `/api/channel/search?keyword=${searchKeyword}`;
    if (sortBy) {
      url += `&sort=${sortBy}&order=${sortOrder}`;
    }
    const res = await API.get(url);
    const { success, message, data } = res.data;
    if (success) {
      const processedChannels = (data || []).map(processChannelData);
      setChannels(processedChannels);
      const total = processedChannels.length;
      setTotalPages(Math.max(1, Math.ceil(total / ITEMS_PER_PAGE)));
      setActivePage(1);
    } else {
      showError(message);
    }
    setSearching(false);
  };

  const sortChannel = async (key) => {
    const newSortOrder = sortBy === key && sortOrder === 'desc' ? 'asc' : 'desc';
    setSortBy(key);
    setSortOrder(newSortOrder);
    await loadChannels(activePage - 1, key, newSortOrder);
  };

  const getSortIcon = (columnKey) => {
    if (sortBy !== columnKey) {
      return <Icon name="sort" style={{ opacity: 0.5 }} />;
    }
    return <Icon name={sortOrder === 'asc' ? 'sort up' : 'sort down'} />;
  };

  const handleSortChange = async (e, { value }) => {
    setSortBy(value);
    setSortOrder('desc');
    setActivePage(1);
    await loadChannels(0, value, 'desc');
  };

  const handleKeywordChange = async (e, { value }) => {
    setSearchKeyword(value.trim());
  };

  const refresh = async () => {
    setLoading(true);
    await loadChannels(0, sortBy, sortOrder);
    setActivePage(1);
  };

  const headerCells = [
    {
      content: (
        <>
          ID {getSortIcon('id')}
        </>
      ),
      sortable: true,
      onClick: () => sortChannel('id'),
    },
    {
      content: (
        <>
          {t('name')} {getSortIcon('name')}
        </>
      ),
      sortable: true,
      onClick: () => sortChannel('name'),
    },
    {
      content: (
        <>
          {t('type')} {getSortIcon('type')}
        </>
      ),
      sortable: true,
      onClick: () => sortChannel('type'),
    },
    {
      content: (
        <>
          {t('status')} {getSortIcon('status')}
        </>
      ),
      sortable: true,
      onClick: () => sortChannel('status'),
    },
    {
      content: (
        <>
          {t('response_time')} {getSortIcon('response_time')}
        </>
      ),
      sortable: true,
      onClick: () => sortChannel('response_time'),
    },
    {
      content: (
        <>
          {t('created_time')} {getSortIcon('created_time')}
        </>
      ),
      sortable: true,
      onClick: () => sortChannel('created_time'),
    },
    {
      content: t('actions'),
      sortable: false,
    },
  ];

  const footerButtons = [
    {
      content: t('common.add'),
      as: Link,
      to: '/channel/add',
      loading: loading,
    },
    {
      content: t('common.refresh'),
      onClick: refresh,
      loading: loading,
    },
  ];

  return (
    <>
      <Form onSubmit={searchChannels}>
        <Form.Group>
          <Form.Input
            width={12}
            icon='search'
            iconPosition='left'
            placeholder={t('channels.search.placeholder', 'Search channels...')}
            value={searchKeyword}
            loading={searching}
            onChange={handleKeywordChange}
          />
          <Form.Dropdown
            width={4}
            selection
            placeholder={t('channels.sort.placeholder', 'Sort by...')}
            options={SORT_OPTIONS}
            value={sortBy}
            onChange={handleSortChange}
          />
        </Form.Group>
      </Form>

      <BaseTable
        loading={loading}
        activePage={activePage}
        totalPages={totalPages}
        onPageChange={onPaginationChange}
        headerCells={headerCells}
        footerButtons={footerButtons}
        colSpan={7}
      >
        {channels.map((channel, idx) => {
          if (channel.deleted) return null;
          return (
            <Table.Row key={channel.id}>
              <Table.Cell data-label="ID">
                <Label circular>
                  {channel.id}
                </Label>
              </Table.Cell>
              <Table.Cell data-label="Name">
                <div>
                  <strong>{cleanDisplay(channel.name)}</strong>
                  {channel.group && (
                    <div style={{ fontSize: '0.9em', color: '#666' }}>
                      {renderGroup(channel.group)}
                    </div>
                  )}
                </div>
              </Table.Cell>
              <Table.Cell data-label="Type">{renderChannelType(channel.type, t)}</Table.Cell>
              <Table.Cell data-label="Status">{renderChannelStatus(channel.status, channel.priority)}</Table.Cell>
              <Table.Cell data-label="Response Time">{renderResponseTime(channel.response_time)}</Table.Cell>
              <Table.Cell data-label="Created Time">
                {renderTimestamp(channel.created_time)}
              </Table.Cell>
              <Table.Cell data-label="Actions">
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
      </BaseTable>
    </>
  );
};

export default ChannelsTable;
