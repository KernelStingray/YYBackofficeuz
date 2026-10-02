import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '../api/axios';
import { Search, Package, MapPin, CheckCircle, Truck, Clock, Inbox } from 'lucide-react';

export const Parcels = () => {
  const [search, setSearch] = useState('');

  const { data: parcels, isLoading, error } = useQuery({
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

  const filteredParcels = (parcels || []).filter((p: any) => 
    (p.id && p.id.toLowerCase().includes(search.toLowerCase())) || 
    (p.receiver_phone && p.receiver_phone.includes(search)) ||
    (p.sender_id && p.sender_id.toLowerCase().includes(search.toLowerCase()))
  );

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'delivered': return <CheckCircle className="h-4 w-4 text-green-500" />;
      case 'in_transit': 
      case 'picked_up': return <Truck className="h-4 w-4 text-blue-500" />;
      case 'accepted': return <Truck className="h-4 w-4 text-purple-500" />;
      default: return <Clock className="h-4 w-4 text-orange-500" />;
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'delivered': return 'bg-green-500/10 text-green-600 dark:text-green-400 border-green-500/20';
      case 'in_transit':
      case 'picked_up': return 'bg-blue-500/10 text-blue-600 dark:text-blue-400 border-blue-500/20';
      case 'accepted': return 'bg-purple-500/10 text-purple-600 dark:text-purple-400 border-purple-500/20';
      default: return 'bg-orange-500/10 text-orange-600 dark:text-orange-400 border-orange-500/20';
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case 'delivered': return 'Доставлено';
      case 'in_transit': return 'В пути';
      case 'picked_up': return 'Забрана курьером';
      case 'accepted': return 'Принято';
      case 'pending': return 'Ожидает курьера';
      default: return status || 'Ожидает';
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Доставка посылок</h2>
          <p className="text-sm text-muted-foreground mt-1">Отслеживание и мониторинг междугородних доставок</p>
        </div>
      </div>

      <div className="flex items-center gap-2 max-w-sm">
        <div className="relative w-full">
          <Search className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
          <input
            type="text"
            placeholder="Поиск по ID, Отправителю или Телефону..."
            className="w-full pl-9 pr-4 py-2 bg-card border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
      </div>

      <div className="rounded-xl border bg-card overflow-hidden shadow-sm">
        <table className="w-full text-sm text-left">
          <thead className="bg-muted/50 text-muted-foreground font-medium">
            <tr>
              <th className="px-4 py-3">ID Посылки</th>
              <th className="px-4 py-3">Дата создания</th>
              <th className="px-4 py-3">Отправитель</th>
              <th className="px-4 py-3">Получатель</th>
              <th className="px-4 py-3">Курьер (Водитель)</th>
              <th className="px-4 py-3">Стоимость</th>
              <th className="px-4 py-3 text-right">Статус</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border/40">
            {isLoading ? (
              <tr>
                <td colSpan={7} className="px-4 py-12 text-center text-muted-foreground">
                  <div className="inline-block animate-spin rounded-full h-6 w-6 border-b-2 border-primary mb-2"></div>
                  <p>Загрузка посылок...</p>
                </td>
              </tr>
            ) : error ? (
              <tr>
                <td colSpan={7} className="px-4 py-8 text-center text-red-500">
                  Ошибка при загрузке: {(error as any).message}
                </td>
              </tr>
            ) : filteredParcels.length === 0 ? (
              <tr>
                <td colSpan={7} className="px-4 py-12 text-center text-muted-foreground">
                  <Inbox className="h-10 w-10 mx-auto mb-2 opacity-40" />
                  <p className="font-medium text-base">Посылки не найдены</p>
                  <p className="text-xs text-muted-foreground mt-1">В системе пока нет активных или завершенных посылок</p>
                </td>
              </tr>
            ) : (
              filteredParcels.map((pkg: any) => (
                <tr key={pkg.id} className="hover:bg-muted/30 transition-colors">
                  <td className="px-4 py-3 font-mono text-xs text-muted-foreground flex items-center gap-2">
                    <Package className="h-4 w-4 text-primary" />
                    {pkg.id ? pkg.id.substring(0, 8) + '...' : '—'}
                  </td>
                  <td className="px-4 py-3 text-muted-foreground text-xs font-mono">
                    {pkg.created_at ? new Date(pkg.created_at).toLocaleString() : '—'}
                  </td>
                  <td className="px-4 py-3 font-mono text-xs" title={pkg.sender_id}>
                    {pkg.sender_id ? pkg.sender_id.substring(0, 8) : '—'}
                  </td>
                  <td className="px-4 py-3 flex items-center gap-1 font-mono text-xs">
                    <MapPin className="h-3.5 w-3.5 text-muted-foreground" />
                    {pkg.receiver_phone || 'Скрыт'}
                  </td>
                  <td className="px-4 py-3 text-muted-foreground font-mono text-xs">
                    {pkg.driver_id ? pkg.driver_id.substring(0, 8) : 'Не назначен'}
                  </td>
                  <td className="px-4 py-3 font-semibold">
                    {pkg.price ? `${Number(pkg.price).toLocaleString()} UZS` : '—'}
                  </td>
                  <td className="px-4 py-3 text-right">
                    <span className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium border capitalize ${getStatusBadge(pkg.status)}`}>
                      {getStatusIcon(pkg.status)}
                      {getStatusText(pkg.status)}
                    </span>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};
