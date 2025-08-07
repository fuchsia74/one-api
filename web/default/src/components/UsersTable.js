import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Button,
  Form,
  Icon,
  Label,
  Popup,
  Table,
  Dropdown,
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
import BaseTable from './shared/BaseTable';

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

const UsersTable = () => {
  const { t } = useTranslation();
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [sortBy, setSortBy] = useState('');
  const [sortOrder, setSortOrder] = useState('desc');
  const [userOptions, setUserOptions] = useState([]);
  const [userSearchLoading, setUserSearchLoading] = useState(false);

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

  const searchUsersByKeyword = async () => {
    if (searchKeyword === '') {
      // if keyword is blank, load users instead.
      await loadUsers(0, sortBy, sortOrder);
      setActivePage(1);
      return;
    }
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

  const headerCells = [
    {
      content: (
        <>
          ID {getSortIcon('id')}
        </>
      ),
      sortable: true,
      onClick: () => sortUser('id'),
    },
    {
      content: (
        <>
          {t('users.table.username', 'Username')} {getSortIcon('username')}
        </>
      ),
      sortable: true,
      onClick: () => sortUser('username'),
    },
    {
      content: (
        <>
          {t('users.table.role', 'Role')} {getSortIcon('role')}
        </>
      ),
      sortable: true,
      onClick: () => sortUser('role'),
    },
    {
      content: (
        <>
          {t('users.table.status', 'Status')} {getSortIcon('status')}
        </>
      ),
      sortable: true,
      onClick: () => sortUser('status'),
    },
    {
      content: (
        <>
          {t('users.table.remaining_quota', 'Remaining Quota')} {getSortIcon('quota')}
        </>
      ),
      sortable: true,
      onClick: () => sortUser('quota'),
    },
    {
      content: (
        <>
          {t('users.table.used_quota', 'Used Quota')} {getSortIcon('used_quota')}
        </>
      ),
      sortable: true,
      onClick: () => sortUser('used_quota'),
    },
    {
      content: (
        <>
          {t('users.table.group', 'Group')} {getSortIcon('group')}
        </>
      ),
      sortable: true,
      onClick: () => sortUser('group'),
    },
    {
      content: t('users.table.actions', 'Actions'),
      sortable: false,
    },
  ];

  const footerButtons = [
    {
      content: t('add'),
      as: Link,
      to: '/user/add',
      loading: loading,
    },
    {
      content: t('refresh'),
      onClick: refresh,
      loading: loading,
    },
  ];

  return (
    <>
      <Form onSubmit={searchUsersByKeyword}>
        <Form.Group>
          <Form.Field width={12}>
            <Dropdown
              fluid
              selection
              search
              clearable
              allowAdditions
              placeholder={t('users.search.placeholder', 'Search by username...')}
              value={searchKeyword}
              options={userOptions}
              onSearchChange={(_, { searchQuery }) => searchUsers(searchQuery)}
              onChange={(_, { value }) => setSearchKeyword(value)}
              loading={userSearchLoading}
              noResultsMessage={t('users.no_users_found', 'No users found')}
              additionLabel={t('users.use_username', 'Use username: ')}
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

      <BaseTable
        loading={loading}
        activePage={activePage}
        totalPages={totalPages}
        onPageChange={onPaginationChange}
        headerCells={headerCells}
        footerButtons={footerButtons}
        colSpan={8}
      >
        {users.map((user, idx) => {
          if (user.deleted) return null;
          return (
            <Table.Row key={user.id}>
              <Table.Cell data-label="ID">{user.id}</Table.Cell>
              <Table.Cell data-label="Username">
                {cleanDisplay(user.username)}
              </Table.Cell>
              <Table.Cell data-label="Role">{renderUserRole(user.role, t)}</Table.Cell>
              <Table.Cell data-label="Status">{renderUserStatus(user.status, t)}</Table.Cell>
              <Table.Cell data-label="Remaining Quota">
                {user.quota === -1 ? (
                  <Label basic color="green">
                    {t('unlimited')}
                  </Label>
                ) : (
                  renderQuota(user.quota, t)
                )}
              </Table.Cell>
              <Table.Cell data-label="Used Quota">
                {user.used_quota ? renderQuota(user.used_quota, t) : '0'}
              </Table.Cell>
              <Table.Cell data-label="Group">
                {cleanDisplay(user.group, 'default')}
              </Table.Cell>
              <Table.Cell data-label="Actions">
                <div>
                  <Button
                    size={'tiny'}
                    color="blue"
                    as={Link}
                    to={'/user/edit/' + user.id}
                  >
                    {t('edit')}
                  </Button>
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
      </BaseTable>
    </>
  );
};

export default UsersTable;
