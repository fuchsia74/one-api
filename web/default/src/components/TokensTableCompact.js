import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Button,
  Form,
  Label,
  Popup,
  Table,
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
import FixedPagination from './FixedPagination';

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
  const [searchKeyword, setSearchKeyword] = useState('');
  const [searching, setSearching] = useState(false);

  const loadTokens = async (startIdx) => {
    const res = await API.get(`/api/token/?p=${startIdx}`);
    const { success, message, data } = res.data;
    if (success) {
      if (startIdx === 0) {
        setTokens(data);
      } else {
        let newTokens = tokens;
        newTokens.push(...data);
        setTokens(newTokens);
      }
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const onPaginationChange = (e, { activePage }) => {
    (async () => {
      if (activePage === Math.ceil(tokens.length / ITEMS_PER_PAGE) + 1) {
        // In this case we have to load more data and then append them.
        await loadTokens(activePage - 1);
      }
      setActivePage(activePage);
    })();
  };

  useEffect(() => {
    loadTokens(0)
      .then()
      .catch((reason) => {
        showError(reason);
      });
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

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
      let realIdx = (activePage - 1) * ITEMS_PER_PAGE + idx;
      if (action === 'delete') {
        newTokens[realIdx].deleted = true;
      } else {
        newTokens[realIdx].status = token.status;
      }
      setTokens(newTokens);
    } else {
      showError(message);
    }
  };

  const searchTokens = async () => {
    if (searchKeyword === '') {
      // if keyword is blank, load tokens instead.
      await loadTokens(0);
      setActivePage(1);
      return;
    }
    setSearching(true);
    const res = await API.get(
      `/api/token/search?keyword=${searchKeyword}`
    );
    const { success, message, data } = res.data;
    if (success) {
      setTokens(data);
      setActivePage(1);
    } else {
      showError(message);
    }
    setSearching(false);
  };

  const handleKeywordChange = async (e, { value }) => {
    setSearchKeyword(value.trim());
  };

  const sortToken = (key) => {
    if (tokens.length === 0) return;
    setLoading(true);
    let sortedTokens = [...tokens];
    sortedTokens.sort((a, b) => {
      if (!isNaN(a[key])) {
        // If the value is numeric, subtract to sort
        return a[key] - b[key];
      } else {
        // If the value is not numeric, sort as strings
        return ('' + a[key]).localeCompare(b[key]);
      }
    });
    if (sortedTokens[0].id === tokens[0].id) {
      sortedTokens.reverse();
    }
    setTokens(sortedTokens);
    setLoading(false);
  };

  const refresh = async () => {
    setLoading(true);
    await loadTokens(0);
    setActivePage(1);
  };

  return (
    <>
      <Form onSubmit={searchTokens}>
        <Form.Input
          icon='search'
          fluid
          iconPosition='left'
          placeholder={t('token.search_placeholder')}
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
                sortToken('id');
              }}
            >
              {t('id')}
            </Table.HeaderCell>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortToken('name');
              }}
            >
              {t('name')}
            </Table.HeaderCell>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortToken('status');
              }}
            >
              {t('status')}
            </Table.HeaderCell>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortToken('used_quota');
              }}
            >
              {t('used_quota')}
            </Table.HeaderCell>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortToken('remain_quota');
              }}
            >
              {t('remain_quota')}
            </Table.HeaderCell>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortToken('created_time');
              }}
            >
              {t('created_time')}
            </Table.HeaderCell>
            <Table.HeaderCell>{t('actions')}</Table.HeaderCell>
          </Table.Row>
        </Table.Header>

        <Table.Body>
          {tokens
            .slice(
              (activePage - 1) * ITEMS_PER_PAGE,
              activePage * ITEMS_PER_PAGE
            )
            .map((token, idx) => {
              if (token.deleted) return <></>;
              return (
                <Table.Row key={token.id}>
                  <Table.Cell>{token.id}</Table.Cell>
                  <Table.Cell>
                    {cleanDisplay(token.name)}
                  </Table.Cell>
                  <Table.Cell>{renderTokenStatus(token.status, t)}</Table.Cell>
                  <Table.Cell>{renderQuota(token.used_quota, t)}</Table.Cell>
                  <Table.Cell>
                    {token.unlimited_quota ? (
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
                        positive
                        onClick={async () => {
                          if (await copy(token.key)) {
                            showSuccess(t('token.messages.copy_success'));
                          } else {
                            showWarning(t('token.messages.copy_failed'));
                            setSearchKeyword(token.key);
                          }
                        }}
                      >
                        {t('token.buttons.copy')}
                      </Button>
                      <Popup
                        trigger={
                          <Button size='tiny' negative>
                            {t('token.buttons.delete')}
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
                          {t('token.buttons.confirm_delete')}
                        </Button>
                      </Popup>
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
                          ? t('token.buttons.disable')
                          : t('token.buttons.enable')}
                      </Button>
                      <Button
                        size={'tiny'}
                        as={Link}
                        to={'/token/edit/' + token.id}
                      >
                        {t('token.buttons.edit')}
                      </Button>
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
                {t('token.buttons.add')}
              </Button>
              <Button size='small' onClick={refresh} loading={loading}>
                {t('token.buttons.refresh')}
              </Button>
            </Table.HeaderCell>
          </Table.Row>
        </Table.Footer>
      </Table>
      {(() => {
        // Calculate total pages based on loaded data, but always allow +1 for potential more data
        const currentPages = Math.ceil(tokens.length / ITEMS_PER_PAGE);
        const totalPages = Math.max(currentPages, activePage + (tokens.length % ITEMS_PER_PAGE === 0 ? 1 : 0));

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

export default TokensTableCompact;
