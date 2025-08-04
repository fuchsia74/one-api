import React, { useEffect, useCallback } from 'react';
import './FixedPagination.css';

const FixedPagination = ({ activePage, totalPages, onPageChange }) => {
  const handlePrevious = useCallback((e) => {
    e.preventDefault();
    e.stopPropagation();
    if (activePage > 1 && onPageChange) {
      onPageChange(null, { activePage: activePage - 1 });
    }
  }, [activePage, onPageChange]);

  const handleNext = useCallback((e) => {
    e.preventDefault();
    e.stopPropagation();
    if (activePage < totalPages && onPageChange) {
      onPageChange(null, { activePage: activePage + 1 });
    }
  }, [activePage, totalPages, onPageChange]);

  // Add global click handler to prevent interference
  useEffect(() => {
    const preventBubbling = (e) => {
      if (e.target.closest('.fixed-pagination-container')) {
        e.stopPropagation();
      }
    };

    document.addEventListener('click', preventBubbling, true);
    return () => {
      document.removeEventListener('click', preventBubbling, true);
    };
  }, []);

  // Always show pagination to display page info
  if (!totalPages || totalPages < 1) {
    return null;
  }

  return (
    <div className="fixed-pagination-container">
      <button
        className={`pagination-button ${activePage <= 1 ? 'disabled' : ''}`}
        onClick={handlePrevious}
        disabled={activePage <= 1}
        type="button"
        aria-label="Previous page"
      >
        ‹
      </button>

      <span className="page-info">
        {activePage} / {totalPages}
      </span>

      <button
        className={`pagination-button ${activePage >= totalPages ? 'disabled' : ''}`}
        onClick={handleNext}
        disabled={activePage >= totalPages}
        type="button"
        aria-label="Next page"
      >
        ›
      </button>
    </div>
  );
};

export default FixedPagination;
