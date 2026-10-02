import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '../api/axios';
import { Search, FileSpreadsheet, Inbox, Wallet } from 'lucide-react';

export const Billing = () => {
  const [activeTab, setActiveTab] = useState<'transactions' | 'balances'>('transactions');
  const [search, setSearch] = useState('');

  // Fetch real users (to render balances of drivers)
  const { data: users, isLoading: isUsersLoading } = useQuery({
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

  // Filter drivers for balance view
  const driversList = (users || []).filter((u: any) => u.role === 'driver');

  const filteredDrivers = driversList.filter((d: any) => {
    const query = search.toLowerCase();
    const name = `${d.first_name || ''} ${d.last_name || ''}`.toLowerCase();
    const phone = d.phone_number || '';
    return name.includes(query) || phone.includes(query) || (d.id && d.id.includes(query));
  });

  const handleExportCSV = () => {
    const data = activeTab === 'balances' ? filteredDrivers : [];
    if (data.length === 0) return;

    const headers = Object.keys(data[0]).join(',');
    const rows = data.map(obj => Object.values(obj).join(',')).join('\n');
    const csvContent = "data:text/csv;charset=utf-8," + headers + "\n" + rows;

    const encodedUri = encodeURI(csvContent);
    const link = document.createElement('a');
    link.setAttribute('href', encodedUri);
    link.setAttribute('download', `${activeTab}_export_${new Date().toISOString().slice(0, 10)}.csv`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Биллинг и Финансы</h2>
          <p className="text-sm text-muted-foreground mt-1">Управление балансами водителей и финансовой отчетностью</p>
        </div>
        <button 
          onClick={handleExportCSV}
          className="flex items-center gap-2 bg-secondary text-secondary-foreground border px-4 py-2 rounded-lg hover:bg-secondary/80 transition-colors font-medium text-sm"
        >
          <FileSpreadsheet className="h-4 w-4" />
          Скачать CSV
        </button>
      </div>

      <div className="grid gap-4 md:grid-cols-3 mb-6">
        <div className="rounded-xl border bg-card p-6 shadow-sm">
          <div className="text-sm font-medium pb-2 text-muted-foreground">
            Активных водителей
          </div>
          <div className="text-2xl font-bold text-primary">{driversList.length}</div>
        </div>
        <div className="rounded-xl border bg-card p-6 shadow-sm">
          <div className="text-sm font-medium pb-2 text-muted-foreground">
            Общая сумма пополнений
          </div>
          <div className="text-2xl font-bold">0 UZS</div>
        </div>
        <div className="rounded-xl border bg-card p-6 shadow-sm">
          <div className="text-sm font-medium pb-2 text-muted-foreground">
            Задолженность (Сумма)
          </div>
          <div className="text-2xl font-bold text-muted-foreground">0 UZS</div>
        </div>
      </div>

      <div className="flex space-x-1 border-b border-border/60">
        <button
          onClick={() => setActiveTab('transactions')}
          className={`px-4 py-2 font-medium text-sm transition-colors relative ${
            activeTab === 'transactions' ? 'text-primary' : 'text-muted-foreground hover:text-foreground'
          }`}
        >
          История транзакций
          {activeTab === 'transactions' && (
            <div className="absolute bottom-[-1px] left-0 right-0 h-[2px] bg-primary rounded-t-full" />
          )}
        </button>
        <button
          onClick={() => setActiveTab('balances')}
          className={`px-4 py-2 font-medium text-sm transition-colors relative ${
            activeTab === 'balances' ? 'text-primary' : 'text-muted-foreground hover:text-foreground'
          }`}
        >
          Балансы водителей ({driversList.length})
          {activeTab === 'balances' && (
            <div className="absolute bottom-[-1px] left-0 right-0 h-[2px] bg-primary rounded-t-full" />
          )}
        </button>
      </div>

      <div className="flex items-center gap-2 max-w-sm">
        <div className="relative w-full">
          <Search className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
          <input
            type="text"
            placeholder={activeTab === 'transactions' ? 'Поиск транзакций...' : 'Поиск водителя...'}
            className="w-full pl-9 pr-4 py-2 bg-card border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
      </div>

      <div className="rounded-xl border bg-card overflow-hidden shadow-sm">
        {activeTab === 'transactions' && (
          <div className="p-12 text-center text-muted-foreground">
            <Inbox className="h-10 w-10 mx-auto mb-2 opacity-40" />
            <p className="font-medium text-base">История транзакций пуста</p>
            <p className="text-xs text-muted-foreground mt-1">Транзакции пополнения и списания комиссий появятся после первых платежей</p>
          </div>
        )}

        {activeTab === 'balances' && (
          <table className="w-full text-sm text-left">
            <thead className="bg-muted/50 text-muted-foreground font-medium">
              <tr>
                <th className="px-4 py-3">ID Водителя</th>
                <th className="px-4 py-3">ФИО</th>
                <th className="px-4 py-3">Телефон</th>
                <th className="px-4 py-3">Текущий баланс (UZS)</th>
                <th className="px-4 py-3 text-right">Действия</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/40">
              {isUsersLoading ? (
                <tr>
                  <td colSpan={5} className="px-4 py-12 text-center text-muted-foreground">
                    <div className="inline-block animate-spin rounded-full h-6 w-6 border-b-2 border-primary mb-2"></div>
                    <p>Загрузка балансов водителей...</p>
                  </td>
                </tr>
              ) : filteredDrivers.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-4 py-12 text-center text-muted-foreground">
                    <Inbox className="h-10 w-10 mx-auto mb-2 opacity-40" />
                    <p className="font-medium text-base">Водители не найдены</p>
                    <p className="text-xs text-muted-foreground mt-1">В системе пока нет зарегистрированных водителей</p>
                  </td>
                </tr>
              ) : (
                filteredDrivers.map((driver: any) => (
                  <tr key={driver.id} className="hover:bg-muted/30 transition-colors">
                    <td className="px-4 py-3 font-mono text-xs text-muted-foreground">{driver.id ? driver.id.substring(0, 8) + '...' : '—'}</td>
                    <td className="px-4 py-3 font-medium">{driver.first_name || 'Без'} {driver.last_name || 'имени'}</td>
                    <td className="px-4 py-3 font-mono text-xs">{driver.phone_number || '—'}</td>
                    <td className="px-4 py-3 font-semibold">0 UZS</td>
                    <td className="px-4 py-3 text-right">
                      <button className="flex items-center justify-center ml-auto gap-1 text-primary hover:text-primary/80 transition-colors text-xs font-medium bg-primary/10 px-2.5 py-1.5 rounded-lg border border-primary/20">
                        <Wallet className="h-3.5 w-3.5" />
                        Управление балансом
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
};
