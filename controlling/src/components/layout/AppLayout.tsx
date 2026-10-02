import { Outlet, Link, useNavigate, useLocation } from 'react-router-dom';
import { useAuthStore } from '../../store/useAuthStore';
import { ThemeToggle } from './ThemeToggle';
import { 
  LayoutDashboard, 
  Users, 
  CreditCard, 
  Server, 
  Settings, 
  Map, 
  Package, 
  LogOut 
} from 'lucide-react';

export const AppLayout = () => {
  const { logout, user } = useAuthStore();
  const navigate = useNavigate();
  const location = useLocation();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const navItems = [
    { name: 'Дашборд', path: '/', icon: LayoutDashboard },
    { name: 'Пользователи (KYC)', path: '/users', icon: Users },
    { name: 'Биллинг', path: '/billing', icon: CreditCard },
    { name: 'Инфраструктура', path: '/infrastructure', icon: Server },
    { name: 'Карта', path: '/map', icon: Map },
    { name: 'Посылки', path: '/parcels', icon: Package },
    { name: 'Настройки', path: '/settings', icon: Settings },
  ];

  const adminName = user?.name || 'KernelStingray';
  const adminRole = user?.role || 'Super Admin';

  return (
    <div className="flex h-screen bg-muted/20 text-foreground transition-colors duration-300">
      {/* Sidebar */}
      <aside className="w-64 bg-background/95 backdrop-blur-md border-r border-border/50 flex flex-col transition-all duration-300 shadow-[2px_0_10px_-3px_rgba(0,0,0,0.1)] z-20">
        <div className="h-16 flex items-center px-6 border-b border-border/50 bg-gradient-to-r from-primary/10 to-transparent">
          <h1 className="text-xl font-bold text-primary tracking-tight bg-clip-text text-transparent bg-gradient-to-r from-primary to-primary/70">
            YolYolakay Admin
          </h1>
        </div>
        <nav className="flex-1 overflow-y-auto py-6">
          <ul className="space-y-1.5 px-3">
            {navItems.map((item) => {
              const isActive = location.pathname === item.path || 
                               (item.path !== '/' && location.pathname.startsWith(item.path));
              return (
                <li key={item.path}>
                  <Link
                    to={item.path}
                    className={`group flex items-center gap-3 px-3 py-2.5 rounded-lg font-medium text-sm transition-all duration-300 ${
                      isActive 
                        ? 'bg-primary text-primary-foreground shadow-md shadow-primary/20 scale-[1.02]' 
                        : 'text-muted-foreground hover:bg-muted/60 hover:text-foreground hover:scale-[1.01]'
                    }`}
                  >
                    <item.icon className={`h-5 w-5 transition-transform duration-300 ${isActive ? 'scale-110' : 'group-hover:scale-110 group-hover:text-primary/70'}`} />
                    {item.name}
                  </Link>
                </li>
              );
            })}
          </ul>
        </nav>
        <div className="p-4 border-t border-border/50 bg-muted/5">
          <button 
            onClick={handleLogout}
            className="flex items-center justify-center gap-3 px-3 py-2.5 w-full rounded-lg text-sm font-medium text-destructive hover:bg-destructive/10 hover:shadow-sm transition-all duration-300 group"
          >
            <LogOut className="h-5 w-5 transition-transform group-hover:-translate-x-1" />
            Выйти
          </button>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 flex flex-col min-w-0 overflow-hidden relative">
        <header className="h-16 bg-background/80 backdrop-blur-lg border-b border-border/50 flex items-center justify-between px-6 transition-all duration-300 sticky top-0 z-10 shadow-sm">
          <div className="font-semibold text-lg tracking-tight bg-clip-text text-transparent bg-gradient-to-r from-foreground to-foreground/80">
            {navItems.find(i => i.path === location.pathname)?.name || 'Панель управления'}
          </div>
          <div className="flex items-center gap-6">
            <ThemeToggle />
            <div className="h-8 w-px bg-border/60 hidden sm:block" />
            <div className="flex items-center gap-3 cursor-pointer group hover:bg-muted/30 p-1.5 rounded-lg transition-colors">
              <div className="text-right hidden sm:block">
                <div className="text-sm font-bold leading-none text-foreground/90 group-hover:text-foreground">{adminName}</div>
                <div className="text-xs text-primary font-medium mt-1">{adminRole}</div>
              </div>
              <div className="h-10 w-10 rounded-xl bg-gradient-to-br from-primary to-primary/60 text-primary-foreground shadow-md flex items-center justify-center font-bold text-sm transform transition-transform group-hover:scale-105">
                {adminName.charAt(0).toUpperCase()}
              </div>
            </div>
          </div>
        </header>
        <div className="flex-1 overflow-auto p-6 bg-gradient-to-br from-background via-muted/10 to-muted/20">
          <div className="max-w-7xl mx-auto">
            <Outlet />
          </div>
        </div>
      </main>
    </div>
  );
};
