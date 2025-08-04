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
  showError,
  showSuccess,
  renderQuota,
} from '../helpers';

import { ITEMS_PER_PAGE } from '../constants';
import { cleanDisplay } from './shared/tableUtils';
import FixedPagination from './FixedPagination';

function renderUserRole(role, t) {
  switch (role) {
    case 1:
      return (
        <Label basic color='blue'>
          {t('role_types.normal')}
        </Label>
      );
    case 10:
      return (
        <Label basic color='yellow'>
          {t('role_types.admin')}
        </Label>
      );
    case 100:
      return (
        <Label basic color='orange'>
          {t('role_types.super_admin')}
        </Label>
      );
    default:
      return (
        <Label basic color='red'>
          {t('role_types.unknown')}
        </Label>
      );
  }
}

function renderUserStatus(status, t) {
  switch (status) {
    case 1:
      return (
        <Label basic color='green'>
          {t('status.enabled')}
        </Label>
      );
    case 2:
      return (
        <Label basic color='red'>
          {t('status.disabled')}
        </Label>
      );
    default:
      return (
        <Label basic color='black'>
          {t('status.unknown')}
        </Label>
      );
  }
}

const UsersTableCompact = () => {
  const { t } = useTranslation();
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [searching, setSearching] = useState(false);

  const loadUsers = async (startIdx) => {
    const res = await API.get(`/api/user/?p=${startIdx}`);
    const { success, message, data } = res.data;
    if (success) {
      if (startIdx === 0) {
        setUsers(data);
      } else {
        let newUsers = users;
        newUsers.push(...data);
        setUsers(newUsers);
      }
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const onPaginationChange = (e, { activePage }) => {
    (async () => {
      if (activePage === Math.ceil(users.length / ITEMS_PER_PAGE) + 1) {
        // In this case we have to load more data and then append them.
        await loadUsers(activePage - 1);
      }
      setActivePage(activePage);
    })();
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
      showSuccess(t('user.messages.operation_success'));
      let user = res.data.data;
      let newUsers = [...users];
      let realIdx = (activePage - 1) * ITEMS_PER_PAGE + idx;
      if (action === 'delete') {
        newUsers[realIdx].deleted = true;
      } else {
        newUsers[realIdx].status = user.status;
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
          placeholder={t('user.search_placeholder')}
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
          {users
            .slice(
              (activePage - 1) * ITEMS_PER_PAGE,
              activePage * ITEMS_PER_PAGE
            )
            .map((user, idx) => {
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
            </Table.HeaderCell>
          </Table.Row>
        </Table.Footer>
      </Table>
      {(() => {
        // Calculate total pages based on loaded data, but always allow +1 for potential more data
        const currentPages = Math.ceil(users.length / ITEMS_PER_PAGE);
        const totalPages = Math.max(currentPages, activePage + (users.length % ITEMS_PER_PAGE === 0 ? 1 : 0));

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

export default UsersTableCompact;
