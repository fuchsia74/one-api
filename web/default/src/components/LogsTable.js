import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Button,
  Form,
  Label,
  Popup,
  Table,
  Select,
  Dropdown,
  Header,
  Segment,
  Icon,
} from 'semantic-ui-react';
import {
  API,
  copy,
  isAdmin,
  showError,
  showSuccess,
  showWarning,
  timestamp2string,
  renderQuota,
} from '../helpers';
import BaseTable from './shared/BaseTable';
import TracingModal from './TracingModal';

import { ITEMS_PER_PAGE } from '../constants';

function ExpandableDetail({ content, isStream, systemPromptReset }) {
  const [expanded, setExpanded] = useState(false);
  const maxLength = 100;
  const shouldTruncate = content && content.length > maxLength;

  return (
    <div style={{ maxWidth: '300px' }}>
      <div style={{
        wordBreak: 'break-word',
        whiteSpace: expanded ? 'normal' : 'nowrap',
        overflow: 'hidden',
        textOverflow: shouldTruncate && !expanded ? 'ellipsis' : 'visible'
      }}>
        {expanded || !shouldTruncate ? content : content.slice(0, maxLength)}
        {shouldTruncate && (
          <Button
            basic
            size="mini"
            style={{ marginLeft: '4px', padding: '2px 6px' }}
            onClick={() => setExpanded(!expanded)}
          >
            {expanded ? 'Show Less' : 'Show More'}
          </Button>
        )}
      </div>
      <div style={{ marginTop: '4px' }}>
        {isStream && (
          <Label size="mini" color="pink" style={{ marginRight: '4px' }}>
            Stream
          </Label>
        )}
        {systemPromptReset && (
          <Label basic size="mini" color="red">
            System Prompt Reset
          </Label>
        )}
      </div>
    </div>
  );
}

function renderTimestamp(timestamp, request_id) {
  const fullTimestamp = timestamp2string(timestamp);

  return (
    <Popup
      content={`${fullTimestamp}${request_id ? `\nRequest ID: ${request_id}` : ''}`}
      trigger={
        <code
          onClick={async () => {
            if (request_id && await copy(request_id)) {
              showSuccess(`Request ID copied: ${request_id}`);
            } else if (request_id) {
              showWarning(`Failed to copy request ID: ${request_id}`);
            }
          }}
          className="timestamp-code"
          style={{ cursor: request_id ? 'pointer' : 'default' }}
        >
          {fullTimestamp}
        </code>
      }
    />
  );
}

function renderLogType(type, t) {
  const typeConfig = {
    1: { color: 'green', text: t('log.type.topup', 'Recharge') },
    2: { color: 'olive', text: t('log.type.usage', 'Consumed') },
    3: { color: 'orange', text: t('log.type.admin', 'Management') },
    4: { color: 'purple', text: t('log.type.system', 'System') },
    5: { color: 'violet', text: t('log.type.test', 'Test') },
  };

  const config = typeConfig[type] || {
    color: 'black',
    text: t('common.unknown', 'Unknown')
  };

  return (
    <Label basic color={config.color}>
      {config.text}
    </Label>
  );
}

