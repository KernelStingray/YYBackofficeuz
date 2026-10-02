import { create } from 'zustand';

interface User {
  id: string;
  name: string;
  role: string;
}

interface AuthState {
  token: string | null;
  user: User | null;
  isAuthenticated: boolean;
  setAuth: (token: string, user: User) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>((set) => {
  const token = localStorage.getItem('token');
  const user = localStorage.getItem('user');

  return {
    token: token || null,
    user: user ? JSON.parse(user) : null,
    isAuthenticated: !!token,

    setAuth: (token, user) => {
      localStorage.setItem('token', token);
      localStorage.setItem('user', JSON.stringify(user));
      set({ token, user, isAuthenticated: true });
    },

    logout: () => {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      set({ token: null, user: null, isAuthenticated: false });
    },
  };
});
