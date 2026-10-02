import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '../store/useAuthStore';
import { api } from '../api/axios';

export const Login = () => {
  const [username, setUsername] = useState('KernelStingray');
  const [password, setPassword] = useState('KeRSt!n90$8!');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  
  const setAuth = useAuthStore((state) => state.setAuth);
  const navigate = useNavigate();

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError('');

    try {
      const response = await api.post('/backoffice/admins/login', {
        login: username,
        password: password,
      });
      
      const { admin, token } = response.data;
      setAuth(token, { id: admin.id, name: admin.login, role: admin.role });
      navigate('/');
    } catch (err: any) {
      if (err.response?.status === 401) {
        setError('Неверный логин или пароль');
      } else {
        setError('Ошибка при авторизации');
      }
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-muted/30 p-4">
      <div className="w-full max-w-md bg-card rounded-xl shadow-lg border p-8">
        <div className="text-center mb-8">
          <h1 className="text-2xl font-bold text-foreground">YolYolakay Admin</h1>
          <p className="text-muted-foreground mt-2">Вход в панель управления</p>
        </div>

        <form onSubmit={handleLogin} className="space-y-4">
          {error && (
            <div className="bg-destructive/10 text-destructive text-sm p-3 rounded-md">
              {error}
            </div>
          )}
          
          <div className="space-y-2">
            <label className="text-sm font-medium text-foreground" htmlFor="username">
              Логин
            </label>
            <input
              id="username"
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="w-full px-3 py-2 border rounded-md bg-background focus:outline-none focus:ring-2 focus:ring-primary/50"
              placeholder="Введите логин"
              required
            />
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium text-foreground" htmlFor="password">
              Пароль
            </label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full px-3 py-2 border rounded-md bg-background focus:outline-none focus:ring-2 focus:ring-primary/50"
              placeholder="Введите пароль"
              required
            />
          </div>

          <button
            type="submit"
            disabled={isLoading}
            className="w-full bg-primary text-primary-foreground py-2 rounded-md font-medium hover:bg-primary/90 transition-colors disabled:opacity-50"
          >
            {isLoading ? 'Вход...' : 'Войти'}
          </button>
        </form>
      </div>
    </div>
  );
};
