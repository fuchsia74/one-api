import { showError } from './common';
import axios from 'axios';
import { store } from 'store/index';
import { LOGIN } from 'store/actions';
import config from 'config';

export const API = axios.create({
  baseURL: process.env.REACT_APP_SERVER ? process.env.REACT_APP_SERVER : '/'
});

// Disable caching for all GET /api requests to ensure fresh data
API.interceptors.request.use((config) => {
  if (config.method && config.method.toLowerCase() === 'get' && config.url && config.url.startsWith('/api')) {
    config.headers['Cache-Control'] = 'no-cache, no-store, must-revalidate';
    config.headers['Pragma'] = 'no-cache';
    config.headers['Expires'] = '0';
    try {
      const urlObj = new URL(config.url, window.location.origin);
      urlObj.searchParams.set('_', Date.now().toString());
      config.url = urlObj.pathname + urlObj.search;
    } catch (e) {
      // ignore
    }
  }
  return config;
});

API.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('user');
      store.dispatch({ type: LOGIN, payload: null });
      window.location.href = config.basename + 'login';
    }

    if (error.response?.data?.message) {
      error.message = error.response.data.message;
    }

    showError(error);
  }
);
