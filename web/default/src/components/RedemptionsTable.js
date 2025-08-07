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
  copy,
  showError,
  showSuccess,
  showWarning,
  timestamp2string,
  renderQuota,
} from '../helpers';
import { ITEMS_PER_PAGE } from '../constants';
import BaseTable from './shared/BaseTable';

function renderTimestamp(timestamp) {
  return <>{timestamp2string(timestamp)}</>;
}

function renderStatus(status, t) {
  switch (status) {
    case 1:
      return (
        <Label basic color='green'>
          {t('redemption.status.unused')}
        </Label>
      );
    case 2:
      return (
        <Label basic color='red'>
          {t('redemption.status.disabled')}
        </Label>
      );
    case 3:
      return (
        <Label basic color='grey'>
          {t('redemption.status.used')}
        </Label>
      );
    default:
      return (
        <Label basic color='black'>
          {t('redemption.status.unknown')}
        </Label>
      );
  }
}

const RedemptionsTable = () => {
  const { t } = useTranslation();
  const [redemptions, setRedemptions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [searching, setSearching] = useState(false);
  const [sortBy, setSortBy] = useState('');
  const [sortOrder, setSortOrder] = useState('desc');

  const SORT_OPTIONS = [
    { key: '', text: t('redemptions.sort.default', 'Default'), value: '' },
    { key: 'id', text: t('redemptions.sort.id', 'ID'), value: 'id' },
    { key: 'name', text: t('redemptions.sort.name', 'Name'), value: 'name' },
    { key: 'status', text: t('redemptions.sort.status', 'Status'), value: 'status' },
    { key: 'quota', text: t('redemptions.sort.quota', 'Quota'), value: 'quota' },
    { key: 'created_time', text: t('redemptions.sort.created_time', 'Created Time'), value: 'created_time' },
    { key: 'redeemed_time', text: t('redemptions.sort.redeemed_time', 'Redeemed Time'), value: 'redeemed_time' },
  ];

  const loadRedemptions = async (page = 0, sortBy = '', sortOrder = 'desc') => {
    setLoading(true);
    let url = `/api/redemption/?p=${page}`;
    if (sortBy) {
      url += `&sort=${sortBy}&order=${sortOrder}`;
    }
    const res = await API.get(url);
    const { success, message, data, total } = res.data;
    if (success) {
      setRedemptions(data);
      setTotalPages(Math.ceil(total / ITEMS_PER_PAGE));
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const onPaginationChange = (e, { activePage }) => {
    setActivePage(activePage);
    loadRedemptions(activePage - 1, sortBy, sortOrder);
  };

  useEffect(() => {
    loadRedemptions(0, sortBy, sortOrder)
      .then()
      .catch((reason) => {
        showError(reason);
      });
  }, [sortBy, sortOrder]);

  const manageRedemption = async (id, action, idx) => {
    let data = { id };
    let res;
    switch (action) {
      case 'delete':
        res = await API.delete(`/api/redemption/${id}`);
        break;
      case 'enable':
        data.status = 1;
        res = await API.put('/api/redemption/?status_only=true', data);
        break;
      case 'disable':
        data.status = 2;
        res = await API.put('/api/redemption/?status_only=true', data);
        break;
      default:
        return;
    }
    const { success, message } = res.data;
    if (success) {
      showSuccess(t('token.messages.operation_success'));
      let redemption = res.data.data;
      let newRedemptions = [...redemptions];
      let realIdx = (activePage - 1) * ITEMS_PER_PAGE + idx;
      if (action === 'delete') {
        newRedemptions[realIdx].deleted = true;
      } else {
        newRedemptions[realIdx].status = redemption.status;
      }
      setRedemptions(newRedemptions);
    } else {
      showError(message);
    }
  };

  const searchRedemptions = async () => {
    if (searchKeyword === '') {
      await loadRedemptions(0, sortBy, sortOrder);
      setActivePage(1);
      return;
    }
    setSearching(true);
    let url = `/api/redemption/search?keyword=${searchKeyword}`;
    if (sortBy) {
      url += `&sort=${sortBy}&order=${sortOrder}`;
    }
    const res = await API.get(url);
    const { success, message, data } = res.data;
    if (success) {
      setRedemptions(data);
      setActivePage(1);
    } else {
      showError(message);
    }
    setSearching(false);
  };

  const handleKeywordChange = async (e, { value }) => {
    setSearchKeyword(value.trim());
  };

  const sortRedemption = async (key) => {
    const newSortOrder = sortBy === key && sortOrder === 'desc' ? 'asc' : 'desc';
    setSortBy(key);
    setSortOrder(newSortOrder);
    await loadRedemptions(activePage - 1, key, newSortOrder);
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
    await loadRedemptions(0, value, 'desc');
  };

  const refresh = async () => {
    setLoading(true);
    await loadRedemptions(0, sortBy, sortOrder);
    setActivePage(1);
  };

  const headerCells = [
    {
      content: (
        <>
          {t('redemption.table.id')} {getSortIcon('id')}
        </>
      ),
      sortable: true,
      onClick: () => sortRedemption('id'),
    },
    {
      content: (
        <>
          {t('redemption.table.name')} {getSortIcon('name')}
        </>
      ),
      sortable: true,
      onClick: () => sortRedemption('name'),
    },
    {
      content: (
        <>
          {t('redemption.table.status')} {getSortIcon('status')}
        </>
      ),
      sortable: true,
      onClick: () => sortRedemption('status'),
    },
    {
      content: (
        <>
          {t('redemption.table.quota')} {getSortIcon('quota')}
        </>
      ),
      sortable: true,
      onClick: () => sortRedemption('quota'),
    },
    {
      content: (
        <>
          {t('redemption.table.created_time')} {getSortIcon('created_time')}
        </>
      ),
      sortable: true,
      onClick: () => sortRedemption('created_time'),
    },
    {
      content: (
        <>
          {t('redemption.table.redeemed_time')} {getSortIcon('redeemed_time')}
        </>
      ),
      sortable: true,
      onClick: () => sortRedemption('redeemed_time'),
    },
    {
      content: t('redemption.table.actions'),
      sortable: false,
    },
  ];

  const footerButtons = [
    {
      content: t('redemption.buttons.add'),
      as: Link,
      to: '/redemption/add',
      loading: loading,
    },
    {
      content: t('redemption.buttons.refresh'),
      onClick: refresh,
      loading: loading,
    },
  ];

  return (
    <>
      <Form onSubmit={searchRedemptions}>
        <Form.Group>
          <Form.Input
            width={12}
            icon='search'
            iconPosition='left'
            placeholder={t('redemptions.search.placeholder', 'Search redemptions...')}
            value={searchKeyword}
            loading={searching}
            onChange={handleKeywordChange}
          />
          <Form.Dropdown
            width={4}
            selection
            placeholder={t('redemptions.sort.placeholder', 'Sort by...')}
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
        {redemptions.map((redemption, idx) => {
          if (redemption.deleted) return null;
          return (
            <Table.Row key={redemption.id}>
              <Table.Cell data-label="ID">{redemption.id}</Table.Cell>
              <Table.Cell data-label="Name">
                {redemption.name ? redemption.name : t('redemption.table.no_name')}
              </Table.Cell>
              <Table.Cell data-label="Status">{renderStatus(redemption.status, t)}</Table.Cell>
              <Table.Cell data-label="Quota">{renderQuota(redemption.quota, t)}</Table.Cell>
              <Table.Cell data-label="Created Time">
                {renderTimestamp(redemption.created_time)}
              </Table.Cell>
              <Table.Cell data-label="Redeemed Time">
                {redemption.redeemed_time
                  ? renderTimestamp(redemption.redeemed_time)
                  : t('redemption.table.not_redeemed')}{' '}
              </Table.Cell>
              <Table.Cell data-label="Actions">
                <div>
                  <Button
                    size={'tiny'}
                    positive
                    onClick={async () => {
                      if (await copy(redemption.key)) {
                        showSuccess(t('token.messages.copy_success'));
                      } else {
                        showWarning(t('token.messages.copy_failed'));
                        setSearchKeyword(redemption.key);
                      }
                    }}
                  >
                    {t('redemption.buttons.copy')}
                  </Button>
                  <Popup
                    trigger={
                      <Button size='tiny' negative>
                        {t('redemption.buttons.delete')}
                      </Button>
                    }
                    on='click'
                    flowing
                    hoverable
                  >
                    <Button
                      negative
                      onClick={() => {
                        manageRedemption(redemption.id, 'delete', idx);
                      }}
                    >
                      {t('redemption.buttons.confirm_delete')}
                    </Button>
                  </Popup>
                  <Button
                    size={'tiny'}
                    disabled={redemption.status === 3}
                    onClick={() => {
                      manageRedemption(
                        redemption.id,
                        redemption.status === 1 ? 'disable' : 'enable',
                        idx
                      );
                    }}
                  >
                    {redemption.status === 1
                      ? t('redemption.buttons.disable')
                      : t('redemption.buttons.enable')}
                  </Button>
                  <Button
                    size={'tiny'}
                    as={Link}
                    to={'/redemption/edit/' + redemption.id}
                  >
                    {t('redemption.buttons.edit')}
                  </Button>
                </div>
              </Table.Cell>
            </Table.Row>
          );
        })}
      </BaseTable>
    </>
  );
};

export default RedemptionsTable;
