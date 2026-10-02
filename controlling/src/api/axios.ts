import axios from 'axios';
import { useAuthStore } from '../store/useAuthStore';

export const getApiUrl = (): string => {
  if (import.meta.env.VITE_API_URL) {
    return import.meta.env.VITE_API_URL;
  }
  if (typeof window !== 'undefined') {
    const hostname = window.location.hostname;
    if (hostname.endsWith('yolyolakay.com')) {
      return 'https://api.yolyolakay.com/api/v1';
    }
  }
  return 'http://localhost:8080/api/v1';
};

export const api = axios.create({
  baseURL: getApiUrl(),
  headers: {
    'Content-Type': 'application/json',
  },
});

api.interceptors.request.use((config) => {
  const token = useAuthStore.getState().token;
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

api.interceptors.response.use(
  (response) => response,
  (error) => {
    // If it's a 401 from any endpoint EXCEPT login, then logout and reload
    if (error.response?.status === 401 && !error.config.url.includes('/admins/login')) {
      useAuthStore.getState().logout();
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);