const LogsTable = () => {
  const { t } = useTranslation();
  const [logs, setLogs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [searching, setSearching] = useState(false);

  // Tracing modal state
  const [tracingModalOpen, setTracingModalOpen] = useState(false);
  const [selectedLogId, setSelectedLogId] = useState(null);

  // Advanced search filters
  const [showStat, setShowStat] = useState(false);
  const [logType, setLogType] = useState(0);
  const isAdminUser = isAdmin();
  let now = new Date();
  let sevenDaysAgo = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);

  const [inputs, setInputs] = useState({
    username: '',
    token_name: '',
    model_name: '',
    start_timestamp: timestamp2string(sevenDaysAgo.getTime() / 1000),
    end_timestamp: timestamp2string(now.getTime() / 1000 + 3600),
    channel: '',
  });

  const {
    username,
    token_name,
    model_name,
    start_timestamp,
    end_timestamp,
    channel,
  } = inputs;

  const [stat, setStat] = useState({
    quota: 0,
    token: 0,
  });
  const [isStatRefreshing, setIsStatRefreshing] = useState(false);
  const [userOptions, setUserOptions] = useState([]);
  const [userSearchLoading, setUserSearchLoading] = useState(false);
  const [sortBy, setSortBy] = useState('');
  const [sortOrder, setSortOrder] = useState('desc');
  const [sortLoading, setSortLoading] = useState(false);

  const LOG_OPTIONS = [
    { key: '0', text: t('log.type.all', 'All'), value: 0 },
    { key: '1', text: t('log.type.topup', 'Recharge'), value: 1 },
    { key: '2', text: t('log.type.usage', 'Consumed'), value: 2 },
    { key: '3', text: t('log.type.admin', 'Management'), value: 3 },
    { key: '4', text: t('log.type.system', 'System'), value: 4 },
    { key: '5', text: t('log.type.test', 'Test'), value: 5 },
  ];

  const handleInputChange = (_, { name, value }) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };

  const searchUsers = async (searchQuery) => {
    if (!searchQuery.trim()) {
      setUserOptions([]);
      return;
    }

    setUserSearchLoading(true);
    try {
      const res = await API.get(`/api/user/search?keyword=${searchQuery}`);
      const { success, data } = res.data;
      if (success) {
        const options = data.map(user => ({
          key: user.id,
          value: user.username,
          text: `${user.display_name || user.username} (@${user.username})`,
          content: (
            <div style={{ display: 'flex', flexDirection: 'column' }}>
              <div style={{ fontWeight: 'bold' }}>
                {user.display_name || user.username}
              </div>
              <div style={{ fontSize: '0.9em', color: '#666' }}>
                @{user.username} • ID: {user.id}
              </div>
            </div>
          )
        }));
        setUserOptions(options);
      }
    } catch (error) {
      console.error('Failed to search users:', error);
    } finally {
      setUserSearchLoading(false);
    }
  };

  const handleStatRefresh = async () => {
    setIsStatRefreshing(true);
    try {
      if (isAdminUser) {
        await getLogStat();
      } else {
        await getLogSelfStat();
      }
    } finally {
      setIsStatRefreshing(false);
    }
  };

  const getLogSelfStat = async () => {
    let localStartTimestamp = Date.parse(start_timestamp) / 1000;
    let localEndTimestamp = Date.parse(end_timestamp) / 1000;
    let res = await API.get(
      `/api/log/self/stat?type=${logType}&token_name=${token_name}&model_name=${model_name}&start_timestamp=${localStartTimestamp}&end_timestamp=${localEndTimestamp}`
    );
    const { success, message, data } = res.data;
    if (success) {
      setStat(data);
    } else {
      showError(message);
    }
  };

  const getLogStat = async () => {
    let localStartTimestamp = Date.parse(start_timestamp) / 1000;
    let localEndTimestamp = Date.parse(end_timestamp) / 1000;
    let res = await API.get(
      `/api/log/stat?type=${logType}&username=${username}&token_name=${token_name}&model_name=${model_name}&start_timestamp=${localStartTimestamp}&end_timestamp=${localEndTimestamp}&channel=${channel}`
    );
    const { success, message, data } = res.data;
    if (success) {
      setStat(data);
    } else {
      showError(message);
    }
  };

  const handleEyeClick = async () => {
    if (!showStat) {
      if (isAdminUser) {
        await getLogStat();
      } else {
        await getLogSelfStat();
      }
    }
    setShowStat(!showStat);
  };

  const showUserTokenQuota = () => {
    return logType !== 5;
  };

  const loadLogs = async (page = 0) => {
    setLoading(true);
    let url = '';
    let localStartTimestamp = Date.parse(start_timestamp) / 1000;
    let localEndTimestamp = Date.parse(end_timestamp) / 1000;
    let sortParams = '';
    if (sortBy) {
      sortParams = `&sort_by=${sortBy}&sort_order=${sortOrder}`;
    }
    if (isAdminUser) {
      url = `/api/log/?p=${page}&type=${logType}&username=${username}&token_name=${token_name}&model_name=${model_name}&start_timestamp=${localStartTimestamp}&end_timestamp=${localEndTimestamp}&channel=${channel}${sortParams}`;
    } else {
      url = `/api/log/self?p=${page}&type=${logType}&token_name=${token_name}&model_name=${model_name}&start_timestamp=${localStartTimestamp}&end_timestamp=${localEndTimestamp}${sortParams}`;
    }

    const res = await API.get(url);
    const { success, message, data, total } = res.data;
    if (success) {
      setLogs(data);
      setTotalPages(Math.ceil(total / ITEMS_PER_PAGE));
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const onPaginationChange = (e, { activePage }) => {
    setActivePage(activePage);
    loadLogs(activePage - 1);
  };

  useEffect(() => {
    refresh();
  }, [logType, sortBy, sortOrder]); // eslint-disable-line react-hooks/exhaustive-deps

  const searchLogs = async () => {
    if (searchKeyword === '') {
      // if keyword is blank, load logs instead.
      await loadLogs(0);
      setActivePage(1);
      return;
    }
    setSearching(true);
    const url = isAdminUser ? '/api/log/search' : '/api/log/self/search';
    const res = await API.get(
      `${url}?keyword=${searchKeyword}`
    );
    const { success, message, data } = res.data;
    if (success) {
      setLogs(data);
      setActivePage(1);
    } else {
      showError(message);
    }
    setSearching(false);
  };

  const handleKeywordChange = (e, { value }) => {
    setSearchKeyword(value.trim());
  };

  const sortLog = async (key) => {
    // Prevent multiple sort requests
    if (sortLoading) return;

    // Toggle sort order if clicking the same column
    let newSortOrder = 'desc';
    if (sortBy === key && sortOrder === 'desc') {
      newSortOrder = 'asc';
    }

    setSortBy(key);
    setSortOrder(newSortOrder);
    setActivePage(1);
    setSortLoading(true);

    try {
      // Reload data with new sorting
      await loadLogs(0);
    } finally {
      setSortLoading(false);
    }
  };

  const refresh = async () => {
    setLoading(true);
    setActivePage(1);
    await loadLogs(0);
  };

  const getSortIcon = (columnKey) => {
    if (sortBy !== columnKey) {
      return <Icon name="sort" style={{ opacity: 0.5 }} />;
    }
    return <Icon name={sortOrder === 'asc' ? 'sort up' : 'sort down'} />;
  };

  const handleRowClick = (logId) => {
    setSelectedLogId(logId);
    setTracingModalOpen(true);
  };

  const handleTracingModalClose = () => {
    setTracingModalOpen(false);
    setSelectedLogId(null);
  };

  const headerCells = [
    {
      content: (
        <>
          {t('log.table.time', 'Time')}{getSortIcon('created_time')}
          {sortLoading && sortBy === 'created_time' && <span> ⏳</span>}
        </>
      ),
      sortable: true,
      onClick: () => sortLog('created_time'),
    },
    ...(isAdminUser ? [{
      content: (
        <>
          {t('log.table.channel', 'Channel')}{getSortIcon('channel')}
          {sortLoading && sortBy === 'channel' && <span> ⏳</span>}
        </>
      ),
      sortable: true,
      onClick: () => sortLog('channel'),
    }] : []),
    {
      content: (
        <>
          {t('log.table.type', 'Type')}{getSortIcon('type')}
          {sortLoading && sortBy === 'type' && <span> ⏳</span>}
        </>
      ),
      sortable: true,
      onClick: () => sortLog('type'),
    },
    {
      content: (
        <>
          {t('log.table.model', 'Model')}{getSortIcon('model_name')}
          {sortLoading && sortBy === 'model_name' && <span> ⏳</span>}
        </>
      ),
      sortable: true,
      onClick: () => sortLog('model_name'),
    },
    ...(showUserTokenQuota() ? [
      ...(isAdminUser ? [{
        content: (
          <>
            {t('log.table.username', 'Username')}{getSortIcon('username')}
            {sortLoading && sortBy === 'username' && <span> ⏳</span>}
          </>
        ),
        sortable: true,
        onClick: () => sortLog('username'),
      }] : []),
      {
        content: (
          <>
            {t('log.table.token_name', 'Token Name')}{getSortIcon('token_name')}
            {sortLoading && sortBy === 'token_name' && <span> ⏳</span>}
          </>
        ),
        sortable: true,
        onClick: () => sortLog('token_name'),
      },
      {
        content: (
          <>
            {t('log.table.prompt_tokens', 'Prompt Tokens')}{getSortIcon('prompt_tokens')}
            {sortLoading && sortBy === 'prompt_tokens' && <span> ⏳</span>}
          </>
        ),
        sortable: true,
        onClick: () => sortLog('prompt_tokens'),
      },
      {
        content: (
          <>
            {t('log.table.completion_tokens', 'Completion Tokens')}{getSortIcon('completion_tokens')}
            {sortLoading && sortBy === 'completion_tokens' && <span> ⏳</span>}
          </>
        ),
        sortable: true,
        onClick: () => sortLog('completion_tokens'),
      },
      {
        content: (
          <>
            {t('log.table.quota', 'Quota')}{getSortIcon('quota')}
            {sortLoading && sortBy === 'quota' && <span> ⏳</span>}
          </>
        ),
        sortable: true,
        onClick: () => sortLog('quota'),
      },
      {
        content: (
          <>
            {t('log.table.latency', 'Latency')}{getSortIcon('elapsed_time')}
            {sortLoading && sortBy === 'elapsed_time' && <span> ⏳</span>}
          </>
        ),
        sortable: true,
        onClick: () => sortLog('elapsed_time'),
      },
    ] : []),
    {
      content: (
        <>
          {t('log.table.detail', 'Detail')}{getSortIcon('content')}
          {sortLoading && sortBy === 'content' && <span> ⏳</span>}
        </>
      ),
      sortable: true,
      onClick: () => sortLog('content'),
    },
  ];

  const footerButtons = [
    {
      content: t('log.buttons.refresh', 'Refresh'),
      onClick: refresh,
      loading: loading,
    },
    {
      content: t('log.buttons.search', 'Search'),
      onClick: searchLogs,
      loading: searching,
    },
    ...(isAdmin() ? [{
      content: t('log.buttons.clear', 'Clear'),
      onClick: async () => {
        try {
          const res = await API.delete('/api/log');
          const { success, message } = res.data;
          if (success) {
            showSuccess('Clear successful');
            await refresh();
          } else {
            showError(message);
          }
        } catch (error) {
          showError('Clear failed');
        }
      },
      loading: false,
    }] : []),
  ];

  // Calculate column span based on visible columns
  const colSpan = headerCells.length;

  return (
    <>
      <Header as='h3'>
        {t('log.usage_details', 'Usage Details')}（{t('log.total_quota', 'Total Quota')}：
        {showStat && (
          <>
            {renderQuota(stat.quota, t)}
            <Button
              size='mini'
              circular
              icon='refresh'
              onClick={handleStatRefresh}
              loading={isStatRefreshing}
              disabled={isStatRefreshing}
              style={{
                marginLeft: '8px',
                padding: '4px',
                minHeight: '20px',
                minWidth: '20px',
                fontSize: '10px'
              }}
              title={t('log.refresh_quota', 'Refresh quota data')}
            />
          </>
        )}
        {!showStat && (
          <span
            onClick={handleEyeClick}
            style={{ cursor: 'pointer', color: 'gray' }}
          >
            {t('log.click_to_view', 'Click to view')}
          </span>
        )}
        ）
      </Header>

      <Segment>
        <Form onSubmit={searchLogs}>
          <Form.Group>
            <Form.Input
              fluid
              label={t('log.table.token_name', 'Token Name')}
              size={'small'}
              width={3}
              value={token_name}
              placeholder={t('log.table.token_name_placeholder', 'Search by token name')}
              name='token_name'
              onChange={handleInputChange}
            />
            <Form.Input
              fluid
              label={t('log.table.model_name', 'Model Name')}
              size={'small'}
              width={3}
              value={model_name}
              placeholder={t('log.table.model_name_placeholder', 'Search by model name')}
              name='model_name'
              onChange={handleInputChange}
            />
            <Form.Input
              fluid
              label={t('log.table.start_time', 'Start Time')}
              size={'small'}
              width={4}
              value={start_timestamp}
              type='datetime-local'
              name='start_timestamp'
              onChange={handleInputChange}
            />
            <Form.Input
              fluid
              label={t('log.table.end_time', 'End Time')}
              size={'small'}
              width={4}
              value={end_timestamp}
              type='datetime-local'
              name='end_timestamp'
              onChange={handleInputChange}
            />
            <Form.Button
              fluid
              label={t('log.buttons.query', 'Query')}
              size={'small'}
              width={2}
              onClick={refresh}
            >
              {t('log.buttons.submit', 'Submit')}
            </Form.Button>
          </Form.Group>
          {isAdminUser && (
            <>
              <Form.Group>
                <Form.Input
                  fluid
                  label={t('log.table.channel_id', 'Channel ID')}
                  size={'small'}
                  width={3}
                  value={channel}
                  placeholder={t('log.table.channel_id_placeholder', 'Search by channel ID')}
                  name='channel'
                  onChange={handleInputChange}
                />
                <Form.Field width={3}>
                  <label>{t('log.table.username', 'Username')}</label>
                  <Dropdown
                    fluid
                    selection
                    search
                    clearable
                    allowAdditions
                    value={username}
                    placeholder={t('log.table.username_placeholder', 'Search by username')}
                    options={userOptions}
                    onSearchChange={(_, { searchQuery }) => searchUsers(searchQuery)}
                    onChange={(_, { value }) => handleInputChange(_, { name: 'username', value })}
                    loading={userSearchLoading}
                    noResultsMessage={t('log.no_users_found', 'No users found')}
                    additionLabel={t('log.use_username', 'Use username: ')}
                    onAddItem={(_, { value }) => {
                      const newOption = {
                        key: value,
                        value: value,
                        text: value
                      };
                      setUserOptions([...userOptions, newOption]);
                    }}
                  />
                </Form.Field>
              </Form.Group>
            </>
          )}
          <Form.Input
            icon='search'
            placeholder={t('log.search', 'Search logs')}
            value={searchKeyword}
            onChange={handleKeywordChange}
          />
        </Form>
      </Segment>

      <Segment>
        <Form>
          <Form.Group>
            <Form.Field width={4}>
              <label>{t('log.type.select', 'Select log type')}</label>
              <Select
                placeholder={t('log.type.select', 'Select log type')}
                options={LOG_OPTIONS}
                name='logType'
                value={logType}
                onChange={(_, { value }) => {
                  setLogType(value);
                }}
                fluid
              />
            </Form.Field>
          </Form.Group>
        </Form>
      </Segment>

      <BaseTable
        loading={loading}
        activePage={activePage}
        totalPages={totalPages}
        onPageChange={onPaginationChange}
        headerCells={headerCells}
        footerButtons={footerButtons}
        colSpan={colSpan}
        size="small"
      >
        {logs.map((log, idx) => {
          if (log.deleted) return null;
          return (
            <Table.Row
              key={log.id}
              data-label={`Log ${log.id}`}
              style={{ cursor: 'pointer' }}
              onClick={() => handleRowClick(log.id)}
            >
              <Table.Cell data-label={t('log.table.time', 'Time')}>
                {renderTimestamp(log.created_at, log.request_id)}
              </Table.Cell>
              {isAdminUser && (
                <Table.Cell data-label={t('log.table.channel', 'Channel')}>
                  {log.channel ? (
                    <Label basic>
                      {log.channel}
                    </Label>
                  ) : (
                    ''
                  )}
                </Table.Cell>
              )}
              <Table.Cell data-label={t('log.table.type', 'Type')}>
                {renderLogType(log.type, t)}
              </Table.Cell>
              <Table.Cell data-label={t('log.table.model', 'Model')}>
                {log.model_name ? (
                  <Label basic color='blue'>
                    {log.model_name}
                  </Label>
                ) : (
                  ''
                )}
              </Table.Cell>
              {showUserTokenQuota() && (
                <>
                  {isAdminUser && (
                    <Table.Cell data-label={t('log.table.username', 'Username')}>
                      {log.username ? (
                        <Label basic>
                          {log.username}
                        </Label>
                      ) : (
                        ''
                      )}
                    </Table.Cell>
                  )}
                  <Table.Cell data-label={t('log.table.token_name', 'Token Name')}>
                    {log.token_name ? (
                      <Label basic color='teal'>
                        {log.token_name}
                      </Label>
                    ) : (
                      ''
                    )}
                  </Table.Cell>
                  <Table.Cell data-label={t('log.table.prompt_tokens', 'Prompt Tokens')}>
                    {log.prompt_tokens || ''}
                  </Table.Cell>
                  <Table.Cell data-label={t('log.table.completion_tokens', 'Completion Tokens')}>
                    {log.completion_tokens || ''}
                  </Table.Cell>
                  <Table.Cell data-label={t('log.table.quota', 'Quota')}>
                    {log.quota ? renderQuota(log.quota, t, 6) : 'free'}
                  </Table.Cell>
                  <Table.Cell data-label={t('log.table.latency', 'Latency')}>
                    {log.elapsed_time ? `${log.elapsed_time} ms` : ''}
                  </Table.Cell>
                </>
              )}
              <Table.Cell data-label={t('log.table.detail', 'Detail')}>
                <ExpandableDetail
                  content={log.content}
                  isStream={log.is_stream}
                  systemPromptReset={log.system_prompt_reset}
                />
              </Table.Cell>
            </Table.Row>
          );
        })}
      </BaseTable>

      <TracingModal
        open={tracingModalOpen}
        onClose={handleTracingModalClose}
        logId={selectedLogId}
      />
    </>
  );
};

export default LogsTable;
