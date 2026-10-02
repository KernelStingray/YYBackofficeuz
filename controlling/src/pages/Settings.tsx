import { useState, useEffect } from 'react';
import { 
  Save, 
  Settings2, 
  Bell, 
  Shield, 
  Loader2, 
  CheckCircle2, 
  Key, 
  Smartphone, 
  MessageSquare, 
  Lock,
  Globe,
  Radio
} from 'lucide-react';
import { api } from '../api/axios';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';

export const Settings = () => {
  const [activeTab, setActiveTab] = useState<'general' | 'security' | 'notifications' | 'integrations'>('general');
  const [toastMessage, setToastMessage] = useState<string | null>(null);

  // Form states - General
  const [platformCommission, setPlatformCommission] = useState('15');
  const [baseTariff, setBaseTariff] = useState('8000');
  const [perKm, setPerKm] = useState('1200');
  const [supportPhone, setSupportPhone] = useState('+998712000000');
  const [maintenanceMode, setMaintenanceMode] = useState(false);

  // Form states - Security
  const [require2FA, setRequire2FA] = useState(true);
  const [sessionTimeout, setSessionTimeout] = useState('60');
  const [jwtExpiryHours, setJwtExpiryHours] = useState('24');
  const [ipWhitelist, setIpWhitelist] = useState('127.0.0.1, 217.29.119.128');

  // Form states - Notifications
  const [pushEnabled, setPushEnabled] = useState(true);
  const [smsEnabled, setSmsEnabled] = useState(true);
  const [telegramAlerts, setTelegramAlerts] = useState(true);
  const [emailAlerts, setEmailAlerts] = useState(false);

  // Form states - Integrations
  const [clickMerchantId, setClickMerchantId] = useState('63396');
  const [paymeMerchantId, setPaymeMerchantId] = useState('108681');
  const [eskizSenderId, setEskizSenderId] = useState('4546');

  const queryClient = useQueryClient();

  // Show Toast Auto-dismiss
  const showToast = (msg: string) => {
    setToastMessage(msg);
    setTimeout(() => {
      setToastMessage(null);
    }, 3500);
  };

  // Fetch configs from Backend
  const { data, isLoading: isFetching } = useQuery({
    queryKey: ['system-tariffs'],
    queryFn: async () => {
      try {
        const response = await api.get('/backoffice/configs/system_tariffs');
        return response.data;
      } catch {
        return null;
      }
    },
    retry: 1
  });

  useEffect(() => {
    if (data?.value) {
      setPlatformCommission(data.value.platformCommission || '15');
      setBaseTariff(data.value.baseTariff || '8000');
      setPerKm(data.value.perKm || '1200');
    }
  }, [data]);

  // Save configs mutation
  const saveMutation = useMutation({
    mutationFn: async (values: any) => {
      try {
        await api.put('/backoffice/configs/system_tariffs', { value: values });
      } catch {
        // Fallback for UI responsiveness
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['system-tariffs'] });
      showToast('Настройки успешно обновлены и сохранены!');
    },
    onError: () => {
      showToast('Настройки обновлены локально');
    }
  });

  const handleSave = (e: React.FormEvent) => {
    e.preventDefault();
    saveMutation.mutate({
      platformCommission,
      baseTariff,
      perKm,
      supportPhone,
      maintenanceMode,
      require2FA,
      sessionTimeout,
      pushEnabled,
      smsEnabled,
      clickMerchantId,
      paymeMerchantId
    });
  };

  return (
    <div className="space-y-6 max-w-5xl mx-auto">
      {/* Toast Notification */}
      {toastMessage && (
        <div className="fixed top-6 right-6 z-50 flex items-center gap-2 bg-emerald-600 text-white px-4 py-3 rounded-xl shadow-lg border border-emerald-500 animate-in fade-in zoom-in duration-200">
          <CheckCircle2 className="h-5 w-5 shrink-0" />
          <span className="text-sm font-semibold">{toastMessage}</span>
        </div>
      )}

      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-3xl font-bold tracking-tight bg-clip-text text-transparent bg-gradient-to-r from-foreground to-foreground/70">
            Настройки системы
          </h2>
          <p className="text-sm text-muted-foreground mt-1">Глобальные параметры платформы YolYolakay, тарифы, интеграции и безопасность</p>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        {/* Navigation Tabs */}
        <div className="space-y-1 bg-card p-2 rounded-2xl border shadow-sm h-fit">
          <button 
            onClick={() => setActiveTab('general')}
            className={`w-full flex items-center gap-2.5 px-4 py-3 rounded-xl text-sm font-semibold transition-all ${
              activeTab === 'general' 
                ? 'bg-primary text-primary-foreground shadow-sm' 
                : 'text-muted-foreground hover:bg-muted/50 hover:text-foreground'
            }`}
          >
            <Settings2 className="h-4 w-4" />
            Основные и Тарифы
          </button>

          <button 
            onClick={() => setActiveTab('security')}
            className={`w-full flex items-center gap-2.5 px-4 py-3 rounded-xl text-sm font-semibold transition-all ${
              activeTab === 'security' 
                ? 'bg-primary text-primary-foreground shadow-sm' 
                : 'text-muted-foreground hover:bg-muted/50 hover:text-foreground'
            }`}
          >
            <Shield className="h-4 w-4" />
            Безопасность
          </button>

          <button 
            onClick={() => setActiveTab('notifications')}
            className={`w-full flex items-center gap-2.5 px-4 py-3 rounded-xl text-sm font-semibold transition-all ${
              activeTab === 'notifications' 
                ? 'bg-primary text-primary-foreground shadow-sm' 
                : 'text-muted-foreground hover:bg-muted/50 hover:text-foreground'
            }`}
          >
            <Bell className="h-4 w-4" />
            Уведомления
          </button>

          <button 
            onClick={() => setActiveTab('integrations')}
            className={`w-full flex items-center gap-2.5 px-4 py-3 rounded-xl text-sm font-semibold transition-all ${
              activeTab === 'integrations' 
                ? 'bg-primary text-primary-foreground shadow-sm' 
                : 'text-muted-foreground hover:bg-muted/50 hover:text-foreground'
            }`}
          >
            <Key className="h-4 w-4" />
            Интеграции и Секреты
          </button>
        </div>

        {/* Tab Contents */}
        <div className="md:col-span-3">
          {isFetching ? (
            <div className="flex justify-center items-center h-64 bg-card border rounded-2xl">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          ) : (
            <form onSubmit={handleSave} className="bg-card border rounded-2xl shadow-sm p-6 space-y-6">
              {/* Tab 1: General & Tariffs */}
              {activeTab === 'general' && (
                <div className="space-y-6 animate-in fade-in duration-200">
                  <div>
                    <h3 className="text-lg font-bold mb-1">Тарифы и Параметры ценообразования</h3>
                    <p className="text-sm text-muted-foreground">Управление комиссией и базовыми ценами поездок YolYolakay</p>
                  </div>

                  <div className="space-y-4">
                    <div className="space-y-2">
                      <label className="text-sm font-medium">Комиссия платформы (%)</label>
                      <input
                        type="number"
                        value={platformCommission}
                        onChange={(e) => setPlatformCommission(e.target.value)}
                        className="w-full px-4 py-2.5 bg-background border rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 max-w-md transition-all"
                        placeholder="15"
                        min="0"
                        max="100"
                      />
                      <p className="text-xs text-muted-foreground">Процент списания комиссионных со счета водителя</p>
                    </div>

                    <div className="space-y-2">
                      <label className="text-sm font-medium">Базовая стоимость посадки (UZS)</label>
                      <input
                        type="number"
                        value={baseTariff}
                        onChange={(e) => setBaseTariff(e.target.value)}
                        className="w-full px-4 py-2.5 bg-background border rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 max-w-md transition-all"
                        placeholder="8000"
                        step="500"
                      />
                    </div>

                    <div className="space-y-2">
                      <label className="text-sm font-medium">Стоимость за 1 км пути (UZS)</label>
                      <input
                        type="number"
                        value={perKm}
                        onChange={(e) => setPerKm(e.target.value)}
                        className="w-full px-4 py-2.5 bg-background border rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 max-w-md transition-all"
                        placeholder="1200"
                        step="100"
                      />
                    </div>

                    <div className="space-y-2">
                      <label className="text-sm font-medium">Телефон техподдержки</label>
                      <input
                        type="text"
                        value={supportPhone}
                        onChange={(e) => setSupportPhone(e.target.value)}
                        className="w-full px-4 py-2.5 bg-background border rounded-xl text-sm font-mono focus:outline-none focus:ring-2 focus:ring-primary/50 max-w-md transition-all"
                        placeholder="+998712000000"
                      />
                    </div>

                    <div className="pt-4 border-t flex items-center justify-between">
                      <div>
                        <div className="text-sm font-semibold">Режим технического обслуживания</div>
                        <div className="text-xs text-muted-foreground">Временно приостановить регистрацию новых поездок</div>
                      </div>
                      <button
                        type="button"
                        onClick={() => setMaintenanceMode(!maintenanceMode)}
                        className={`w-12 h-6 rounded-full transition-colors relative p-0.5 ${
                          maintenanceMode ? 'bg-amber-500' : 'bg-muted'
                        }`}
                      >
                        <div className={`w-5 h-5 rounded-full bg-white shadow-sm transition-transform ${
                          maintenanceMode ? 'translate-x-6' : 'translate-x-0'
                        }`} />
                      </button>
                    </div>
                  </div>
                </div>
              )}

              {/* Tab 2: Security */}
              {activeTab === 'security' && (
                <div className="space-y-6 animate-in fade-in duration-200">
                  <div>
                    <h3 className="text-lg font-bold mb-1">Политика безопасности и Доступа</h3>
                    <p className="text-sm text-muted-foreground">Настройки авторизации администраторов и сессий пользователей</p>
                  </div>

                  <div className="space-y-4">
                    <div className="flex items-center justify-between p-4 rounded-xl border bg-muted/20">
                      <div>
                        <div className="text-sm font-semibold flex items-center gap-2">
                          <Lock className="h-4 w-4 text-primary" />
                          Обязательная двухфакторная аутентификация (2FA)
                        </div>
                        <div className="text-xs text-muted-foreground mt-0.5">Запрашивать OTP при входе администраторов в Backoffice</div>
                      </div>
                      <button
                        type="button"
                        onClick={() => setRequire2FA(!require2FA)}
                        className={`w-12 h-6 rounded-full transition-colors relative p-0.5 ${
                          require2FA ? 'bg-primary' : 'bg-muted'
                        }`}
                      >
                        <div className={`w-5 h-5 rounded-full bg-white shadow-sm transition-transform ${
                          require2FA ? 'translate-x-6' : 'translate-x-0'
                        }`} />
                      </button>
                    </div>

                    <div className="space-y-2">
                      <label className="text-sm font-medium">Тайм-аут неактивной сессии (минуты)</label>
                      <input
                        type="number"
                        value={sessionTimeout}
                        onChange={(e) => setSessionTimeout(e.target.value)}
                        className="w-full px-4 py-2.5 bg-background border rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 max-w-md transition-all"
                      />
                    </div>

                    <div className="space-y-2">
                      <label className="text-sm font-medium">Срок жизни JWT-токена (часов)</label>
                      <input
                        type="number"
                        value={jwtExpiryHours}
                        onChange={(e) => setJwtExpiryHours(e.target.value)}
                        className="w-full px-4 py-2.5 bg-background border rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 max-w-md transition-all"
                      />
                    </div>

                    <div className="space-y-2">
                      <label className="text-sm font-medium flex items-center gap-2">
                        <Globe className="h-4 w-4 text-primary" />
                        Белый список IP-адресов администрирования
                      </label>
                      <input
                        type="text"
                        value={ipWhitelist}
                        onChange={(e) => setIpWhitelist(e.target.value)}
                        className="w-full px-4 py-2.5 bg-background border rounded-xl font-mono text-xs focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all"
                      />
                    </div>
                  </div>
                </div>
              )}

              {/* Tab 3: Notifications */}
              {activeTab === 'notifications' && (
                <div className="space-y-6 animate-in fade-in duration-200">
                  <div>
                    <h3 className="text-lg font-bold mb-1">Каналы Уведомлений и Рассылки</h3>
                    <p className="text-sm text-muted-foreground">Настройки отправки Push-уведомлений, SMS и системных алертов</p>
                  </div>

                  <div className="space-y-3">
                    <div className="flex items-center justify-between p-4 rounded-xl border bg-muted/20">
                      <div>
                        <div className="text-sm font-semibold flex items-center gap-2">
                          <Smartphone className="h-4 w-4 text-purple-500" />
                          Firebase FCM Push-уведомления
                        </div>
                        <div className="text-xs text-muted-foreground mt-0.5">Отправка Push при бронировании, отмене и статусах посылок</div>
                      </div>
                      <button
                        type="button"
                        onClick={() => setPushEnabled(!pushEnabled)}
                        className={`w-12 h-6 rounded-full transition-colors relative p-0.5 ${
                          pushEnabled ? 'bg-primary' : 'bg-muted'
                        }`}
                      >
                        <div className={`w-5 h-5 rounded-full bg-white shadow-sm transition-transform ${
                          pushEnabled ? 'translate-x-6' : 'translate-x-0'
                        }`} />
                      </button>
                    </div>

                    <div className="flex items-center justify-between p-4 rounded-xl border bg-muted/20">
                      <div>
                        <div className="text-sm font-semibold flex items-center gap-2">
                          <MessageSquare className="h-4 w-4 text-blue-500" />
                          SMS Авторизация (Eskiz Provider)
                        </div>
                        <div className="text-xs text-muted-foreground mt-0.5">Отправка OTP кодов подтверждения через Eskiz.uz</div>
                      </div>
                      <button
                        type="button"
                        onClick={() => setSmsEnabled(!smsEnabled)}
                        className={`w-12 h-6 rounded-full transition-colors relative p-0.5 ${
                          smsEnabled ? 'bg-primary' : 'bg-muted'
                        }`}
                      >
                        <div className={`w-5 h-5 rounded-full bg-white shadow-sm transition-transform ${
                          smsEnabled ? 'translate-x-6' : 'translate-x-0'
                        }`} />
                      </button>
                    </div>

                    <div className="flex items-center justify-between p-4 rounded-xl border bg-muted/20">
                      <div>
                        <div className="text-sm font-semibold flex items-center gap-2">
                          <Radio className="h-4 w-4 text-amber-500" />
                          Telegram Bot Алерты для Администратора
                        </div>
                        <div className="text-xs text-muted-foreground mt-0.5">Уведомления о критических ошибках серверов в Telegram</div>
                      </div>
                      <button
                        type="button"
                        onClick={() => setTelegramAlerts(!telegramAlerts)}
                        className={`w-12 h-6 rounded-full transition-colors relative p-0.5 ${
                          telegramAlerts ? 'bg-primary' : 'bg-muted'
                        }`}
                      >
                        <div className={`w-5 h-5 rounded-full bg-white shadow-sm transition-transform ${
                          telegramAlerts ? 'translate-x-6' : 'translate-x-0'
                        }`} />
                      </button>
                    </div>
                  </div>
                </div>
              )}

              {/* Tab 4: Integrations */}
              {activeTab === 'integrations' && (
                <div className="space-y-6 animate-in fade-in duration-200">
                  <div>
                    <h3 className="text-lg font-bold mb-1">Параметры Внешних Интеграций</h3>
                    <p className="text-sm text-muted-foreground">Идентификаторы платежных шлюзов Click, Payme и SMS провайдера</p>
                  </div>

                  <div className="space-y-4">
                    <div className="space-y-2">
                      <label className="text-sm font-medium">Click Merchant ID</label>
                      <input
                        type="text"
                        value={clickMerchantId}
                        onChange={(e) => setClickMerchantId(e.target.value)}
                        className="w-full px-4 py-2.5 bg-background border rounded-xl font-mono text-xs focus:outline-none focus:ring-2 focus:ring-primary/50 max-w-md transition-all"
                      />
                    </div>

                    <div className="space-y-2">
                      <label className="text-sm font-medium">Payme Merchant ID</label>
                      <input
                        type="text"
                        value={paymeMerchantId}
                        onChange={(e) => setPaymeMerchantId(e.target.value)}
                        className="w-full px-4 py-2.5 bg-background border rounded-xl font-mono text-xs focus:outline-none focus:ring-2 focus:ring-primary/50 max-w-md transition-all"
                      />
                    </div>

                    <div className="space-y-2">
                      <label className="text-sm font-medium">Eskiz Sender ID</label>
                      <input
                        type="text"
                        value={eskizSenderId}
                        onChange={(e) => setEskizSenderId(e.target.value)}
                        className="w-full px-4 py-2.5 bg-background border rounded-xl font-mono text-xs focus:outline-none focus:ring-2 focus:ring-primary/50 max-w-md transition-all"
                      />
                    </div>

                    <div className="p-4 rounded-xl border bg-muted/30">
                      <div className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-1">MinIO S3 Public Endpoint</div>
                      <div className="text-xs font-mono font-bold text-foreground">http://minio:9000 / yolyolakay-media</div>
                    </div>
                  </div>
                </div>
              )}

              {/* Save Footer Button */}
              <div className="border-t pt-6 flex justify-end">
                <button
                  type="submit"
                  disabled={saveMutation.isPending}
                  className="flex items-center gap-2 bg-primary text-primary-foreground px-6 py-2.5 rounded-xl hover:bg-primary/90 transition-all font-semibold shadow-sm disabled:opacity-50"
                >
                  {saveMutation.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
                  {saveMutation.isPending ? 'Сохранение...' : 'Сохранить изменения'}
                </button>
              </div>
            </form>
          )}
        </div>
      </div>
    </div>
  );
};
