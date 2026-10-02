import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../api/axios';
import { 
  XCircle, 
  ShieldCheck, 
  CarFront,
  Route,
  Package,
  Ban,
  CheckCircle2,
  CalendarDays,
  Star,
  Image as ImageIcon
} from 'lucide-react';

interface UserProfileModalProps {
  userId: string;
  onClose: () => void;
  isKycTab?: boolean;
}

export const UserProfileModal: React.FC<UserProfileModalProps> = ({ userId, onClose, isKycTab }) => {
  const queryClient = useQueryClient();
  const [rejectReason, setRejectReason] = useState('');
  const [showRejectInput, setShowRejectInput] = useState(false);
  const [enlargedImage, setEnlargedImage] = useState<string | null>(null);

  // 1. Fetch Core Profile & Cars (User Service)
  const { data: profileData, isLoading: isLoadingProfile } = useQuery({
    queryKey: ['user_profile', userId],
    queryFn: async () => {
      const res = await api.get(`/users/${userId}`);
      return res.data;
    }
  });

  // 2. Fetch Ride Stats (Ride Service)
  const { data: rideStats, isLoading: isLoadingRides } = useQuery({
    queryKey: ['user_ride_stats', userId],
    queryFn: async () => {
      const res = await api.get(`/rides/admin/user/${userId}/stats`);
      return res.data;
    }
  });

  // 3. Fetch Parcel Stats (Parcel Service)
  const { data: parcelStats, isLoading: isLoadingParcels } = useQuery({
    queryKey: ['user_parcel_stats', userId],
    queryFn: async () => {
      const res = await api.get(`/parcels/admin/user/${userId}/stats`);
      return res.data;
    }
  });

  // 4. Fetch Active Ride
  const { data: activeRide, isLoading: isLoadingActiveRide } = useQuery({
    queryKey: ['user_active_ride', userId],
    queryFn: async () => {
      try {
        const res = await api.get(`/rides/admin/user/${userId}/active`);
        return res.data;
      } catch (e: any) {
        if (e.response && e.response.status === 404) {
          return null;
        }
        throw e;
      }
    }
  });

  // Status Change Mutation
  const updateStatusMutation = useMutation({
    mutationFn: (status: 'active' | 'blocked') => api.post(`/users/admin/${userId}/status`, { status }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      queryClient.invalidateQueries({ queryKey: ['user_profile', userId] });
    }
  });

  const approveKycMutation = useMutation({
    mutationFn: () => api.post(`/users/admin/kyc/${userId}/approve`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      queryClient.invalidateQueries({ queryKey: ['kyc_users'] });
      onClose();
    }
  });

  const rejectKycMutation = useMutation({
    mutationFn: (reason: string) => api.post(`/users/admin/kyc/${userId}/reject`, { reason }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      queryClient.invalidateQueries({ queryKey: ['kyc_users'] });
      onClose();
    }
  });

  const isLoading = isLoadingProfile || isLoadingRides || isLoadingParcels || isLoadingActiveRide;

  if (isLoading) {
    return (
      <div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4">
         <div className="bg-card w-full max-w-lg rounded-2xl shadow-xl border p-12 flex justify-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
         </div>
      </div>
    );
  }

  const user = profileData?.user;
  const cars = profileData?.cars || [];

  if (!user) return null;

  const handleReject = () => {
    if (!showRejectInput) {
      setShowRejectInput(true);
      return;
    }
    if (!rejectReason.trim()) {
      alert('Укажите причину отклонения');
      return;
    }
    rejectKycMutation.mutate(rejectReason);
  };

  const renderDocumentImage = (label: string, url?: string | null) => {
    if (!url) return null;
    return (
      <div className="border rounded-lg p-2 bg-card">
        <p className="text-xs text-muted-foreground mb-2 font-medium">{label}</p>
        <div 
          className="relative aspect-video bg-muted rounded cursor-pointer overflow-hidden group"
          onClick={() => setEnlargedImage(url)}
        >
          <img src={url} alt={label} className="w-full h-full object-cover group-hover:scale-105 transition-transform" />
          <div className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
            <ImageIcon className="h-6 w-6 text-white" />
          </div>
        </div>
      </div>
    );
  };

  return (
    <>
      <div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4">
        <div className="bg-card w-full max-w-3xl rounded-2xl shadow-xl border overflow-hidden animate-in fade-in zoom-in duration-200 flex flex-col max-h-[90vh]">
          
          {/* Header */}
          <div className="px-6 py-4 border-b flex justify-between items-center bg-muted/20 shrink-0">
            <div className="flex items-center gap-3">
              <h3 className="font-semibold text-lg flex items-center gap-2">
                Профиль пользователя
                {user.role === 'driver' && <span title="Водитель"><ShieldCheck className="h-5 w-5 text-blue-500" /></span>}
              </h3>
              <span className={`px-2.5 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider ${
                user.status === 'blocked' 
                  ? 'bg-red-500/10 text-red-600 border border-red-500/20' 
                  : 'bg-green-500/10 text-green-600 border border-green-500/20'
              }`}>
                {user.status}
              </span>
              {isKycTab && user.kyc_status === 'pending' && (
                <span className="px-2.5 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider bg-amber-500/10 text-amber-600 border border-amber-500/20">
                  KYC Pending
                </span>
              )}
            </div>
            <button onClick={onClose} className="text-muted-foreground hover:text-foreground p-1 rounded-lg hover:bg-muted transition-colors">
              <XCircle className="h-6 w-6" />
            </button>
          </div>

          {/* Scrollable Content */}
          <div className="p-6 overflow-y-auto space-y-8 flex-1 custom-scrollbar">
            
            {/* Active Ride Banner */}
            {!isKycTab && activeRide && (
              <div className="bg-gradient-to-r from-green-500/10 to-emerald-500/5 border border-green-500/20 rounded-xl p-4 flex items-center justify-between animate-in fade-in slide-in-from-top-2">
                <div className="flex items-center gap-3">
                  <div className="relative flex h-3 w-3">
                    <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                    <span className="relative inline-flex rounded-full h-3 w-3 bg-green-500"></span>
                  </div>
                  <div>
                    <h4 className="font-semibold text-green-700 dark:text-green-400">Пользователь сейчас в пути!</h4>
                    <p className="text-sm text-green-600/80 dark:text-green-400/80 mt-0.5">
                      Маршрут: {activeRide.origin_address || 'Неизвестно'} → {activeRide.destination_address || 'Неизвестно'}
                    </p>
                  </div>
                </div>
                <div className="text-right">
                  <span className="text-xs font-mono bg-green-500/10 text-green-600 px-2 py-1 rounded-md border border-green-500/20">
                    ID: {activeRide.id?.substring(0,8)}...
                  </span>
                </div>
              </div>
            )}

            {/* Main Info */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div className="bg-muted/10 p-4 rounded-xl border border-border/50">
                <p className="text-xs text-muted-foreground uppercase tracking-wider font-semibold mb-1">Имя</p>
                <p className="font-medium truncate">{user.first_name || '—'} {user.last_name || ''}</p>
              </div>
              <div className="bg-muted/10 p-4 rounded-xl border border-border/50">
                <p className="text-xs text-muted-foreground uppercase tracking-wider font-semibold mb-1">Телефон</p>
                <p className="font-mono font-medium truncate">{user.phone_number || '—'}</p>
              </div>
              <div className="bg-muted/10 p-4 rounded-xl border border-border/50">
                <p className="text-xs text-muted-foreground uppercase tracking-wider font-semibold mb-1 flex items-center gap-1">
                  <Star className="h-3 w-3" /> Рейтинг
                </p>
                <p className="font-medium">{user.rating?.toFixed(1) || '0.0'} <span className="text-muted-foreground text-xs font-normal">({user.rating_count || 0})</span></p>
              </div>
              <div className="bg-muted/10 p-4 rounded-xl border border-border/50">
                <p className="text-xs text-muted-foreground uppercase tracking-wider font-semibold mb-1 flex items-center gap-1">
                  <CalendarDays className="h-3 w-3" /> Регистрация
                </p>
                <p className="font-medium text-sm">{new Date(user.created_at).toLocaleDateString('ru-RU')}</p>
              </div>
            </div>

            {/* Statistics (hide in KYC mode to focus on docs) */}
            {!isKycTab && (
              <div>
                <h4 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider mb-3">Статистика активности</h4>
                <div className="grid grid-cols-2 gap-4">
                  
                  {/* Rides Card */}
                  <div className="bg-gradient-to-br from-blue-500/5 to-purple-500/5 p-4 rounded-xl border border-blue-500/10 flex items-start gap-4">
                    <div className="bg-blue-500/10 p-3 rounded-lg text-blue-600 dark:text-blue-400">
                      <Route className="h-6 w-6" />
                    </div>
                    <div>
                      <h5 className="font-medium mb-2">Поездки</h5>
                      <div className="space-y-1 text-sm">
                        <div className="flex justify-between gap-4">
                          <span className="text-muted-foreground">Как водитель:</span>
                          <span className="font-mono font-medium">{rideStats?.total_rides_as_driver || 0}</span>
                        </div>
                        <div className="flex justify-between gap-4">
                          <span className="text-muted-foreground">Как пассажир:</span>
                          <span className="font-mono font-medium">{rideStats?.total_rides_as_passenger || 0}</span>
                        </div>
                      </div>
                    </div>
                  </div>

                  {/* Parcels Card */}
                  <div className="bg-gradient-to-br from-amber-500/5 to-orange-500/5 p-4 rounded-xl border border-amber-500/10 flex items-start gap-4">
                    <div className="bg-amber-500/10 p-3 rounded-lg text-amber-600 dark:text-amber-400">
                      <Package className="h-6 w-6" />
                    </div>
                    <div>
                      <h5 className="font-medium mb-2">Посылки</h5>
                      <div className="space-y-1 text-sm">
                        <div className="flex justify-between gap-4">
                          <span className="text-muted-foreground">Отправлено:</span>
                          <span className="font-mono font-medium">{parcelStats?.total_sent || 0}</span>
                        </div>
                        <div className="flex justify-between gap-4">
                          <span className="text-muted-foreground">Доставлено (водитель):</span>
                          <span className="font-mono font-medium">{parcelStats?.total_delivered || 0}</span>
                        </div>
                      </div>
                    </div>
                  </div>

                </div>
              </div>
            )}

            {/* Cars Section & Documents */}
            {cars.length > 0 && (
              <div>
                <h4 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider mb-3 flex items-center gap-2">
                  <CarFront className="h-4 w-4" /> Автомобили и Документы ({cars.length})
                </h4>
                <div className="space-y-6">
                  {cars.map((car: any) => (
                    <div key={car.id} className="border rounded-xl p-5 bg-muted/5">
                      <div className="flex items-center justify-between mb-4 border-b pb-4">
                        <div className="flex items-center gap-4">
                          <div className="bg-muted p-2 rounded-lg">
                            <CarFront className="h-6 w-6 text-muted-foreground" />
                          </div>
                          <div>
                            <p className="font-medium text-base">{car.brand} {car.model} <span className="text-muted-foreground">({car.year})</span></p>
                            <p className="text-sm text-muted-foreground mt-0.5 font-mono">{car.license_plate} • {car.color}</p>
                          </div>
                        </div>
                        <div>
                          <span className={`px-2.5 py-1 rounded-md text-[11px] font-bold uppercase tracking-wider ${
                            car.moderation_status === 'approved' ? 'bg-green-500/10 text-green-600' :
                            car.moderation_status === 'rejected' ? 'bg-red-500/10 text-red-600' :
                            'bg-amber-500/10 text-amber-600'
                          }`}>
                            {car.moderation_status}
                          </span>
                        </div>
                      </div>

                      {/* Documents Grid */}
                      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                        {renderDocumentImage('ВУ (Лицевая)', car.driver_license_front || car.driver_license_photo)}
                        {renderDocumentImage('ВУ (Обратная)', car.driver_license_back)}
                        {renderDocumentImage('Техпаспорт (Лицевая)', car.tech_passport_front || car.tech_passport_photo)}
                        {renderDocumentImage('Техпаспорт (Обратная)', car.tech_passport_back)}
                        {renderDocumentImage('Авто (Спереди)', car.car_front_photo)}
                        {renderDocumentImage('Авто (Сзади)', car.car_back_photo)}
                        {renderDocumentImage('Авто (Сбоку)', car.car_side_photo)}
                      </div>
                      
                      {!car.driver_license_front && !car.tech_passport_front && !car.car_front_photo && !car.tech_passport_photo && !car.driver_license_photo && (
                        <p className="text-sm text-muted-foreground italic text-center py-4 bg-muted/20 rounded">
                          Документы не загружены
                        </p>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            )}

          </div>

          {/* Footer Actions */}
          <div className="px-6 py-4 border-t bg-muted/10 flex flex-col gap-3 shrink-0">
            {showRejectInput && (
              <div className="flex gap-2 w-full animate-in slide-in-from-bottom-2">
                <input 
                  type="text" 
                  autoFocus
                  placeholder="Причина отклонения (например: нечеткое фото прав)..."
                  className="flex-1 px-4 py-2 bg-card border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-destructive/50"
                  value={rejectReason}
                  onChange={(e) => setRejectReason(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && handleReject()}
                />
                <button 
                  onClick={() => setShowRejectInput(false)}
                  className="px-3 py-2 rounded-lg border text-muted-foreground hover:bg-muted text-sm transition-colors"
                >
                  Отмена
                </button>
              </div>
            )}

            <div className="flex justify-between w-full">
              <div>
                {!isKycTab && (user.status === 'active' ? (
                  <button 
                    onClick={() => {
                      if (confirm('Вы уверены, что хотите ЗАБЛОКИРОВАТЬ пользователя? Он потеряет доступ к приложению.')) {
                        updateStatusMutation.mutate('blocked');
                      }
                    }}
                    disabled={updateStatusMutation.isPending}
                    className="px-4 py-2 rounded-lg text-sm font-medium border border-destructive/30 text-destructive hover:bg-destructive/10 transition-colors disabled:opacity-50 flex items-center gap-2"
                  >
                    <Ban className="h-4 w-4" /> Заблокировать
                  </button>
                ) : !isKycTab && (
                  <button 
                    onClick={() => updateStatusMutation.mutate('active')}
                    disabled={updateStatusMutation.isPending}
                    className="px-4 py-2 rounded-lg text-sm font-medium border border-green-500/30 text-green-600 hover:bg-green-500/10 transition-colors disabled:opacity-50 flex items-center gap-2"
                  >
                    <CheckCircle2 className="h-4 w-4" /> Разблокировать
                  </button>
                ))}
              </div>
              
              <div className="flex gap-3">
                {isKycTab && (
                  <>
                    <button 
                      onClick={handleReject}
                      disabled={rejectKycMutation.isPending}
                      className="px-4 py-2 rounded-lg text-sm font-medium border border-destructive/30 text-destructive hover:bg-destructive/10 transition-colors disabled:opacity-50"
                    >
                      Отклонить KYC
                    </button>
                    <button 
                      onClick={() => approveKycMutation.mutate()}
                      disabled={approveKycMutation.isPending || showRejectInput}
                      className="px-4 py-2 rounded-lg text-sm font-medium bg-green-600 text-white hover:bg-green-700 transition-colors disabled:opacity-50 flex items-center gap-2 shadow-sm"
                    >
                      <CheckCircle2 className="h-4 w-4" />
                      Одобрить KYC
                    </button>
                  </>
                )}
                <button 
                  onClick={onClose} 
                  className="px-5 py-2 rounded-lg text-sm font-medium border hover:bg-muted transition-colors bg-card"
                >
                  Закрыть
                </button>
              </div>
            </div>
          </div>

        </div>
      </div>

      {/* Enlarged Image Modal */}
      {enlargedImage && (
        <div 
          className="fixed inset-0 bg-black/90 z-[60] flex items-center justify-center p-4 cursor-pointer animate-in fade-in"
          onClick={() => setEnlargedImage(null)}
        >
          <img 
            src={enlargedImage} 
            alt="Enlarged Document" 
            className="max-w-full max-h-[95vh] object-contain rounded-lg shadow-2xl"
          />
          <button 
            className="absolute top-6 right-6 p-2 bg-black/50 text-white rounded-full hover:bg-black/70 transition-colors"
            onClick={() => setEnlargedImage(null)}
          >
            <XCircle className="h-8 w-8" />
          </button>
        </div>
      )}
    </>
  );
};
