import React from 'react';
import LogsTableCompact from './LogsTableCompact';

// Legacy wrapper for backwards compatibility
const LogsTable = (props) => {
  return <LogsTableCompact {...props} />;
};

export default LogsTable;
