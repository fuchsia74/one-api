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
} from 'semantic-ui-react';
import { Link } from 'react-router-dom';
import {
  API,
  showError,
  showSuccess,
  renderQuota,
} from '../helpers';

import { ITEMS_PER_PAGE } from '../constants';
import { cleanDisplay } from './shared/tableUtils';

function renderUserRole(role, t) {
  switch (role) {
    case 1:
      return (
        <Label basic color='blue'>
          Normal
        </Label>
      );
    case 10:
      return (
        <Label basic color='yellow'>
          Admin
        </Label>
      );
    case 100:
      return (
        <Label basic color='orange'>
          Super Admin
        </Label>
      );
    default:
      return (
        <Label basic color='red'>
          Unknown
        </Label>
      );
  }
}

function renderUserStatus(status, t) {
  switch (status) {
    case 1:
      return (
        <Label basic color='green'>
          Enabled
        </Label>
      );
    case 2:
      return (
        <Label basic color='red'>
          Disabled
        </Label>
      );
    default:
      return (
        <Label basic color='black'>
          Unknown
        </Label>
      );
  }
}

const UsersTableCompact = () => {
  const { t } = useTranslation();
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [searching, setSearching] = useState(false);
  const [sortBy, setSortBy] = useState('');
  const [sortOrder, setSortOrder] = useState('desc');

  const SORT_OPTIONS = [
    { key: '', text: t('users.sort.default', 'Default'), value: '' },
    { key: 'quota', text: t('users.sort.remaining_quota', 'Remaining Quota'), value: 'quota' },
    { key: 'used_quota', text: t('users.sort.used_quota', 'Used Quota'), value: 'used_quota' },
    { key: 'username', text: t('users.sort.username', 'Username'), value: 'username' },
    { key: 'id', text: t('users.sort.id', 'ID'), value: 'id' },
    { key: 'created_time', text: t('users.sort.created_time', 'Created Time'), value: 'created_time' },
  ];

  const loadUsers = async (page = 0, sortBy = '', sortOrder = 'desc') => {
    setLoading(true);
    let url = `/api/user/?p=${page}`;
    if (sortBy) {
      url += `&sort=${sortBy}&order=${sortOrder}`;
    }
    const res = await API.get(url);
    const { success, message, data, total } = res.data;
    if (success) {
      setUsers(data);
      setTotalPages(Math.ceil(total / ITEMS_PER_PAGE));
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const onPaginationChange = (e, { activePage }) => {
    setActivePage(activePage);
    loadUsers(activePage - 1, sortBy, sortOrder);
  };

  useEffect(() => {
    loadUsers(0, sortBy, sortOrder)
      .then()
      .catch((reason) => {
        showError(reason);
      });
  }, [sortBy, sortOrder]); // eslint-disable-line react-hooks/exhaustive-deps

  const manageUser = async (id, action, idx) => {
    let data = { id };
    let res;
    switch (action) {
      case 'delete':
        res = await API.delete(`/api/user/${id}`);
        break;
      case 'enable':
        data.status = 1;
        res = await API.put('/api/user/?status_only=true', data);
        break;
      case 'disable':
        data.status = 2;
        res = await API.put('/api/user/?status_only=true', data);
        break;
      default:
        return;
    }
    const { success, message } = res.data;
    if (success) {
      showSuccess('Operation successful');
      let user = res.data.data;
      let newUsers = [...users];
      if (action === 'delete') {
        newUsers[idx].deleted = true;
      } else {
        newUsers[idx].status = user.status;
      }
      setUsers(newUsers);
    } else {
      showError(message);
    }
  };

  const searchUsers = async () => {
    if (searchKeyword === '') {
      // if keyword is blank, load users instead.
      await loadUsers(0, sortBy, sortOrder);
      setActivePage(1);
      return;
    }
    setSearching(true);
    let url = `/api/user/search?keyword=${searchKeyword}`;
    if (sortBy) {
      url += `&sort=${sortBy}&order=${sortOrder}`;
    }
    const res = await API.get(url);
    const { success, message, data } = res.data;
    if (success) {
      setUsers(data);
      setActivePage(1);
    } else {
      showError(message);
    }
    setSearching(false);
  };

  const handleKeywordChange = async (e, { value }) => {
    setSearchKeyword(value.trim());
  };

  const sortUser = async (key) => {
    const newSortOrder = sortBy === key && sortOrder === 'desc' ? 'asc' : 'desc';
    setSortBy(key);
    setSortOrder(newSortOrder);
    await loadUsers(activePage - 1, key, newSortOrder);
  };

  const handleSortChange = async (e, { value }) => {
    setSortBy(value);
    setSortOrder('desc');
    setActivePage(1);
    await loadUsers(0, value, 'desc');
  };

  const getSortIcon = (columnKey) => {
    if (sortBy !== columnKey) {
      return <Icon name="sort" style={{ opacity: 0.5 }} />;
    }
    return <Icon name={sortOrder === 'asc' ? 'sort up' : 'sort down'} />;
  };

  const refresh = async () => {
    setLoading(true);
    await loadUsers(0, sortBy, sortOrder);
    setActivePage(1);
  };

  return (
    <>
      <Form onSubmit={searchUsers}>
        <Form.Group>
          <Form.Input
            width={12}
            icon='search'
            iconPosition='left'
            placeholder={t('users.search.placeholder', 'Search by username...')}
            value={searchKeyword}
            loading={searching}
            onChange={handleKeywordChange}
          />
          <Form.Dropdown
            width={4}
            selection
            placeholder={t('users.sort.placeholder', 'Sort by...')}
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
                sortUser('id');
              }}
              style={{ cursor: 'pointer' }}
            >
              ID {getSortIcon('id')}
            </Table.HeaderCell>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortUser('username');
              }}
              style={{ cursor: 'pointer' }}
            >
              {t('users.table.username', 'Username')} {getSortIcon('username')}
            </Table.HeaderCell>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortUser('role');
              }}
              style={{ cursor: 'pointer' }}
            >
              {t('users.table.role', 'Role')} {getSortIcon('role')}
            </Table.HeaderCell>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortUser('status');
              }}
              style={{ cursor: 'pointer' }}
            >
              {t('users.table.status', 'Status')} {getSortIcon('status')}
            </Table.HeaderCell>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortUser('quota');
              }}
              style={{ cursor: 'pointer' }}
            >
              {t('users.table.remaining_quota', 'Remaining Quota')} {getSortIcon('quota')}
            </Table.HeaderCell>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortUser('used_quota');
              }}
              style={{ cursor: 'pointer' }}
            >
              {t('users.table.used_quota', 'Used Quota')} {getSortIcon('used_quota')}
            </Table.HeaderCell>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortUser('group');
              }}
              style={{ cursor: 'pointer' }}
            >
              {t('users.table.group', 'Group')} {getSortIcon('group')}
            </Table.HeaderCell>
            <Table.HeaderCell>{t('users.table.actions', 'Actions')}</Table.HeaderCell>
          </Table.Row>
        </Table.Header>

        <Table.Body>
          {users.map((user, idx) => {
              if (user.deleted) return <></>;
              return (
                <Table.Row key={user.id}>
                  <Table.Cell>{user.id}</Table.Cell>
                  <Table.Cell>
                    {cleanDisplay(user.username)}
                  </Table.Cell>
                  <Table.Cell>{renderUserRole(user.role, t)}</Table.Cell>
                  <Table.Cell>{renderUserStatus(user.status, t)}</Table.Cell>
                  <Table.Cell>
                    {user.quota === -1 ? (
                      <Label basic color="green">
                        {t('unlimited')}
                      </Label>
                    ) : (
                      renderQuota(user.quota, t)
                    )}
                  </Table.Cell>
                  <Table.Cell>
                    {user.used_quota ? renderQuota(user.used_quota, t) : '0'}
                  </Table.Cell>
                  <Table.Cell>
                    {cleanDisplay(user.group, 'default')}
                  </Table.Cell>
                  <Table.Cell>
                    <div>
                      <Button
                        size={'tiny'}
                        onClick={() => {
                          manageUser(
                            user.id,
                            user.status === 1 ? 'disable' : 'enable',
                            idx
                          );
                        }}
                      >
                        {user.status === 1
                          ? t('disable')
                          : t('enable')}
                      </Button>
                      <Button
                        size={'tiny'}
                        as={Link}
                        to={'/user/edit/' + user.id}
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
                            manageUser(user.id, 'delete', idx);
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
            <Table.HeaderCell colSpan='8'>
              <Button
                size='small'
                as={Link}
                to='/user/add'
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

export default UsersTableCompact;
