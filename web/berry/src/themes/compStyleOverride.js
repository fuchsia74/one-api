export default function componentStyleOverrides(theme) {
  const bgColor = theme.mode === 'dark' ? theme.backgroundDefault : theme.colors?.grey50;
  return {
    MuiButton: {
      styleOverrides: {
        root: {
          fontWeight: 500,
          borderRadius: '4px',
          '&.Mui-disabled': {
            color: theme.colors?.grey600
          }
        }
      }
    },
    //MuiAutocomplete-popper MuiPopover-root
    MuiAutocomplete: {
      styleOverrides: {
        popper: {
          // 继承 MuiPopover-root
          boxShadow: '0px 5px 5px -3px rgba(0,0,0,0.2),0px 8px 10px 1px rgba(0,0,0,0.14),0px 3px 14px 2px rgba(0,0,0,0.12)',
          borderRadius: '12px',
          color: '#364152'
        },
        listbox: {
          // 继承 MuiPopover-root
          padding: '0px',
          paddingTop: '8px',
          paddingBottom: '8px'
        },
        option: {
          fontSize: '16px',
          fontWeight: '400',
          lineHeight: '1.334em',
          alignItems: 'center',
          paddingTop: '6px',
          paddingBottom: '6px',
          paddingLeft: '16px',
          paddingRight: '16px'
        }
      }
    },
    MuiIconButton: {
      styleOverrides: {
        root: {
          color: theme.darkTextPrimary,
          '&:hover': {
            backgroundColor: theme.colors?.grey200
          }
        }
      }
    },
    MuiPaper: {
      defaultProps: {
        elevation: 0
      },
      styleOverrides: {
        root: {
          backgroundImage: 'none',
          // Responsive styles for narrow screens
          '@media (max-width: 768px)': {
            borderLeft: 'none',
            borderRight: 'none',
            borderRadius: '0',
            boxShadow: 'none',
            margin: '0'
          }
        },
        rounded: {
          borderRadius: `${theme?.customization?.borderRadius}px`,
          '@media (max-width: 768px)': {
            borderRadius: '0'
          }
        }
      }
    },
    MuiCardHeader: {
      styleOverrides: {
        root: {
          color: theme.colors?.textDark,
          padding: '24px'
        },
        title: {
          fontSize: '1.125rem'
        }
      }
    },
    MuiCardContent: {
      styleOverrides: {
        root: {
          padding: '24px',
          // Responsive padding for narrow screens
          '@media (max-width: 768px)': {
            padding: '12px',
            borderLeft: 'none',
            borderRight: 'none',
            borderRadius: '0'
          }
        }
      }
    },
    MuiCardActions: {
      styleOverrides: {
        root: {
          padding: '24px',
          // Responsive padding for narrow screens
          '@media (max-width: 768px)': {
            padding: '4px' /* Minimal padding */
          }
        }
      }
    },
    MuiContainer: {
      styleOverrides: {
        root: {
          // Responsive container styles for narrow screens
          '@media (max-width: 768px)': {
            paddingLeft: '0', /* Remove all padding */
            paddingRight: '0', /* Remove all padding */
            margin: '0'
          },
          '@media (min-width: 769px) and (max-width: 1366px)': {
            paddingLeft: '16px',
            paddingRight: '16px'
          }
        }
      }
    },
    MuiTableContainer: {
      styleOverrides: {
        root: {
          '@media (max-width: 768px)': {
            border: 'none',
            borderRadius: '0',
            boxShadow: 'none',
            margin: '0'
          }
        }
      }
    },
    MuiTable: {
      styleOverrides: {
        root: {
          '@media (max-width: 768px)': {
            borderLeft: 'none',
            borderRight: 'none',
            boxShadow: 'none',
            '& .MuiTableHead-root': {
              display: 'none'
            },
            '& .MuiTableBody-root .MuiTableRow-root': {
              display: 'block',
              border: 'none',
              borderRadius: '0',
              marginBottom: '8px',
              padding: '12px',
              backgroundColor: 'transparent'
            },
            '& .MuiTableBody-root .MuiTableRow-root .MuiTableCell-root': {
              display: 'block',
              border: 'none',
              padding: '4px 0',
              textAlign: 'left',
              position: 'relative',
              paddingLeft: '40%',
              '&:before': {
                position: 'absolute',
                left: '0',
                width: '35%',
                fontWeight: 'bold',
                color: theme.colors?.grey700 || '#666',
                fontSize: '0.9em'
              },
              '&:nth-of-type(1):before': {
                content: '"时间"'
              },
              '&:nth-of-type(2):before': {
                content: '"渠道"'
              },
              '&:nth-of-type(3):before': {
                content: '"类型"'
              },
              '&:nth-of-type(4):before': {
                content: '"模型"'
              },
              '&:nth-of-type(5):before': {
                content: '"用户"'
              },
              '&:nth-of-type(6):before': {
                content: '"令牌"'
              },
              '&:nth-of-type(7):before': {
                content: '"提示"'
              },
              '&:nth-of-type(8):before': {
                content: '"补全"'
              },
              '&:nth-of-type(9):before': {
                content: '"花费"'
              },
              '&:nth-of-type(10):before': {
                content: '"延迟"'
              },
              '&:nth-of-type(11):before': {
                content: '"详情"'
              }
            }
          }
        }
      }
    },
    MuiListItemButton: {
      styleOverrides: {
        root: {
          color: theme.darkTextPrimary,
          paddingTop: '10px',
          paddingBottom: '10px',
          '&.Mui-selected': {
            color: theme.menuSelected,
            backgroundColor: theme.menuSelectedBack,
            '&:hover': {
              backgroundColor: theme.menuSelectedBack
            },
            '& .MuiListItemIcon-root': {
              color: theme.menuSelected
            }
          },
          '&:hover': {
            backgroundColor: theme.menuSelectedBack,
            color: theme.menuSelected,
            '& .MuiListItemIcon-root': {
              color: theme.menuSelected
            }
          }
        }
      }
    },
    MuiListItemIcon: {
      styleOverrides: {
        root: {
          color: theme.darkTextPrimary,
          minWidth: '36px'
        }
      }
    },
    MuiListItemText: {
      styleOverrides: {
        primary: {
          color: theme.textDark
        }
      }
    },
    MuiInputBase: {
      styleOverrides: {
        input: {
          color: theme.textDark,
          '&::placeholder': {
            color: theme.darkTextSecondary,
            fontSize: '0.875rem'
          }
        }
      }
    },
    MuiOutlinedInput: {
      styleOverrides: {
        root: {
          background: bgColor,
          borderRadius: `${theme?.customization?.borderRadius}px`,
          '& .MuiOutlinedInput-notchedOutline': {
            borderColor: theme.colors?.grey400
          },
          '&:hover $notchedOutline': {
            borderColor: theme.colors?.primaryLight
          },
          '&.MuiInputBase-multiline': {
            padding: 1
          }
        },
        input: {
          fontWeight: 500,
          background: bgColor,
          padding: '15.5px 14px',
          borderRadius: `${theme?.customization?.borderRadius}px`,
          '&.MuiInputBase-inputSizeSmall': {
            padding: '10px 14px',
            '&.MuiInputBase-inputAdornedStart': {
              paddingLeft: 0
            }
          }
        },
        inputAdornedStart: {
          paddingLeft: 4
        },
        notchedOutline: {
          borderRadius: `${theme?.customization?.borderRadius}px`
        }
      }
    },
    MuiSlider: {
      styleOverrides: {
        root: {
          '&.Mui-disabled': {
            color: theme.colors?.grey300
          }
        },
        mark: {
          backgroundColor: theme.paper,
          width: '4px'
        },
        valueLabel: {
          color: theme?.colors?.primaryLight
        }
      }
    },
    MuiDivider: {
      styleOverrides: {
        root: {
          borderColor: theme.divider,
          opacity: 1
        }
      }
    },
    MuiAvatar: {
      styleOverrides: {
        root: {
          color: theme.colors?.primaryDark,
          background: theme.colors?.primary200
        }
      }
    },
    MuiChip: {
      styleOverrides: {
        root: {
          '&.MuiChip-deletable .MuiChip-deleteIcon': {
            color: 'inherit'
          }
        }
      }
    },
    MuiTableCell: {
      styleOverrides: {
        root: {
          borderBottom: '1px solid ' + theme.tableBorderBottom,
          textAlign: 'center'
        },
        head: {
          color: theme.darkTextSecondary,
          backgroundColor: theme.headBackgroundColor
        }
      }
    },
    MuiTableRow: {
      styleOverrides: {
        root: {
          '&:hover': {
            backgroundColor: theme.headBackgroundColor
          }
        }
      }
    },
    MuiTooltip: {
      styleOverrides: {
        tooltip: {
          color: theme.colors.paper,
          background: theme.colors?.grey700
        }
      }
    },
    MuiCssBaseline: {
      styleOverrides: `
      .apexcharts-title-text {
          fill: ${theme.textDark} !important
        }
      .apexcharts-text {
        fill: ${theme.textDark} !important
      }
      .apexcharts-legend-text {
        color: ${theme.textDark} !important
      }
      .apexcharts-menu {
        background: ${theme.backgroundDefault} !important
      }
      .apexcharts-gridline, .apexcharts-xaxistooltip-background, .apexcharts-yaxistooltip-background {
        stroke: ${theme.divider} !important;
      }
      `
    }
  };
}
