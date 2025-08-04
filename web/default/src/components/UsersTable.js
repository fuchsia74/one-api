import React from 'react';
import UsersTableCompact from './UsersTableCompact';

// Legacy wrapper for backwards compatibility
const UsersTable = (props) => {
  return <UsersTableCompact {...props} />;
};

export default UsersTable;
