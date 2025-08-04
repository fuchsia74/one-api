import React, { useEffect } from 'react';
import { styled } from '@mui/material/styles';
import TablePagination from '@mui/material/TablePagination';
import Paper from '@mui/material/Paper';

const StyledPaginationContainer = styled(Paper)(({ theme }) => ({
  position: 'fixed !important',
  bottom: '20px !important',
  left: '50% !important',
  transform: 'translateX(-50%) !important',
  zIndex: '1000 !important',
  backgroundColor: `${theme.palette.background.paper} !important`,
  borderRadius: '8px !important',
  padding: '8px 16px !important',
  boxShadow: `${theme.shadows[8]} !important`,
  backdropFilter: 'blur(10px) !important',
  border: `1px solid ${theme.palette.divider} !important`,
  display: 'flex !important',
  alignItems: 'center !important',
  justifyContent: 'center !important',
  visibility: 'visible !important',
  opacity: '1 !important',

  [theme.breakpoints.down('md')]: {
    bottom: '10px !important',
    left: '10px !important',
    right: '10px !important',
    transform: 'none !important',
    width: 'auto !important',
    maxWidth: 'calc(100vw - 20px) !important',
    padding: '4px 8px !important',
    display: 'flex !important',
    visibility: 'visible !important',
    opacity: '1 !important',
  },

  '& .MuiTablePagination-root': {
    overflow: 'visible',
  },

  '& .MuiTablePagination-toolbar': {
    minHeight: '40px',
    paddingLeft: '8px',
    paddingRight: '8px',

    [theme.breakpoints.down('md')]: {
      minHeight: '36px',
      paddingLeft: '4px',
      paddingRight: '4px',
    },
  },

  '& .MuiTablePagination-selectLabel, & .MuiTablePagination-displayedRows': {
    fontSize: '0.875rem',

    [theme.breakpoints.down('md')]: {
      fontSize: '0.75rem',
    },
  },

  '& .MuiTablePagination-actions': {
    marginLeft: '8px',

    [theme.breakpoints.down('md')]: {
      marginLeft: '4px',
    },
  },

  '& .MuiIconButton-root': {
    padding: '8px !important',
    minWidth: '40px !important',
    height: '40px !important',
    backgroundColor: 'rgba(25, 118, 210, 0.1) !important',
    border: '1px solid rgba(25, 118, 210, 0.3) !important',
    borderRadius: '6px !important',
    color: '#1976d2 !important',
    fontWeight: 'bold !important',
    transition: 'all 0.2s ease !important',

    '&:hover': {
      backgroundColor: 'rgba(25, 118, 210, 0.2) !important',
      borderColor: 'rgba(25, 118, 210, 0.5) !important',
      transform: 'translateY(-1px) !important',
    },

    '&:disabled': {
      opacity: '0.4 !important',
      backgroundColor: 'rgba(0, 0, 0, 0.05) !important',
      borderColor: 'rgba(0, 0, 0, 0.1) !important',
      color: 'rgba(0, 0, 0, 0.3) !important',
    },

    [theme.breakpoints.down('md')]: {
      padding: '6px !important',
      minWidth: '36px !important',
      height: '36px !important',
    },
  },
}));

const FixedPagination = ({
  page,
  count,
  rowsPerPage,
  onPageChange,
  rowsPerPageOptions = [10],
  ...props
}) => {
  useEffect(() => {
    console.log(`[Berry FixedPagination] Component rendered - page: ${page}, count: ${count}, rowsPerPage: ${rowsPerPage}`);
  }, [page, count, rowsPerPage]);

  // Don't render if there's only one page or no data
  if (!count || count <= rowsPerPage) {
    console.log(`[Berry FixedPagination] Hiding pagination - count: ${count}, rowsPerPage: ${rowsPerPage}`);
    return null;
  }

  const handlePageChange = (event, newPage) => {
    event.preventDefault();
    event.stopPropagation();
    console.log(`[Berry FixedPagination] Page change - from ${page} to ${newPage}`);
    if (onPageChange) {
      onPageChange(event, newPage);
    }
  };

  console.log(`[Berry FixedPagination] Rendering pagination`);
  return (
    <StyledPaginationContainer elevation={3}>
      <TablePagination
        page={page}
        component="div"
        count={count}
        rowsPerPage={rowsPerPage}
        onPageChange={handlePageChange}
        rowsPerPageOptions={rowsPerPageOptions}
        {...props}
      />
    </StyledPaginationContainer>
  );
};

export default FixedPagination;
