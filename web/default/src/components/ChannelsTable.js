import React from 'react';
import ChannelsTableCompact from './ChannelsTableCompact';

// Legacy wrapper for backwards compatibility
const ChannelsTable = (props) => {
  return <ChannelsTableCompact {...props} />;
};

export default ChannelsTable;
