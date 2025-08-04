import React, { useContext, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { UserContext } from '../context/User';
import { useTheme } from '../hooks/useTheme';
import { useTranslation } from 'react-i18next';

import {
  Button,
  Container,
  Dropdown,
  Icon,
  Menu,
  Segment,
} from 'semantic-ui-react';
import {
  API,
  getLogo,
  getSystemName,
  isAdmin,
  isMobile,
  showSuccess,
} from '../helpers';
import '../index.css';

// Priority buttons - always visible on medium+ screens
let priorityButtons = [
  {
    name: 'header.channel',
    to: '/channel',
    icon: 'sitemap',
    admin: true,
  },
  {
    name: 'header.token',
    to: '/token',
    icon: 'key',
  },
  {
    name: 'header.dashboard',
    to: '/dashboard',
    icon: 'chart bar',
  },
  {
    name: 'header.log',
    to: '/log',
    icon: 'book',
  },
];

// Secondary buttons - shown in "More" dropdown on medium screens, inline on large screens
let secondaryButtons = [
  {
    name: 'header.user',
    to: '/user',
    icon: 'user',
    admin: true,
  },
  {
    name: 'header.redemption',
    to: '/redemption',
    icon: 'dollar sign',
    admin: true,
  },
  {
    name: 'header.topup',
    to: '/topup',
    icon: 'cart',
  },
  {
    name: 'header.setting',
    to: '/setting',
    icon: 'setting',
  },
  {
    name: 'header.about',
    to: '/about',
    icon: 'info circle',
  },
];

// Add chat button to priority if chat_link exists
if (localStorage.getItem('chat_link')) {
  priorityButtons.splice(1, 0, {
    name: 'header.chat',
    to: '/chat',
    icon: 'comments',
  });
}

// All buttons combined for mobile and large screens
let allButtons = [...priorityButtons, ...secondaryButtons];

// Hook to detect screen size
const useScreenSize = () => {
  const [screenSize, setScreenSize] = useState('large');

  React.useEffect(() => {
    const checkScreenSize = () => {
      const width = window.innerWidth;
      if (width < 768) {
        setScreenSize('mobile');
      } else if (width < 1400) {
        setScreenSize('medium');
      } else {
        setScreenSize('large');
      }
    };

    checkScreenSize();
    window.addEventListener('resize', checkScreenSize);
    return () => window.removeEventListener('resize', checkScreenSize);
  }, []);

  return screenSize;
};

const Header = () => {
  const { t, i18n } = useTranslation();
  const [userState, userDispatch] = useContext(UserContext);
  const { state: themeState, setTheme } = useTheme();
  let navigate = useNavigate();

  const [showSidebar, setShowSidebar] = useState(false);
  const systemName = getSystemName();
  const logo = getLogo();
  const screenSize = useScreenSize();

  async function logout() {
    setShowSidebar(false);
    await API.get('/api/user/logout');
    showSuccess('Logout successful!');
    userDispatch({ type: 'logout' });
    localStorage.removeItem('user');
    navigate('/login');
  }

  const toggleSidebar = () => {
    setShowSidebar(!showSidebar);
  };

  const renderRightMenuItems = () => {
    if (screenSize === 'large') {
      // Large screens: show language, theme, and user dropdown
      return (
        <>
          <Dropdown
            item
            trigger={
              <Icon name='language' style={{ margin: 0, fontSize: '18px' }} />
            }
            options={languageOptions}
            value={i18n.language}
            onChange={(_, { value }) => changeLanguage(value)}
            style={{
              fontSize: '16px',
              fontWeight: '400',
              color: 'var(--text-secondary)',
              padding: '0 10px',
            }}
          />
          <Dropdown
            item
            className="theme-dropdown"
            trigger={
              <span style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                <Icon
                  name={
                    themeState.theme === 'system'
                      ? 'desktop'
                      : themeState.effectiveTheme === 'light'
                      ? 'sun'
                      : 'moon'
                  }
                  style={{ margin: 0, fontSize: '18px' }}
                />
              </span>
            }
            style={{
              fontSize: '16px',
              fontWeight: '400',
              color: 'var(--text-secondary)',
              padding: '0 10px',
            }}
          >
            <Dropdown.Menu
              style={{
                backgroundColor: 'var(--card-bg)',
                border: '1px solid var(--border-color)',
                boxShadow: '0 4px 12px var(--shadow-color)',
                minWidth: '140px'
              }}
            >
              <Dropdown.Item
                onClick={() => setTheme('light')}
                active={themeState.theme === 'light'}
                style={{
                  backgroundColor: themeState.theme === 'light' ? 'var(--button-primary)' : 'var(--card-bg)',
                  color: themeState.theme === 'light' ? 'white' : 'var(--text-primary)',
                  padding: '12px 16px',
                  borderBottom: '1px solid var(--border-color)',
                  fontWeight: themeState.theme === 'light' ? '600' : '400',
                  cursor: 'pointer'
                }}
              >
                <Icon name="sun" style={{ marginRight: '8px' }} />
                {t('header.theme.light')}
              </Dropdown.Item>
              <Dropdown.Item
                onClick={() => setTheme('dark')}
                active={themeState.theme === 'dark'}
                style={{
                  backgroundColor: themeState.theme === 'dark' ? 'var(--button-primary)' : 'var(--card-bg)',
                  color: themeState.theme === 'dark' ? 'white' : 'var(--text-primary)',
                  padding: '12px 16px',
                  borderBottom: '1px solid var(--border-color)',
                  fontWeight: themeState.theme === 'dark' ? '600' : '400',
                  cursor: 'pointer'
                }}
              >
                <Icon name="moon" style={{ marginRight: '8px' }} />
                {t('header.theme.dark')}
              </Dropdown.Item>
              <Dropdown.Item
                onClick={() => setTheme('system')}
                active={themeState.theme === 'system'}
                style={{
                  backgroundColor: themeState.theme === 'system' ? 'var(--button-primary)' : 'var(--card-bg)',
                  color: themeState.theme === 'system' ? 'white' : 'var(--text-primary)',
                  padding: '12px 16px',
                  borderBottom: 'none',
                  fontWeight: themeState.theme === 'system' ? '600' : '400',
                  cursor: 'pointer'
                }}
              >
                <Icon name="desktop" style={{ marginRight: '8px' }} />
                {t('header.theme.system')}
              </Dropdown.Item>
            </Dropdown.Menu>
          </Dropdown>
          {userState.user ? (
            <Dropdown
              text={userState.user.username}
              pointing
              className='link item'
              style={{
                fontSize: '15px',
                fontWeight: '400',
                color: 'var(--text-secondary)',
              }}
            >
              <Dropdown.Menu>
                <Dropdown.Item
                  onClick={logout}
                  style={{
                    fontSize: '15px',
                    fontWeight: '400',
                    color: 'var(--text-secondary)',
                  }}
                >
                  {t('header.logout')}
                </Dropdown.Item>
              </Dropdown.Menu>
            </Dropdown>
          ) : (
            <Menu.Item
              name={t('header.login')}
              as={Link}
              to='/login'
              className='btn btn-link'
              style={{
                fontSize: '15px',
                fontWeight: '400',
                color: 'var(--text-secondary)',
              }}
            />
          )}
        </>
      );
    } else {
      // Medium screens: show "More" dropdown with secondary buttons + settings
      const secondaryItems = secondaryButtons.filter(button => !button.admin || isAdmin());

      return (
        <Dropdown
          item
          trigger={
            <Icon
              name="ellipsis horizontal"
              style={{
                margin: 0,
                fontSize: '18px',
                color: 'var(--text-secondary)',
              }}
            />
          }
          className="header-more-dropdown"
        >
          <Dropdown.Menu>
            {secondaryItems.map((button) => (
              <Dropdown.Item
                key={button.name}
                as={Link}
                to={button.to}
                style={{
                  fontSize: '15px',
                  fontWeight: '400',
                  color: 'var(--text-secondary)',
                }}
              >
                <Icon name={button.icon} style={{ marginRight: '8px' }} />
                {t(button.name)}
              </Dropdown.Item>
            ))}
            {secondaryItems.length > 0 && <Dropdown.Divider />}
            <Dropdown.Item
              style={{
                fontSize: '15px',
                fontWeight: '400',
                color: 'var(--text-secondary)',
              }}
            >
              <Icon name="language" style={{ marginRight: '8px' }} />
              <Dropdown
                trigger={<span>{t('header.language')}</span>}
                options={languageOptions}
                value={i18n.language}
                onChange={(_, { value }) => changeLanguage(value)}
                direction="left"
                style={{ border: 'none', background: 'none' }}
              />
            </Dropdown.Item>
            <Dropdown.Item
              style={{
                fontSize: '15px',
                fontWeight: '400',
                color: 'var(--text-secondary)',
              }}
            >
              <Icon name={
                themeState.theme === 'system'
                  ? 'desktop'
                  : themeState.effectiveTheme === 'light'
                  ? 'sun'
                  : 'moon'
              } style={{ marginRight: '8px' }} />
              <Dropdown
                trigger={<span>{t('header.theme.' + themeState.theme)}</span>}
                direction="left"
                style={{ border: 'none', background: 'none' }}
              >
                <Dropdown.Menu>
                  <Dropdown.Item onClick={() => setTheme('light')}>
                    <Icon name="sun" style={{ marginRight: '8px' }} />
                    {t('header.theme.light')}
                  </Dropdown.Item>
                  <Dropdown.Item onClick={() => setTheme('dark')}>
                    <Icon name="moon" style={{ marginRight: '8px' }} />
                    {t('header.theme.dark')}
                  </Dropdown.Item>
                  <Dropdown.Item onClick={() => setTheme('system')}>
                    <Icon name="desktop" style={{ marginRight: '8px' }} />
                    {t('header.theme.system')}
                  </Dropdown.Item>
                </Dropdown.Menu>
              </Dropdown>
            </Dropdown.Item>
            <Dropdown.Divider />
            {userState.user ? (
              <Dropdown.Item
                onClick={logout}
                style={{
                  fontSize: '15px',
                  fontWeight: '400',
                  color: 'var(--text-secondary)',
                }}
              >
                <Icon name="user" style={{ marginRight: '8px' }} />
                {t('header.logout')} ({userState.user.username})
              </Dropdown.Item>
            ) : (
              <Dropdown.Item
                as={Link}
                to='/login'
                style={{
                  fontSize: '15px',
                  fontWeight: '400',
                  color: 'var(--text-secondary)',
                }}
              >
                <Icon name="sign in" style={{ marginRight: '8px' }} />
                {t('header.login')}
              </Dropdown.Item>
            )}
          </Dropdown.Menu>
        </Dropdown>
      );
    }
  };

  const renderButton = (button, isMobile = false, iconOnly = false) => {
    if (button.admin && !isAdmin()) return null;

    if (isMobile) {
      return (
        <Menu.Item
          key={button.name}
          onClick={() => {
            navigate(button.to);
            setShowSidebar(false);
          }}
          style={{ fontSize: '15px' }}
        >
          {t(button.name)}
        </Menu.Item>
      );
    }

    if (iconOnly) {
      return (
        <Menu.Item
          key={button.name}
          as={Link}
          to={button.to}
          style={{
            fontSize: '15px',
            fontWeight: '400',
            color: 'var(--text-secondary)',
            padding: '0.8em',
          }}
          title={t(button.name)} // Tooltip for accessibility
        >
          <Icon name={button.icon} style={{ margin: 0, fontSize: '18px' }} />
        </Menu.Item>
      );
    }

    return (
      <Menu.Item
        key={button.name}
        as={Link}
        to={button.to}
        style={{
          fontSize: '15px',
          fontWeight: '400',
          color: 'var(--text-secondary)',
        }}
      >
        <Icon name={button.icon} style={{ marginRight: '4px' }} />
        {t(button.name)}
      </Menu.Item>
    );
  };

  const renderButtons = (isMobile) => {
    if (isMobile) {
      // Mobile: show all buttons in sidebar
      return allButtons.map((button) => renderButton(button, true));
    }

    if (screenSize === 'large') {
      // Large screens: show all buttons with text
      return allButtons.map((button) => renderButton(button, false, false));
    }

    // Medium screens: show only priority buttons with icons only
    return priorityButtons.map((button) => renderButton(button, false, true));
  };



  // Add language switcher dropdown
  const languageOptions = [
    { key: 'zh', text: '中文', value: 'zh' },
    { key: 'en', text: 'English', value: 'en' },
  ];

  const changeLanguage = (language) => {
    i18n.changeLanguage(language);
  };

  if (isMobile()) {
    return (
      <>
        <Menu
          borderless
          size='large'
          style={
            showSidebar
              ? {
                  borderBottom: 'none',
                  marginBottom: '0',
                  borderTop: 'none',
                  height: '51px',
                }
              : { borderTop: 'none', height: '52px' }
          }
        >
          <Container
            style={{
              width: '100%',
              maxWidth: isMobile() ? '100%' : '1200px',
              padding: isMobile() ? '0 10px' : '0 20px',
            }}
          >
            <Menu.Item as={Link} to='/'>
              <img src={logo} alt='logo' style={{ marginRight: '0.75em' }} />
              <div style={{ fontSize: '20px' }}>
                <b>{systemName}</b>
              </div>
            </Menu.Item>
            <Menu.Menu position='right'>
              <Menu.Item onClick={toggleSidebar}>
                <Icon name={showSidebar ? 'close' : 'sidebar'} />
              </Menu.Item>
            </Menu.Menu>
          </Container>
        </Menu>
        {showSidebar ? (
          <Segment style={{ marginTop: 0, borderTop: '0' }}>
            <Menu secondary vertical style={{ width: '100%', margin: 0 }}>
              {renderButtons(true)}
              <Menu.Item>
                <Dropdown
                  selection
                  trigger={
                    <Icon
                      name='language'
                      style={{ margin: 0, fontSize: '18px' }}
                    />
                  }
                  options={languageOptions}
                  value={i18n.language}
                  onChange={(_, { value }) => changeLanguage(value)}
                />
              </Menu.Item>
              <Menu.Item>
                <Dropdown
                  className="theme-dropdown"
                  trigger={
                    <Button className="theme-toggle-button">
                      <Icon
                        name={
                          themeState.theme === 'system'
                            ? 'desktop'
                            : themeState.effectiveTheme === 'light'
                            ? 'sun'
                            : 'moon'
                        }
                        style={{ margin: 0, fontSize: '18px' }}
                      />
                    </Button>
                  }
                >
                  <Dropdown.Menu
                    style={{
                      backgroundColor: 'var(--card-bg)',
                      border: '1px solid var(--border-color)',
                      boxShadow: '0 4px 12px var(--shadow-color)',
                      minWidth: '140px'
                    }}
                  >
                    <Dropdown.Item
                      onClick={() => setTheme('light')}
                      active={themeState.theme === 'light'}
                      style={{
                        backgroundColor: themeState.theme === 'light' ? 'var(--button-primary)' : 'var(--card-bg)',
                        color: themeState.theme === 'light' ? 'white' : 'var(--text-primary)',
                        padding: '12px 16px',
                        borderBottom: '1px solid var(--border-color)',
                        fontWeight: themeState.theme === 'light' ? '600' : '400',
                        cursor: 'pointer'
                      }}
                    >
                      <Icon name="sun" style={{ marginRight: '8px' }} />
                      {t('header.theme.light')}
                    </Dropdown.Item>
                    <Dropdown.Item
                      onClick={() => setTheme('dark')}
                      active={themeState.theme === 'dark'}
                      style={{
                        backgroundColor: themeState.theme === 'dark' ? 'var(--button-primary)' : 'var(--card-bg)',
                        color: themeState.theme === 'dark' ? 'white' : 'var(--text-primary)',
                        padding: '12px 16px',
                        borderBottom: '1px solid var(--border-color)',
                        fontWeight: themeState.theme === 'dark' ? '600' : '400',
                        cursor: 'pointer'
                      }}
                    >
                      <Icon name="moon" style={{ marginRight: '8px' }} />
                      {t('header.theme.dark')}
                    </Dropdown.Item>
                    <Dropdown.Item
                      onClick={() => setTheme('system')}
                      active={themeState.theme === 'system'}
                      style={{
                        backgroundColor: themeState.theme === 'system' ? 'var(--button-primary)' : 'var(--card-bg)',
                        color: themeState.theme === 'system' ? 'white' : 'var(--text-primary)',
                        padding: '12px 16px',
                        borderBottom: 'none',
                        fontWeight: themeState.theme === 'system' ? '600' : '400',
                        cursor: 'pointer'
                      }}
                    >
                      <Icon name="desktop" style={{ marginRight: '8px' }} />
                      {t('header.theme.system')}
                    </Dropdown.Item>
                  </Dropdown.Menu>
                </Dropdown>
              </Menu.Item>
              <Menu.Item>
                {userState.user ? (
                  <Button onClick={logout} style={{ color: 'var(--text-secondary)' }}>
                    {t('header.logout')}
                  </Button>
                ) : (
                  <>
                    <Button
                      onClick={() => {
                        setShowSidebar(false);
                        navigate('/login');
                      }}
                    >
                      {t('header.login')}
                    </Button>
                    <Button
                      onClick={() => {
                        setShowSidebar(false);
                        navigate('/register');
                      }}
                    >
                      {t('header.register')}
                    </Button>
                  </>
                )}
              </Menu.Item>
            </Menu>
          </Segment>
        ) : (
          <></>
        )}
      </>
    );
  }

  return (
    <>
      <Menu
        borderless
        style={{
          borderTop: 'none',
          boxShadow: 'rgba(0, 0, 0, 0.04) 0px 2px 12px 0px',
          border: 'none',
        }}
      >
        <Container
          style={{
            width: '100%',
            maxWidth: isMobile() ? '100%' : '1200px',
            padding: isMobile() ? '0 10px' : '0 20px',
          }}
        >
          <Menu.Item as={Link} to='/'>
            <img src={logo} alt='logo' style={{ marginRight: '0.75em' }} />
            <div
              style={{
                fontSize: '18px',
                fontWeight: '500',
                color: 'var(--text-primary)',
              }}
            >
              {systemName}
            </div>
          </Menu.Item>
          {renderButtons(false)}
          <Menu.Menu position='right'>
            {renderRightMenuItems()}
          </Menu.Menu>
        </Container>
      </Menu>
    </>
  );
};

export default Header;
