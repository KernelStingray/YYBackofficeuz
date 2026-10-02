import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../api/axios';
import { 
  Search, 
  CheckCircle2, 
  XCircle, 
  Eye,
  ShieldCheck,
  FileText,
  UserCheck,
  Inbox,
  ChevronLeft,
  ChevronRight
} from 'lucide-react';
import { UserProfileModal } from '../components/UserProfileModal';

export const Users = () => {
  const [activeTab, setActiveTab] = useState<'users' | 'kyc'>('users');
  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [page, setPage] = useState(1);
  const limit = 20;
  
  const [selectedUser, setSelectedUser] = useState<any>(null);
  const queryClient = useQueryClient();

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedSearch(search);
      setPage(1); // reset page on search
    }, 500);
    return () => clearTimeout(handler);
  }, [search]);

  const approveKycMutation = useMutation({
    mutationFn: (userId: string) => api.post(`/users/admin/kyc/${userId}/approve`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      queryClient.invalidateQueries({ queryKey: ['kyc_users'] });
      setSelectedUser(null);
    }
  });

  const rejectKycMutation = useMutation({
    mutationFn: (userId: string) => api.post(`/users/admin/kyc/${userId}/reject`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      queryClient.invalidateQueries({ queryKey: ['kyc_users'] });
      setSelectedUser(null);
    }
  });

  // Fetch paginated general users
  const { data: usersData, isLoading: isLoadingUsers, error: usersError } = useQuery({
    queryKey: ['users', page, limit, debouncedSearch],
    queryFn: async () => {
      const params = new URLSearchParams();
      params.append('page', page.toString());
      params.append('limit', limit.toString());
      if (debouncedSearch) params.append('search', debouncedSearch);
      
      const response = await api.get(`/users/admin/search?${params.toString()}`);
      return response.data;
    }
  });

  // Fetch KYC pending users
  const { data: kycData, isLoading: isLoadingKyc } = useQuery({
    queryKey: ['kyc_users'],
    queryFn: async () => {
      // kycStatus=pending
      const response = await api.get('/users/admin/search?page=1&limit=100&kyc_status=pending');
      return response.data;
    }
  });

  const users = usersData?.users || [];
  const totalCount = usersData?.total_count || 0;
  const totalPages = Math.ceil(totalCount / limit);

  const kycPendingUsers = kycData?.users || [];
  const totalKycCount = kycData?.total_count || 0;

  // Fetch online statuses for current users
  const activeUsersToFetch = activeTab === 'users' ? users : kycPendingUsers;
  const userIds = activeUsersToFetch.map((u: any) => u.id).join(',');
  const { data: onlineStatuses } = useQuery({
    queryKey: ['online_statuses', userIds],
    queryFn: async () => {
      if (!userIds) return {};
      const response = await api.get(`/notifications/admin/online-status?user_ids=${userIds}`);
      return response.data;
    },
    enabled: !!userIds,
    refetchInterval: 15000,
  });

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Пользователи и Модерация</h2>
          <p className="text-sm text-muted-foreground mt-1">Управление аккаунтами пользователей и заявками на верификацию водителей</p>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex space-x-1 border-b border-border/60">
        <button
          onClick={() => setActiveTab('users')}
          className={`px-4 py-2 font-medium text-sm transition-colors relative ${
            activeTab === 'users' ? 'text-primary' : 'text-muted-foreground hover:text-foreground'
          }`}
        >
          Все пользователи
          {activeTab === 'users' && (
            <div className="absolute bottom-[-1px] left-0 right-0 h-[2px] bg-primary rounded-t-full" />
          )}
        </button>
        <button
          onClick={() => setActiveTab('kyc')}
          className={`px-4 py-2 font-medium text-sm transition-colors relative flex items-center ${
            activeTab === 'kyc' ? 'text-primary' : 'text-muted-foreground hover:text-foreground'
          }`}
        >
          Заявки KYC
          {totalKycCount > 0 && (
            <span className="ml-2 bg-amber-500 text-white text-[10px] px-2 py-0.5 rounded-full font-bold">
              {totalKycCount}
            </span>
          )}
          {activeTab === 'kyc' && (
            <div className="absolute bottom-[-1px] left-0 right-0 h-[2px] bg-primary rounded-t-full" />
          )}
        </button>
      </div>

      {activeTab === 'users' && (
        <div className="space-y-4">
          <div className="flex items-center gap-2 max-w-sm">
            <div className="relative w-full">
              <Search className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
              <input
                type="text"
                placeholder="Поиск по имени или телефону..."
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
                  <th className="px-4 py-3">ID</th>
                  <th className="px-4 py-3">Имя</th>
                  <th className="px-4 py-3">Телефон</th>
                  <th className="px-4 py-3">Роль</th>
                  <th className="px-4 py-3">Статус</th>
                  <th className="px-4 py-3 text-right">Действия</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/40">
                {isLoadingUsers ? (
                  <tr>
                    <td colSpan={6} className="px-4 py-12 text-center text-muted-foreground">
                      <div className="inline-block animate-spin rounded-full h-6 w-6 border-b-2 border-primary mb-2"></div>
                      <p>Загрузка списка пользователей...</p>
                    </td>
                  </tr>
                ) : usersError ? (
                  <tr>
                    <td colSpan={6} className="px-4 py-8 text-center text-red-500">
                      Ошибка при загрузке: {(usersError as any).message}
                    </td>
                  </tr>
                ) : users.length === 0 ? (
                  <tr>
                    <td colSpan={6} className="px-4 py-12 text-center text-muted-foreground">
                      <Inbox className="h-10 w-10 mx-auto mb-2 opacity-40" />
                      <p className="font-medium text-base">Пользователи не найдены</p>
                      <p className="text-xs text-muted-foreground mt-1">В системе пока нет записей или поисковый запрос не дал результатов</p>
                    </td>
                  </tr>
                ) : (
                  users.map((user: any) => (
                    <tr key={user.id} className="hover:bg-muted/30 transition-colors">
                      <td className="px-4 py-3 font-mono text-xs text-muted-foreground">{user.id?.substring(0, 8)}...</td>
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-2 font-medium">
                          <div className={`w-2 h-2 rounded-full ${onlineStatuses?.[user.id] ? 'bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.6)]' : 'bg-gray-300 dark:bg-gray-600'}`} title={onlineStatuses?.[user.id] ? 'В сети' : 'Оффлайн'} />
                          {user.first_name || 'Без'} {user.last_name || 'имени'}
                          {user.role === 'driver' && (
                            <span title="Водитель">
                              <ShieldCheck className="h-4 w-4 text-primary" />
                            </span>
                          )}
                        </div>
                      </td>
                      <td className="px-4 py-3 font-mono text-xs">{user.phone_number || '—'}</td>
                      <td className="px-4 py-3 capitalize">
                        <span className={`px-2 py-0.5 rounded text-xs font-semibold ${
                          user.role === 'driver' ? 'bg-blue-500/10 text-blue-600 dark:text-blue-400' : 'bg-slate-500/10 text-slate-600 dark:text-slate-400'
                        }`}>
                          {user.role === 'driver' ? 'Водитель' : 'Пассажир'}
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <span className={`px-2.5 py-1 rounded-full text-xs font-medium border ${
                          user.status === 'blocked' 
                            ? 'bg-red-500/10 text-red-600 border-red-500/20' 
                            : 'bg-green-500/10 text-green-600 border-green-500/20'
                        }`}>
                          {user.status === 'blocked' ? 'Заблокирован' : 'Активен'}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-right">
                        <button 
                          onClick={() => setSelectedUser(user)}
                          className="p-1.5 rounded-lg hover:bg-muted text-muted-foreground hover:text-foreground transition-colors"
                          title="Просмотреть детально"
                        >
                          <Eye className="h-4 w-4" />
                        </button>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
            
            {/* Pagination Controls */}
            {totalPages > 1 && (
              <div className="px-4 py-3 border-t flex items-center justify-between bg-muted/20">
                <p className="text-sm text-muted-foreground">
                  Показано {(page - 1) * limit + 1} - {Math.min(page * limit, totalCount)} из {totalCount}
                </p>
                <div className="flex items-center gap-2">
                  <button 
                    onClick={() => setPage(p => Math.max(1, p - 1))}
                    disabled={page === 1}
                    className="p-1.5 rounded border bg-card text-muted-foreground hover:bg-muted disabled:opacity-50 transition-colors"
                  >
                    <ChevronLeft className="h-4 w-4" />
                  </button>
                  <span className="text-sm font-medium px-2">
                    {page} из {totalPages}
                  </span>
                  <button 
                    onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                    disabled={page === totalPages}
                    className="p-1.5 rounded border bg-card text-muted-foreground hover:bg-muted disabled:opacity-50 transition-colors"
                  >
                    <ChevronRight className="h-4 w-4" />
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {activeTab === 'kyc' && (
        <div className="space-y-4">
          <div className="rounded-xl border bg-card overflow-hidden shadow-sm">
            <table className="w-full text-sm text-left">
              <thead className="bg-muted/50 text-muted-foreground font-medium">
                <tr>
                  <th className="px-4 py-3">Пользователь</th>
                  <th className="px-4 py-3">Телефон</th>
                  <th className="px-4 py-3">Статус Верификации</th>
                  <th className="px-4 py-3 text-right">Действия</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/40">
                {isLoadingKyc ? (
                   <tr>
                   <td colSpan={4} className="px-4 py-12 text-center text-muted-foreground">
                     <div className="inline-block animate-spin rounded-full h-6 w-6 border-b-2 border-primary mb-2"></div>
                     <p>Загрузка...</p>
                   </td>
                 </tr>
                ) : kycPendingUsers.length === 0 ? (
                  <tr>
                    <td colSpan={4} className="px-4 py-12 text-center text-muted-foreground">
                      <UserCheck className="h-10 w-10 mx-auto mb-2 opacity-40 text-green-500" />
                      <p className="font-medium text-base">Нет активных заявок на KYC</p>
                      <p className="text-xs text-muted-foreground mt-1">Все кандидаты в водители рассмотрены</p>
                    </td>
                  </tr>
                ) : (
                  kycPendingUsers.map((user: any) => (
                    <tr key={user.id} className="hover:bg-muted/30 transition-colors">
                      <td className="px-4 py-3 font-medium">
                        {user.first_name} {user.last_name}
                      </td>
                      <td className="px-4 py-3 font-mono text-xs">{user.phone_number}</td>
                      <td className="px-4 py-3">
                        <span className="bg-amber-500/10 text-amber-600 dark:text-amber-400 px-2.5 py-1 rounded-full text-xs font-semibold border border-amber-500/20">
                          На рассмотрении
                        </span>
                      </td>
                      <td className="px-4 py-3 text-right">
                        <button 
                          onClick={() => setSelectedUser(user)}
                          className="flex items-center gap-1.5 bg-primary text-primary-foreground px-3 py-1.5 rounded-lg hover:bg-primary/90 transition-colors text-xs font-medium ml-auto"
                        >
                          <FileText className="h-4 w-4" />
                          Проверить
                        </button>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Modal */}
      {selectedUser && (
        <UserProfileModal 
          userId={selectedUser.id} 
          onClose={() => setSelectedUser(null)} 
          isKycTab={activeTab === 'kyc'}
        />
      )}
    </div>
  );
};
