import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import { Container } from 'semantic-ui-react';
import App from './App';
import Header from './components/Header';
import Footer from './components/Footer';
import 'semantic-ui-css/semantic.min.css';
import './index.css';
import './mobile-override.css';
import './global-table-fixes.css';
import './compact-table-styles.css';
import { UserProvider } from './context/User';
import { ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import { StatusProvider } from './context/Status';
import { ThemeProvider } from './context/Theme';
import './i18n';

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(
  <React.StrictMode>
    <ThemeProvider>
      <StatusProvider>
        <UserProvider>
          <BrowserRouter>
            <Header />
            <Container className={'main-content'}>
              <App />
            </Container>
            <ToastContainer />
            <Footer />
          </BrowserRouter>
        </UserProvider>
      </StatusProvider>
    </ThemeProvider>
  </React.StrictMode>
);
