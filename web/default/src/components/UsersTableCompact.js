import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Button,
  Form,
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

  const loadUsers = async (page = 0) => {
    setLoading(true);
    const res = await API.get(`/api/user/?p=${page}`);
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
    loadUsers(activePage - 1);
  };

  useEffect(() => {
    loadUsers(0)
      .then()
      .catch((reason) => {
        showError(reason);
      });
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

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
      await loadUsers(0);
      setActivePage(1);
      return;
    }
    setSearching(true);
    const res = await API.get(
      `/api/user/search?keyword=${searchKeyword}`
    );
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

  const sortUser = (key) => {
    if (users.length === 0) return;
    setLoading(true);
    let sortedUsers = [...users];
    sortedUsers.sort((a, b) => {
      if (!isNaN(a[key])) {
        // If the value is numeric, subtract to sort
        return a[key] - b[key];
      } else {
        // If the value is not numeric, sort as strings
        return ('' + a[key]).localeCompare(b[key]);
      }
    });
    if (sortedUsers[0].id === users[0].id) {
      sortedUsers.reverse();
    }
    setUsers(sortedUsers);
    setLoading(false);
  };

  const refresh = async () => {
    setLoading(true);
    await loadUsers(0);
    setActivePage(1);
  };

  return (
    <>
      <Form onSubmit={searchUsers}>
        <Form.Input
          icon='search'
          fluid
          iconPosition='left'
          placeholder='Search by username...'
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
                sortUser('id');
              }}
            >
              ID
            </Table.HeaderCell>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortUser('username');
              }}
            >
              Username
            </Table.HeaderCell>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortUser('role');
              }}
            >
              Role
            </Table.HeaderCell>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortUser('status');
              }}
            >
              Status
            </Table.HeaderCell>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortUser('quota');
              }}
            >
              Quota
            </Table.HeaderCell>
            <Table.HeaderCell
              className='sortable-header'
              onClick={() => {
                sortUser('group');
              }}
            >
              Group
            </Table.HeaderCell>
            <Table.HeaderCell>Actions</Table.HeaderCell>
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
            <Table.HeaderCell colSpan='7'>
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
