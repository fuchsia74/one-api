import React from 'react';
import TokensTableCompact from './TokensTableCompact';

// Legacy wrapper for backwards compatibility
const TokensTable = (props) => {
  return <TokensTableCompact {...props} />;
};

export default TokensTable;
