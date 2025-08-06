import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Button,
  Form,
  Icon,
  Label,
  Pagination,
  Popup,
  Table,
  Dropdown,
} from 'semantic-ui-react';
import { Link } from 'react-router-dom';
import {
  API,
  copy,
  showError,
  showSuccess,
  showWarning,
  timestamp2string,
  renderQuota,
} from '../helpers';

import { ITEMS_PER_PAGE } from '../constants';
import { cleanDisplay } from './shared/tableUtils';

function renderTimestamp(timestamp) {
  return <>{timestamp2string(timestamp)}</>;
}

function renderTokenStatus(status, t) {
  switch (status) {
    case 1:
      return (
        <Label basic color='green'>
          {t('status_enabled')}
        </Label>
      );
    case 2:
      return (
        <Label basic color='red'>
          {t('status_disabled')}
        </Label>
      );
    case 3:
      return (
        <Label basic color='grey'>
          {t('status_expired')}
        </Label>
      );
    case 4:
      return (
        <Label basic color='orange'>
          {t('status_depleted')}
        </Label>
      );
    default:
      return (
        <Label basic color='black'>
          {t('status_unknown')}
        </Label>
      );
  }
}

const TokensTableCompact = () => {
  const { t } = useTranslation();
  const [tokens, setTokens] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [sortBy, setSortBy] = useState('');
  const [sortOrder, setSortOrder] = useState('desc');
  const [tokenOptions, setTokenOptions] = useState([]);
  const [tokenSearchLoading, setTokenSearchLoading] = useState(false);

  const SORT_OPTIONS = [
    { key: '', text: t('tokens.sort.default', 'Default'), value: '' },
    { key: 'id', text: t('tokens.sort.id', 'ID'), value: 'id' },
    { key: 'name', text: t('tokens.sort.name', 'Name'), value: 'name' },
    { key: 'status', text: t('tokens.sort.status', 'Status'), value: 'status' },
    { key: 'used_quota', text: t('tokens.sort.used_quota', 'Used Quota'), value: 'used_quota' },
    { key: 'remain_quota', text: t('tokens.sort.remain_quota', 'Remaining Quota'), value: 'remain_quota' },
    { key: 'created_time', text: t('tokens.sort.created_time', 'Created Time'), value: 'created_time' },
  ];

  const loadTokens = async (page = 0, sortBy = '', sortOrder = 'desc') => {
    setLoading(true);
    let url = `/api/token/?p=${page}`;
    if (sortBy) {
      url += `&sort=${sortBy}&order=${sortOrder}`;
    }
    const res = await API.get(url);
    const { success, message, data, total } = res.data;
    if (success) {
      setTokens(data);
      setTotalPages(Math.ceil(total / ITEMS_PER_PAGE));
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const onPaginationChange = (e, { activePage }) => {
    setActivePage(activePage);
    loadTokens(activePage - 1, sortBy, sortOrder);
  };

  useEffect(() => {
    loadTokens(0, sortBy, sortOrder)
      .then()
      .catch((reason) => {
        showError(reason);
      });
  }, [sortBy, sortOrder]); // eslint-disable-line react-hooks/exhaustive-deps

  const manageToken = async (id, action, idx) => {
    let data = { id };
    let res;
    switch (action) {
      case 'delete':
        res = await API.delete(`/api/token/${id}`);
        break;
      case 'enable':
        data.status = 1;
        res = await API.put('/api/token/?status_only=true', data);
        break;
      case 'disable':
        data.status = 2;
        res = await API.put('/api/token/?status_only=true', data);
        break;
      default:
        return;
    }
    const { success, message } = res.data;
    if (success) {
      showSuccess(t('token.messages.operation_success'));
      let token = res.data.data;
      let newTokens = [...tokens];
      if (action === 'delete') {
        newTokens[idx].deleted = true;
      } else {
        newTokens[idx].status = token.status;
      }
      setTokens(newTokens);
    } else {
      showError(message);
    }
  };

  const searchTokensByName = async (searchQuery) => {
    if (!searchQuery.trim()) {
      setTokenOptions([]);
      return;
    }

    setTokenSearchLoading(true);
    try {
      const res = await API.get(`/api/token/search?keyword=${searchQuery}`);
      const { success, data } = res.data;
      if (success) {
        const options = data.map(token => ({
          key: token.id,
          value: token.name,
          text: `${token.name}`,
          content: (
            <div style={{ display: 'flex', flexDirection: 'column' }}>
              <div style={{ fontWeight: 'bold' }}>
                {token.name}
              </div>
              <div style={{ fontSize: '0.9em', color: '#666' }}>
                ID: {token.id} • Status: {token.status === 1 ? 'Enabled' : 'Disabled'}
              </div>
            </div>
          )
        }));
        setTokenOptions(options);
      }
    } catch (error) {
      console.error('Failed to search tokens:', error);
    } finally {
      setTokenSearchLoading(false);
    }
  };

  const searchTokens = async () => {
    if (searchKeyword === '') {
      // if keyword is blank, load tokens instead.
      await loadTokens(0, sortBy, sortOrder);
      setActivePage(1);
      return;
    }
    let url = `/api/token/search?keyword=${searchKeyword}`;
    if (sortBy) {
      url += `&sort=${sortBy}&order=${sortOrder}`;
    }
    const res = await API.get(url);
    const { success, message, data } = res.data;
    if (success) {
      setTokens(data);
      setActivePage(1);
    } else {
      showError(message);
    }
  };

  const sortToken = async (key) => {
    const newSortOrder = sortBy === key && sortOrder === 'desc' ? 'asc' : 'desc';
    setSortBy(key);
    setSortOrder(newSortOrder);
    await loadTokens(activePage - 1, key, newSortOrder);
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
    await loadTokens(0, value, 'desc');
  };

  const refresh = async () => {
    setLoading(true);
    await loadTokens(0, sortBy, sortOrder);
    setActivePage(1);
  };

  return (
    <>
      <Form onSubmit={searchTokens}>
        <Form.Group>
          <Form.Field width={12}>
            <Dropdown
              fluid
              selection
              search
              clearable
              allowAdditions
              placeholder={t('tokens.search.placeholder', 'Search by token name...')}
              value={searchKeyword}
              options={tokenOptions}
              onSearchChange={(_, { searchQuery }) => searchTokensByName(searchQuery)}
              onChange={(_, { value }) => setSearchKeyword(value)}
              loading={tokenSearchLoading}
              noResultsMessage={t('tokens.no_tokens_found', 'No tokens found')}
              additionLabel={t('tokens.use_token_name', 'Use token name: ')}
              onAddItem={(_, { value }) => {
                const newOption = {
                  key: value,
                  value: value,
                  text: value
                };
                setTokenOptions([...tokenOptions, newOption]);
              }}
            />
          </Form.Field>
          <Form.Dropdown
            width={4}
            selection
            placeholder={t('tokens.sort.placeholder', 'Sort by...')}
            options={SORT_OPTIONS}
            value={sortBy}
            onChange={handleSortChange}
          />
        </Form.Group>
      </Form>

      <Table basic={'very'} compact size='small'>
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortToken('id');
              }}
              style={{ cursor: 'pointer' }}
            >
              {t('id')} {getSortIcon('id')}
            </Table.HeaderCell>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortToken('name');
              }}
              style={{ cursor: 'pointer' }}
            >
              {t('name')} {getSortIcon('name')}
            </Table.HeaderCell>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortToken('status');
              }}
              style={{ cursor: 'pointer' }}
            >
              {t('status')} {getSortIcon('status')}
            </Table.HeaderCell>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortToken('used_quota');
              }}
              style={{ cursor: 'pointer' }}
            >
              {t('used_quota')} {getSortIcon('used_quota')}
            </Table.HeaderCell>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortToken('remain_quota');
              }}
              style={{ cursor: 'pointer' }}
            >
              {t('remain_quota')} {getSortIcon('remain_quota')}
            </Table.HeaderCell>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortToken('created_time');
              }}
              style={{ cursor: 'pointer' }}
            >
              {t('created_time')} {getSortIcon('created_time')}
            </Table.HeaderCell>
            <Table.HeaderCell>{t('actions')}</Table.HeaderCell>
          </Table.Row>
        </Table.Header>

        <Table.Body>
          {tokens.map((token, idx) => {
            if (token.deleted) return <></>;
            return (
              <Table.Row key={token.id}>
                <Table.Cell>{token.id}</Table.Cell>
                <Table.Cell>
                  {cleanDisplay(token.name)}
                </Table.Cell>
                <Table.Cell>{renderTokenStatus(token.status, t)}</Table.Cell>
                <Table.Cell>
                  {token.used_quota ? renderQuota(token.used_quota, t) : '0'}
                </Table.Cell>
                <Table.Cell>
                  {token.remain_quota === -1 ? (
                    <Label basic color="green">
                      {t('unlimited')}
                    </Label>
                  ) : (
                    renderQuota(token.remain_quota, t)
                  )}
                </Table.Cell>
                <Table.Cell>
                  {renderTimestamp(token.created_time)}
                </Table.Cell>
                <Table.Cell>
                  <div>
                    <Button
                      size={'tiny'}
                      onClick={() => {
                        manageToken(
                          token.id,
                          token.status === 1 ? 'disable' : 'enable',
                          idx
                        );
                      }}
                    >
                      {token.status === 1
                        ? t('disable')
                        : t('enable')}
                    </Button>
                    <Button
                      size={'tiny'}
                      as={Link}
                      to={'/token/edit/' + token.id}
                    >
                      {t('edit')}
                    </Button>
                    <Popup
                      trigger={
                        <Button size='tiny' negative>
                          {t('delete')}
                        </Button>
                      }
                      on='click'
                      flowing
                      hoverable
                    >
                      <Button
                        negative
                        onClick={() => {
                          manageToken(token.id, 'delete', idx);
                        }}
                      >
                        {t('confirm_delete')}
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
                to='/token/add'
                loading={loading}
              >
                {t('add')}
              </Button>
              <Button size='small' onClick={refresh} loading={loading}>
                {t('refresh')}
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

export default TokensTableCompact;
