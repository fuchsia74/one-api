import React from 'react';
import { Button, Table, Pagination } from 'semantic-ui-react';
import { useTranslation } from 'react-i18next';

const BaseTable = ({
  children,
  loading = false,
  activePage = 1,
  totalPages = 1,
  onPageChange = () => { },
  headerCells = [],
  footerButtons = [],
  colSpan = 5,
  className = '',
  size = 'small'
}) => {
  const { t } = useTranslation();

  return (
    <div className="table-container">
      <Table basic={'very'} compact size={size} className={className}>
        <Table.Header>
          <Table.Row>
            {headerCells.map((header, index) => (
              <Table.HeaderCell
                key={index}
                className={header.sortable ? 'sortable-header' : ''}
                onClick={header.onClick}
                style={header.sortable ? { cursor: 'pointer' } : {}}
              >
                {header.content}
              </Table.HeaderCell>
            ))}
          </Table.Row>
        </Table.Header>

        <Table.Body>
          {children}
        </Table.Body>

        <Table.Footer>
          <Table.Row>
            <Table.HeaderCell colSpan={colSpan}>
              <div className="table-footer-content">
                <div className="table-footer-buttons">
                  {footerButtons.map((button, index) => (
                    <Button
                      key={index}
                      size="small"
                      loading={button.loading || loading}
                      onClick={button.onClick}
                      as={button.as}
                      to={button.to}
                      color={button.color}
                      className="table-action-button"
                    >
                      {button.content}
                    </Button>
                  ))}
                </div>

                {totalPages > 1 && (
                  <div className="table-pagination-container">
                    <Pagination
                      activePage={activePage}
                      onPageChange={onPageChange}
                      size="small"
                      siblingRange={1}
                      totalPages={totalPages}
                      className="table-pagination"
                    />
                  </div>
                )}
              </div>
            </Table.HeaderCell>
          </Table.Row>
        </Table.Footer>
      </Table>
    </div>
  );
};

export default BaseTable;
