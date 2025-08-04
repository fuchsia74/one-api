import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Button,
  Form,
  Label,
  Popup,
  Table,
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

import { ITEMS_PER_PAGE } from '../constants';
import { cleanDisplay } from './shared/tableUtils';
import FixedPagination from './FixedPagination';

function renderTimestamp(timestamp, request_id) {
  const fullTimestamp = timestamp2string(timestamp);
  const compactTimestamp = fullTimestamp.length > 10 && fullTimestamp.includes('-')
    ? fullTimestamp.slice(5)
    : fullTimestamp;

  return (
    <Popup
      content={`Full time: ${fullTimestamp}${request_id ? `\nRequest ID: ${request_id}` : ''}`}
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
          {compactTimestamp}
        </code>
      }
    />
  );
}

function renderLogType(type, t) {
  const typeConfig = {
    1: { color: 'green', text: 'Consume' },
    2: { color: 'olive', text: 'Recharge' },
    3: { color: 'orange', text: 'Management' },
    4: { color: 'purple', text: 'System' },
  };

  const config = typeConfig[type] || {
    color: 'black',
    text: 'Unknown'
  };

  return (
    <Label basic color={config.color}>
      {config.text}
    </Label>
  );
}

const LogsTableCompact = () => {
  const { t } = useTranslation();
  const [logs, setLogs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [searching, setSearching] = useState(false);

  const loadLogs = async (startIdx) => {
    const res = await API.get(`/api/log/?p=${startIdx}`);
    const { success, message, data } = res.data;
    if (success) {
      if (startIdx === 0) {
        setLogs(data);
      } else {
        let newLogs = logs;
        newLogs.push(...data);
        setLogs(newLogs);
      }
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const onPaginationChange = (e, { activePage }) => {
    (async () => {
      if (activePage === Math.ceil(logs.length / ITEMS_PER_PAGE) + 1) {
        // In this case we have to load more data and then append them.
        await loadLogs(activePage - 1);
      }
      setActivePage(activePage);
    })();
  };

  useEffect(() => {
    loadLogs(0)
      .then()
      .catch((reason) => {
        showError(reason);
      });
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const searchLogs = async () => {
    if (searchKeyword === '') {
      // if keyword is blank, load logs instead.
      await loadLogs(0);
      setActivePage(1);
      return;
    }
    setSearching(true);
    const res = await API.get(
      `/api/log/search?keyword=${searchKeyword}`
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

  const handleKeywordChange = async (e, { value }) => {
    setSearchKeyword(value.trim());
  };

  const sortLog = (key) => {
    if (logs.length === 0) return;
    setLoading(true);
    let sortedLogs = [...logs];
    sortedLogs.sort((a, b) => {
      if (!isNaN(a[key])) {
        // If the value is numeric, subtract to sort
        return a[key] - b[key];
      } else {
        // If the value is not numeric, sort as strings
        return ('' + a[key]).localeCompare(b[key]);
      }
    });
    if (sortedLogs[0].id === logs[0].id) {
      sortedLogs.reverse();
    }
    setLogs(sortedLogs);
    setLoading(false);
  };

  const refresh = async () => {
    setLoading(true);
    await loadLogs(0);
    setActivePage(1);
  };

  return (
    <>
      <Form onSubmit={searchLogs}>
        <Form.Input
          icon='search'
          fluid
          iconPosition='left'
          placeholder={t('log.search_placeholder')}
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
                sortLog('created_at');
              }}
            >
              {t('time')}
            </Table.HeaderCell>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortLog('channel');
              }}
            >
              {t('channel')}
            </Table.HeaderCell>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortLog('username');
              }}
            >
              {t('user')}
            </Table.HeaderCell>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortLog('type');
              }}
            >
              {t('type')}
            </Table.HeaderCell>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortLog('model_name');
              }}
            >
              {t('model')}
            </Table.HeaderCell>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortLog('quota');
              }}
            >
              {t('quota')}
            </Table.HeaderCell>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortLog('content');
              }}
            >
              {t('details')}
            </Table.HeaderCell>
          </Table.Row>
        </Table.Header>

        <Table.Body>
          {logs
            .slice(
              (activePage - 1) * ITEMS_PER_PAGE,
              activePage * ITEMS_PER_PAGE
            )
            .map((log, idx) => {
              return (
                <Table.Row key={log.id}>
                  <Table.Cell>
                    {renderTimestamp(log.created_at, log.request_id)}
                  </Table.Cell>
                  <Table.Cell>
                    {cleanDisplay(log.channel)}
                  </Table.Cell>
                  <Table.Cell>
                    {cleanDisplay(log.username)}
                  </Table.Cell>
                  <Table.Cell>
                    {renderLogType(log.type, t)}
                  </Table.Cell>
                  <Table.Cell>
                    {cleanDisplay(log.model_name)}
                  </Table.Cell>
                  <Table.Cell>
                    {log.quota !== undefined && log.quota !== null && log.quota !== 0 ? renderQuota(log.quota, t) : ''}
                  </Table.Cell>
                  <Table.Cell>
                    <div style={{
                      maxWidth: '200px',
                      overflow: 'hidden',
                      textOverflow: 'ellipsis',
                      whiteSpace: 'nowrap'
                    }}>
                      {cleanDisplay(log.content)}
                    </div>
                  </Table.Cell>
                </Table.Row>
              );
            })}
        </Table.Body>

        <Table.Footer>
          <Table.Row>
            <Table.HeaderCell colSpan='7'>
              <Button size='small' onClick={refresh} loading={loading}>
                {t('refresh')}
              </Button>
              {isAdmin() && (
                <Button
                  size='small'
                  onClick={async () => {
                    try {
                      const res = await API.delete('/api/log');
                      const { success, message } = res.data;
                      if (success) {
                        showSuccess(t('log.messages.clear_success'));
                        await refresh();
                      } else {
                        showError(message);
                      }
                    } catch (error) {
                      showError(t('log.messages.clear_failed'));
                    }
                  }}
                >
                  {t('clear')}
                </Button>
              )}
            </Table.HeaderCell>
          </Table.Row>
        </Table.Footer>
      </Table>
      {(() => {
        // Calculate total pages based on loaded data, but always allow +1 for potential more data
        const currentPages = Math.ceil(logs.length / ITEMS_PER_PAGE);
        const totalPages = Math.max(currentPages, activePage + (logs.length % ITEMS_PER_PAGE === 0 ? 1 : 0));

        return (
          <FixedPagination
            activePage={activePage}
            onPageChange={(e, data) => {
              onPaginationChange(e, data);
            }}
            totalPages={totalPages}
          />
        );
      })()}
    </>
  );
};

export default LogsTableCompact;
