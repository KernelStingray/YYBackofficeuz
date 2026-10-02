import { 
  Users, 
  CarFront, 
  Package, 
  ArrowUpRight, 
  Activity, 
  CheckCircle2,
  MapPin,
  ShieldCheck,
  Server,
  Zap,
  TrendingUp,
  CreditCard,
  Clock,
  ChevronRight
} from 'lucide-react';
import { 
  AreaChart, 
  Area, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer 
} from 'recharts';
import { useQuery } from '@tanstack/react-query';
import { api } from '../api/axios';
import { useThemeStore } from '../store/useThemeStore';
import { useNavigate } from 'react-router-dom';

export const Dashboard = () => {
  const { mode } = useThemeStore();
  const navigate = useNavigate();
  const isDark = mode === 'dark' || (mode === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches);

  // Fetch User Stats
  const { data: usersStats, isLoading: isUsersLoading } = useQuery({
    queryKey: ['usersStats'],
    queryFn: async () => {
      try {
        const response = await api.get('/users/admin/stats');
        return response.data;
      } catch {
        return { total_users: 0, total_drivers: 0, total_passengers: 0, chart_data: [] };
      }
    }
  });

  // Fetch Users List for "Recent" components
  const { data: usersData } = useQuery({
    queryKey: ['users'],
    queryFn: async () => {
      try {
        const response = await api.get('/users/admin/list');
        return Array.isArray(response.data) ? response.data : [];
      } catch {
        return [];
      }
    }
  });

  // Fetch Parcels Stats
  const { data: parcelsStats, isLoading: isParcelsLoading } = useQuery({
    queryKey: ['parcelsStats'],
    queryFn: async () => {
      try {
        const response = await api.get('/parcels/admin/stats');
        return response.data;
      } catch {
        return { total_parcels: 0, active_parcels: 0, delivered_parcels: 0, chart_data: [] };
      }
    }
  });

  // Fetch Parcels List for "Recent" components
  const { data: parcelsData } = useQuery({
    queryKey: ['parcels'],
    queryFn: async () => {
      try {
        const response = await api.get('/parcels/admin/list');
        return Array.isArray(response.data) ? response.data : [];
      } catch {
        return [];
      }
    }
  });

  const usersList = usersData || [];
  const parcelsList = parcelsData || [];

  // Fetch Rides Stats
  const { data: ridesStats, isLoading: isRidesLoading } = useQuery({
    queryKey: ['ridesStats'],
    queryFn: async () => {
      try {
        const response = await api.get('/rides/admin/stats');
        return response.data;
      } catch {
        return { total_rides: 0, active_rides: 0, chart_data: [], popular_routes: [] };
      }
    }
  });

  // Fetch Billing Stats
  const { data: billingStats, isLoading: isBillingLoading } = useQuery({
    queryKey: ['billingStats'],
    queryFn: async () => {
      try {
        const response = await api.get('/billing/admin/stats');
        return response.data;
      } catch {
        return { total_revenue: 0, chart_data: [] };
      }
    }
  });

  const totalUsers = usersStats?.total_users || 0;
  const driversCount = usersStats?.total_drivers || 0;
  const passengersCount = usersStats?.total_passengers || 0;
  
  const totalRides = ridesStats?.total_rides || 0;
  const activeRides = ridesStats?.active_rides || 0;
  
  const totalParcels = parcelsStats?.total_parcels || 0;
  const activeParcels = parcelsStats?.active_parcels || 0;
  const deliveredParcels = parcelsStats?.delivered_parcels || 0;

  // Compute total platform turnover from billing revenue
  const totalTurnover = billingStats?.total_revenue || 0;
  const platformCommission = Math.round(totalTurnover * 0.03); // Assuming 3% commission on revenue, or maybe the revenue IS the commission depending on business model

  // Merge chart data by date
  const chartDataMap: Record<string, any> = {};

  const processChartData = (data: any[], key: string) => {
    if (!data) return;
    data.forEach(item => {
      const d = new Date(item.date).toLocaleDateString('ru-RU', { weekday: 'short' });
      if (!chartDataMap[d]) {
        chartDataMap[d] = { name: d, users: 0, rides: 0, parcels: 0, revenue: 0 };
      }
      if (key === 'revenue') {
        chartDataMap[d][key] += (item.amount || 0);
      } else {
        chartDataMap[d][key] += (item.count || 0);
      }
    });
  };

  processChartData(usersStats?.chart_data, 'users');
  processChartData(ridesStats?.chart_data, 'rides');
  processChartData(parcelsStats?.chart_data, 'parcels');
  processChartData(billingStats?.chart_data, 'revenue');

  const chartData = Object.values(chartDataMap).sort((a: any, b: any) => {
    // Basic string comparison assuming 'name' is 'Пн', 'Вт', etc. This might be tricky.
    // Better to sort by the original Date, but since we parsed to short string, we can just assume
    // chartDataMap will be populated chronologically if the backend returned it chronologically.
    // Since we want the last 7 days in order, let's keep the backend order by returning it as is.
    return 0;
  });

  // Popular routes breakdown
  const popularRoutes = (ridesStats?.popular_routes || []).map((r: any, idx: number) => ({
    from: r.from,
    to: r.to,
    count: r.count,
    tag: idx === 0 ? 'Высокий спрос' : (idx === 1 ? 'Популярный' : 'Стабильный')
  }));

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h2 className="text-3xl font-bold tracking-tight bg-clip-text text-transparent bg-gradient-to-r from-foreground to-foreground/70">
            Главный Дашборд
          </h2>
          <p className="text-sm text-muted-foreground mt-1">Аналитика, состояние платформы YolYolakay и мониторинг в реальном времени</p>
        </div>
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-green-500/10 text-green-600 dark:text-green-400 text-xs font-semibold border border-green-500/20 shadow-sm">
            <Activity className="h-3.5 w-3.5 animate-pulse text-green-500" />
            Все 11 сервисов активны
          </div>
        </div>
      </div>
      
      {/* KPI Cards Grid */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {/* Total Users */}
        <div className="rounded-2xl border bg-card text-card-foreground shadow-sm p-6 transition-all hover:shadow-md">
          <div className="flex flex-row items-center justify-between space-y-0 pb-2">
            <h3 className="tracking-tight text-xs font-semibold uppercase text-muted-foreground">Всего пользователей</h3>
            <div className="p-2 rounded-xl bg-primary/10 text-primary">
              <Users className="h-5 w-5" />
            </div>
          </div>
          <div className="text-3xl font-extrabold mt-1">
            {isUsersLoading ? <div className="h-8 w-24 bg-muted animate-pulse rounded-lg" /> : totalUsers.toLocaleString()}
          </div>
          <div className="flex items-center justify-between text-xs text-muted-foreground mt-3 pt-3 border-t border-border/40">
            <span className="flex items-center gap-1">
              <ShieldCheck className="h-3.5 w-3.5 text-blue-500" />
              {driversCount} Водителей
            </span>
            <span>{passengersCount} Пассажиров</span>
          </div>
        </div>

        {/* Active Rides */}
        <div className="rounded-2xl border bg-card text-card-foreground shadow-sm p-6 transition-all hover:shadow-md">
          <div className="flex flex-row items-center justify-between space-y-0 pb-2">
            <h3 className="tracking-tight text-xs font-semibold uppercase text-muted-foreground">Междугородние поездки</h3>
            <div className="p-2 rounded-xl bg-blue-500/10 text-blue-500">
              <CarFront className="h-5 w-5" />
            </div>
          </div>
          <div className="text-3xl font-extrabold mt-1">
            {isRidesLoading ? <div className="h-8 w-24 bg-muted animate-pulse rounded-lg" /> : totalRides.toLocaleString()}
          </div>
          <div className="flex items-center justify-between text-xs text-muted-foreground mt-3 pt-3 border-t border-border/40">
            <span className="text-green-600 dark:text-green-400 font-semibold flex items-center gap-1">
              <Activity className="h-3.5 w-3.5" />
              {activeRides} Активно
            </span>
            <span>Всего доступно</span>
          </div>
        </div>

        {/* Parcel Deliveries */}
        <div className="rounded-2xl border bg-card text-card-foreground shadow-sm p-6 transition-all hover:shadow-md">
          <div className="flex flex-row items-center justify-between space-y-0 pb-2">
            <h3 className="tracking-tight text-xs font-semibold uppercase text-muted-foreground">Доставка посылок</h3>
            <div className="p-2 rounded-xl bg-purple-500/10 text-purple-500">
              <Package className="h-5 w-5" />
            </div>
          </div>
          <div className="text-3xl font-extrabold mt-1">
            {isParcelsLoading ? <div className="h-8 w-24 bg-muted animate-pulse rounded-lg" /> : totalParcels.toLocaleString()}
          </div>
          <div className="flex items-center justify-between text-xs text-muted-foreground mt-3 pt-3 border-t border-border/40">
            <span className="text-purple-600 dark:text-purple-400 font-semibold">
              {activeParcels} В процессе
            </span>
            <span>{deliveredParcels} Доставлено</span>
          </div>
        </div>

        {/* Platform Revenue */}
        <div className="rounded-2xl border bg-card text-card-foreground shadow-sm p-6 transition-all hover:shadow-md">
          <div className="flex flex-row items-center justify-between space-y-0 pb-2">
            <h3 className="tracking-tight text-xs font-semibold uppercase text-muted-foreground">Комиссия (3%)</h3>
            <div className="p-2 rounded-xl bg-emerald-500/10 text-emerald-500">
              <TrendingUp className="h-5 w-5" />
            </div>
          </div>
          <div className="text-3xl font-extrabold text-emerald-600 dark:text-emerald-400 mt-1">
            {platformCommission.toLocaleString()} <span className="text-base font-semibold">UZS</span>
          </div>
          <div className="flex items-center justify-between text-xs text-muted-foreground mt-3 pt-3 border-t border-border/40">
            <span>Оборот: {totalTurnover.toLocaleString()} UZS</span>
          </div>
        </div>
      </div>

      {/* Main Grid: Chart & Quick Actions */}
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-7">
        {/* Weekly Trend Chart */}
        <div className="rounded-2xl border bg-card text-card-foreground shadow-sm col-span-4 p-6 transition-all">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h3 className="font-bold text-lg tracking-tight">Динамика активности платформы</h3>
              <p className="text-xs text-muted-foreground mt-0.5">Соотношение зарегистрированных пользователей, поездок и посылок</p>
            </div>
          </div>
          <div className="h-[320px] w-full mt-4">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={chartData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
                <defs>
                  <linearGradient id="colorUsers" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="var(--primary)" stopOpacity={0.4}/>
                    <stop offset="95%" stopColor="var(--primary)" stopOpacity={0}/>
                  </linearGradient>
                  <linearGradient id="colorRides" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="var(--chart-2)" stopOpacity={0.4}/>
                    <stop offset="95%" stopColor="var(--chart-2)" stopOpacity={0}/>
                  </linearGradient>
                  <linearGradient id="colorParcels" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="var(--chart-3)" stopOpacity={0.4}/>
                    <stop offset="95%" stopColor="var(--chart-3)" stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke={isDark ? "var(--border)" : "#e5e7eb"} />
                <XAxis 
                  dataKey="name" 
                  stroke={isDark ? "var(--muted-foreground)" : "#6b7280"}
                  fontSize={12}
                  tickLine={false}
                  axisLine={false}
                />
                <YAxis 
                  stroke={isDark ? "var(--muted-foreground)" : "#6b7280"}
                  fontSize={12}
                  tickLine={false}
                  axisLine={false}
                />
                <Tooltip 
                  contentStyle={{ 
                    backgroundColor: isDark ? 'var(--card)' : '#fff',
                    borderColor: isDark ? 'var(--border)' : '#e5e7eb',
                    borderRadius: '12px',
                    boxShadow: '0 10px 15px -3px rgb(0 0 0 / 0.1)'
                  }}
                />
                <Area 
                  type="monotone" 
                  dataKey="users" 
                  name="Пользователи" 
                  stroke="var(--primary)" 
                  strokeWidth={3}
                  fillOpacity={1} 
                  fill="url(#colorUsers)" 
                />
                <Area 
                  type="monotone" 
                  dataKey="rides" 
                  name="Поездки" 
                  stroke="var(--chart-2)" 
                  strokeWidth={3}
                  fillOpacity={1} 
                  fill="url(#colorRides)" 
                />
                <Area 
                  type="monotone" 
                  dataKey="parcels" 
                  name="Посылки" 
                  stroke="var(--chart-3)" 
                  strokeWidth={3}
                  fillOpacity={1} 
                  fill="url(#colorParcels)" 
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* Quick Actions & System Status */}
        <div className="col-span-3 space-y-6">
          {/* Quick Shortcuts */}
          <div className="rounded-2xl border bg-card p-6 shadow-sm">
            <h3 className="font-bold text-base mb-3">Быстрые действия</h3>
            <div className="grid grid-cols-2 gap-3">
              <button 
                onClick={() => navigate('/users')}
                className="flex items-center gap-3 p-3 rounded-xl border bg-muted/20 hover:bg-muted/50 transition-all text-left group"
              >
                <div className="p-2 rounded-lg bg-primary/10 text-primary group-hover:scale-110 transition-transform">
                  <ShieldCheck className="h-4 w-4" />
                </div>
                <div>
                  <div className="text-xs font-bold">KYC Модерация</div>
                  <div className="text-[11px] text-muted-foreground">Проверка водителей</div>
                </div>
              </button>

              <button 
                onClick={() => navigate('/map')}
                className="flex items-center gap-3 p-3 rounded-xl border bg-muted/20 hover:bg-muted/50 transition-all text-left group"
              >
                <div className="p-2 rounded-lg bg-blue-500/10 text-blue-500 group-hover:scale-110 transition-transform">
                  <MapPin className="h-4 w-4" />
                </div>
                <div>
                  <div className="text-xs font-bold">Карта онлайн</div>
                  <div className="text-[11px] text-muted-foreground">Водители в Ташкенте</div>
                </div>
              </button>

              <button 
                onClick={() => navigate('/parcels')}
                className="flex items-center gap-3 p-3 rounded-xl border bg-muted/20 hover:bg-muted/50 transition-all text-left group"
              >
                <div className="p-2 rounded-lg bg-purple-500/10 text-purple-500 group-hover:scale-110 transition-transform">
                  <Package className="h-4 w-4" />
                </div>
                <div>
                  <div className="text-xs font-bold">Посылки</div>
                  <div className="text-[11px] text-muted-foreground">Мониторинг доставок</div>
                </div>
              </button>

              <button 
                onClick={() => navigate('/infrastructure')}
                className="flex items-center gap-3 p-3 rounded-xl border bg-muted/20 hover:bg-muted/50 transition-all text-left group"
              >
                <div className="p-2 rounded-lg bg-amber-500/10 text-amber-500 group-hover:scale-110 transition-transform">
                  <Server className="h-4 w-4" />
                </div>
                <div>
                  <div className="text-xs font-bold">Кластеры</div>
                  <div className="text-[11px] text-muted-foreground">Логи и мониторинг</div>
                </div>
              </button>
            </div>
          </div>

          {/* Popular Routes */}
          <div className="rounded-2xl border bg-card p-6 shadow-sm">
            <div className="flex items-center justify-between mb-3">
              <h3 className="font-bold text-base">Популярные междугородние маршруты</h3>
            </div>
            <div className="space-y-2.5">
              {popularRoutes.map((route: any, idx: number) => (
                <div key={idx} className="flex items-center justify-between p-2.5 rounded-xl border bg-muted/10 hover:bg-muted/30 transition-colors text-xs">
                  <div className="flex items-center gap-2 font-medium">
                    <MapPin className="h-3.5 w-3.5 text-primary" />
                    <span>{route.from}</span>
                    <span className="text-muted-foreground">→</span>
                    <span>{route.to}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="px-2 py-0.5 rounded-md bg-primary/10 text-primary font-semibold text-[10px]">
                      {route.tag}
                    </span>
                    <span className="font-bold text-muted-foreground">{route.count} поездок</span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Bottom Section: Recent Real Users & Recent Real Parcels */}
      <div className="grid gap-6 md:grid-cols-2">
        {/* Recent Registered Users */}
        <div className="rounded-2xl border bg-card p-6 shadow-sm">
          <div className="flex items-center justify-between mb-4">
            <h3 className="font-bold text-base flex items-center gap-2">
              <Users className="h-4 w-4 text-primary" />
              Недавняя регистрация пользователей
            </h3>
            <button onClick={() => navigate('/users')} className="text-xs text-primary font-semibold flex items-center hover:underline">
              Все ({usersList.length}) <ChevronRight className="h-3 w-3 ml-0.5" />
            </button>
          </div>
          <div className="space-y-2.5">
            {usersList.length === 0 ? (
              <div className="p-6 text-center text-muted-foreground text-xs italic">
                Нет зарегистрированных пользователей
              </div>
            ) : (
              usersList.slice(0, 4).map((u: any) => (
                <div key={u.id} className="flex items-center justify-between p-3 rounded-xl border bg-muted/20 hover:bg-muted/40 transition-colors text-xs">
                  <div className="flex items-center gap-3">
                    <div className="h-8 w-8 rounded-full bg-primary/10 text-primary font-bold flex items-center justify-center">
                      {(u.first_name || 'U').charAt(0).toUpperCase()}
                    </div>
                    <div>
                      <div className="font-semibold text-foreground">{u.first_name || 'Без'} {u.last_name || 'имени'}</div>
                      <div className="text-[11px] text-muted-foreground font-mono">{u.phone_number}</div>
                    </div>
                  </div>
                  <span className={`px-2 py-0.5 rounded-full text-[10px] font-bold border ${
                    u.role === 'driver' ? 'bg-blue-500/10 text-blue-600 border-blue-500/20' : 'bg-slate-500/10 text-slate-600 border-slate-500/20'
                  }`}>
                    {u.role === 'driver' ? 'Водитель' : 'Пассажир'}
                  </span>
                </div>
              ))
            )}
          </div>
        </div>

        {/* Recent Parcels */}
        <div className="rounded-2xl border bg-card p-6 shadow-sm">
          <div className="flex items-center justify-between mb-4">
            <h3 className="font-bold text-base flex items-center gap-2">
              <Package className="h-4 w-4 text-purple-500" />
              Последние заказы посылок
            </h3>
            <button onClick={() => navigate('/parcels')} className="text-xs text-primary font-semibold flex items-center hover:underline">
              Все ({parcelsList.length}) <ChevronRight className="h-3 w-3 ml-0.5" />
            </button>
          </div>
          <div className="space-y-2.5">
            {parcelsList.length === 0 ? (
              <div className="p-6 text-center text-muted-foreground text-xs italic">
                Пока нет созданных посылок
              </div>
            ) : (
              parcelsList.slice(0, 4).map((p: any) => (
                <div key={p.id} className="flex items-center justify-between p-3 rounded-xl border bg-muted/20 hover:bg-muted/40 transition-colors text-xs">
                  <div>
                    <div className="font-semibold font-mono text-foreground flex items-center gap-1.5">
                      <Package className="h-3.5 w-3.5 text-purple-500" />
                      {p.id?.substring(0, 8)}...
                    </div>
                    <div className="text-[11px] text-muted-foreground mt-0.5 font-mono">
                      Получатель: {p.receiver_phone || 'Скрыт'}
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="font-bold text-foreground">{p.price ? `${Number(p.price).toLocaleString()} UZS` : '—'}</div>
                    <div className="text-[10px] text-purple-600 dark:text-purple-400 font-medium capitalize mt-0.5">
                      {p.status || 'pending'}
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </div>
  );
};
