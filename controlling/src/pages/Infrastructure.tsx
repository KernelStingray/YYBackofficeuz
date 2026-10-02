import { useState, useEffect, useRef } from 'react';
import { Server, Database, Activity, TerminalSquare, RotateCw, Cable, Layers, CheckCircle2, ShieldAlert } from 'lucide-react';
import { useAuthStore } from '../store/useAuthStore';
import { getApiUrl } from '../api/axios';

interface SystemNode {
  id: string;
  name: string;
  group: 'Saturn' | 'Jupiter' | 'External';
  type: 'service' | 'database' | 'queue' | 'integration';
  status: 'healthy' | 'warning';
}

const systemNodes: SystemNode[] = [
  // Group Saturn
  { id: 'api_gateway', name: 'API Gateway', group: 'Saturn', type: 'service', status: 'healthy' },
  { id: 'auth_service', name: 'Auth Service', group: 'Saturn', type: 'service', status: 'healthy' },
  { id: 'user_service', name: 'User Service', group: 'Saturn', type: 'service', status: 'healthy' },
  { id: 'ride_service', name: 'Ride Service', group: 'Saturn', type: 'service', status: 'healthy' },
  { id: 'routing_service', name: 'Routing Service', group: 'Saturn', type: 'service', status: 'healthy' },
  { id: 'telematics_service', name: 'Telematics Service', group: 'Saturn', type: 'service', status: 'healthy' },
  { id: 'billing_service', name: 'Billing Service', group: 'Saturn', type: 'service', status: 'healthy' },
  { id: 'parcel_service', name: 'Parcel Service', group: 'Saturn', type: 'service', status: 'healthy' },
  { id: 'notification_service', name: 'Notification Service', group: 'Saturn', type: 'service', status: 'healthy' },
  { id: 'redis', name: 'Redis Cache', group: 'Saturn', type: 'database', status: 'healthy' },
  { id: 'rabbitmq', name: 'RabbitMQ Broker', group: 'Saturn', type: 'queue', status: 'healthy' },

  // Group Jupiter
  { id: 'postgres_jupiter', name: 'PostgreSQL Main DB', group: 'Jupiter', type: 'database', status: 'healthy' },

  // Integrations & External
  { id: 'payme_gateway', name: 'Payme Integration', group: 'External', type: 'integration', status: 'healthy' },
  { id: 'click_gateway', name: 'Click Integration', group: 'External', type: 'integration', status: 'healthy' },
  { id: 'sms_provider', name: 'Eskiz SMS Provider', group: 'External', type: 'integration', status: 'healthy' },
];

const sampleLogTemplates: Record<string, string[]> = {
  api_gateway: [
    'GET /api/v1/rides/search 200 OK - 14ms',
    'POST /api/v1/auth/login 200 OK - 42ms',
    'GET /api/v1/users/admin/list 200 OK - 18ms',
    'POST /api/v1/helpdesk/tickets 201 Created - 25ms',
  ],
  auth_service: [
    'JWT Token generated for user_id=d84aa1eb-65a6...',
    'Validated Bearer token authorization header',
    'Password hash verified successfully',
  ],
  user_service: [
    'Fetched user profile by ID d84aa1eb-65a6...',
    'User MyID verification status checked: true',
  ],
  ride_service: [
    'Querying active intercity rides: Tashkent -> Samarkand',
    'Driver status updated: ON_THE_WAY',
    'New ride booking created: booking_id=b771a82...',
  ],
  postgres_jupiter: [
    'Executed query: SELECT * FROM users WHERE phone_number = $1 (2.1ms)',
    'Executed query: UPDATE user_subscriptions SET plan_name = $1 (1.4ms)',
  ],
  redis: [
    'CACHE HIT: key=ride_search_cache_tashkent_samarkand',
    'EXPIRE key=session_user_token 3600s',
  ],
  rabbitmq: [
    'Published message to exchange=notifications.exchange, routing_key=push.send',
    'Consumed message from queue=ride.events',
  ],
  notification_service: [
    'FCM Push notification sent successfully to token=f8c12...',
    'SMS payload dispatched via Eskiz API',
  ],
  billing_service: [
    'Calculated 3% platform commission for transaction trx-9921',
    'Verified user subscription status: ACTIVE (plan=premium)',
  ],
  parcel_service: [
    'Parcel status updated: IN_TRANSIT (parcel_id=pkg-1002)',
    'Fetched active parcel deliveries count: 24',
  ],
  routing_service: [
    'Calculating optimal route from lat,lng to lat,lng using OSRM engine',
    'Found route with ETA: 45 mins, distance: 32 km',
    'Route cached in Redis',
  ],
  telematics_service: [
    'Received live coordinates update from driver ID 992-abc',
    'Updating driver location in Redis geospatial index',
    'Broadcasting driver location to connected riders via WebSocket',
  ],
};

export const Infrastructure = () => {
  const [selectedGroup, setSelectedGroup] = useState<'All' | 'Saturn' | 'Jupiter' | 'External'>('All');
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const [logs, setLogs] = useState<Array<{ timestamp: string; node: string; message: string }>>([]);
  const [isStreaming, setIsStreaming] = useState(true);
  const logsEndRef = useRef<HTMLDivElement>(null);

  // Auto-scroll terminal
  useEffect(() => {
    if (logsEndRef.current) {
      logsEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [logs]);

  // Initial seed logs
  useEffect(() => {
    const initialLogs = [
      { timestamp: new Date(Date.now() - 10000).toLocaleTimeString(), node: 'api_gateway', message: 'API Gateway started on port 8080' },
      { timestamp: new Date(Date.now() - 8000).toLocaleTimeString(), node: 'postgres_jupiter', message: 'Database connection pool established (max=20)' },
      { timestamp: new Date(Date.now() - 6000).toLocaleTimeString(), node: 'rabbitmq', message: 'RabbitMQ broker connected successfully' },
      { timestamp: new Date(Date.now() - 4000).toLocaleTimeString(), node: 'auth_service', message: 'RSA public key loaded for JWT validation' },
    ];
    setLogs(initialLogs);
  }, []);

  // Live log generator stream
  useEffect(() => {
    if (!isStreaming) return;

    const token = useAuthStore.getState().token;
    const API_URL = getApiUrl();
    
    // Connect to real SSE endpoint on backend
    const eventSource = new EventSource(`${API_URL}/backoffice/logs/stream?token=${token}`);

    eventSource.addEventListener('connected', (event) => {
      console.log('Log stream connected:', event.data);
    });

    eventSource.addEventListener('log', (event) => {
      let nodeName = 'system';
      let msg = event.data;
      
      // Parse the format "[service_name] log message" sent by Go backend
      const match = event.data.match(/^\[(.*?)\]\s+(.*)/);
      if (match) {
        // Loki containers often have compose project names, strip yolyolakay- prefix if any
        nodeName = match[1].replace(/^yolyolakay-/, '').replace(/^yybackend-/, '').replace(/-1$/, '');
        msg = match[2];
      }

      // Mask long JWT tokens to prevent log UI from bloating
      msg = msg.replace(/(token=|Bearer\s+)(eyJ[a-zA-Z0-9-_.]*)/g, '$1***');

      const newLog = {
        timestamp: new Date().toLocaleTimeString(),
        node: nodeName,
        message: msg,
      };

      setLogs((prev) => [...prev.slice(-400), newLog]);
    });

    eventSource.onerror = (err) => {
      console.error('SSE Error:', err);
      // EventSource tries to reconnect automatically
    };

    return () => {
      eventSource.close();
    };
  }, [isStreaming]);

  // Filter nodes list by group
  const displayNodes = selectedGroup === 'All' 
    ? systemNodes 
    : systemNodes.filter(n => n.group === selectedGroup);

  // Filter logs by selected node (if any)
  const displayLogs = selectedNodeId 
    ? logs.filter(l => l.node === selectedNodeId)
    : logs;

  const selectedNodeObj = systemNodes.find(n => n.id === selectedNodeId);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h2 className="text-3xl font-bold tracking-tight bg-clip-text text-transparent bg-gradient-to-r from-foreground to-foreground/70">
            Инфраструктура и Логи
          </h2>
          <p className="text-sm text-muted-foreground mt-1">Мониторинг кластеров Group Jupiter, Group Saturn и живая трансляция системных событий</p>
        </div>
        <div className="flex items-center gap-3">
          <button 
            onClick={() => setIsStreaming(!isStreaming)}
            className={`flex items-center gap-2 px-4 py-2 rounded-xl text-xs font-bold border transition-all ${
              isStreaming 
                ? 'bg-green-500/10 text-green-600 dark:text-green-400 border-green-500/20 shadow-sm' 
                : 'bg-amber-500/10 text-amber-600 border-amber-500/20'
            }`}
          >
            <Activity className={`h-4 w-4 ${isStreaming ? 'animate-pulse text-green-500' : ''}`} />
            {isStreaming ? 'Стриминг логов (ВКЛ)' : 'Стриминг (НА ПАУЗЕ)'}
          </button>
        </div>
      </div>

      {/* Cluster Overview Cards */}
      <div className="grid gap-4 md:grid-cols-3">
        <div className="rounded-2xl border bg-card p-5 shadow-sm transition-all hover:shadow-md">
          <div className="flex justify-between items-center mb-2">
            <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Кластер Group Saturn</span>
            <Server className="h-4 w-4 text-primary" />
          </div>
          <div className="text-2xl font-extrabold">11 Микросервисов</div>
          <p className="text-xs text-green-600 dark:text-green-400 mt-1 font-medium flex items-center gap-1">
            <CheckCircle2 className="h-3.5 w-3.5" />
            Все 11 узлов активны
          </p>
        </div>

        <div className="rounded-2xl border bg-card p-5 shadow-sm transition-all hover:shadow-md">
          <div className="flex justify-between items-center mb-2">
            <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Кластер Group Jupiter</span>
            <Database className="h-4 w-4 text-purple-500" />
          </div>
          <div className="text-2xl font-extrabold">PostgreSQL Main DB</div>
          <p className="text-xs text-green-600 dark:text-green-400 mt-1 font-medium flex items-center gap-1">
            <CheckCircle2 className="h-3.5 w-3.5" />
            Соединение установлено
          </p>
        </div>

        <div className="rounded-2xl border bg-card p-5 shadow-sm transition-all hover:shadow-md">
          <div className="flex justify-between items-center mb-2">
            <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Внешние Интеграции</span>
            <Cable className="h-4 w-4 text-amber-500" />
          </div>
          <div className="text-2xl font-extrabold">Payme, Click, Eskiz</div>
          <p className="text-xs text-green-600 dark:text-green-400 mt-1 font-medium flex items-center gap-1">
            <CheckCircle2 className="h-3.5 w-3.5" />
            Все шлюзы доступны
          </p>
        </div>
      </div>

      <div className="grid gap-6 md:grid-cols-3">
        {/* Components List Panel */}
        <div className="rounded-2xl border bg-card p-5 shadow-sm space-y-4 flex flex-col">
          <div className="flex items-center justify-between">
            <h3 className="font-bold text-base flex items-center gap-2">
              <Layers className="h-4 w-4 text-primary" />
              Компоненты ({displayNodes.length})
            </h3>
          </div>

          {/* Group Filter Tabs */}
          <div className="flex p-1 bg-muted/40 rounded-xl gap-1">
            {(['All', 'Saturn', 'Jupiter', 'External'] as const).map((group) => (
              <button
                key={group}
                onClick={() => setSelectedGroup(group)}
                className={`flex-1 py-1.5 text-xs font-semibold rounded-lg transition-all ${
                  selectedGroup === group 
                    ? 'bg-card text-foreground shadow-sm' 
                    : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                {group === 'All' ? 'Все' : group}
              </button>
            ))}
          </div>

          {/* Nodes List */}
          <div className="space-y-2 flex-1 max-h-[460px] overflow-y-auto pr-1">
            {displayNodes.map(node => {
              const isSelected = selectedNodeId === node.id;
              return (
                <div 
                  key={node.id} 
                  onClick={() => setSelectedNodeId(isSelected ? null : node.id)}
                  className={`p-3 rounded-xl border cursor-pointer transition-all ${
                    isSelected 
                      ? 'bg-primary/10 border-primary shadow-sm font-semibold' 
                      : 'hover:bg-muted/40 border-border/50'
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <span className="text-sm">{node.name}</span>
                    <span className="text-[10px] px-2 py-0.5 rounded-full font-bold bg-green-500/10 text-green-600 dark:text-green-400 border border-green-500/20">
                      Online
                    </span>
                  </div>
                  <div className="text-xs text-muted-foreground mt-1 flex items-center justify-between">
                    <span>Группа: {node.group}</span>
                    {isSelected && <span className="text-[10px] text-primary font-bold">● Выбран в фильтре</span>}
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* Live Terminal Log Stream */}
        <div className="md:col-span-2 rounded-2xl border bg-zinc-950 text-zinc-100 shadow-md flex flex-col h-[525px] overflow-hidden font-mono">
          <div className="px-4 py-3 bg-zinc-900 border-b border-zinc-800 flex items-center justify-between">
            <div className="flex items-center gap-2 text-xs font-semibold text-zinc-300">
              <TerminalSquare className="h-4 w-4 text-green-400" />
              <span>
                {selectedNodeObj 
                  ? `Трансляция логов: ${selectedNodeObj.name}` 
                  : 'Трансляция логов (Все микросервисы)'}
              </span>
            </div>
            <div className="flex items-center gap-3">
              {selectedNodeId && (
                <button 
                  onClick={() => setSelectedNodeId(null)}
                  className="text-xs text-primary hover:underline font-semibold"
                >
                  Сбросить фильтр
                </button>
              )}
              <button 
                onClick={() => setLogs([])}
                className="text-xs text-zinc-400 hover:text-zinc-200 transition-colors"
              >
                Очистить экран
              </button>
            </div>
          </div>
          
          <div className="flex-1 p-4 overflow-y-auto space-y-1.5 text-xs">
            {displayLogs.length === 0 ? (
              <div className="text-zinc-500 italic py-8 text-center">
                {selectedNodeId ? `Нет поступающих логов для сервиса ${selectedNodeObj?.name}` : 'Ожидание поступающих логов от сервисов...'}
              </div>
            ) : (
              displayLogs.map((log, index) => (
                <div key={index} className="leading-relaxed break-all hover:bg-zinc-900/60 p-1 rounded transition-colors flex items-start gap-2">
                  <span className="text-zinc-500 shrink-0 font-mono text-[11px]">[{log.timestamp}]</span>
                  <span className="px-1.5 py-0.5 rounded bg-zinc-800 text-primary-foreground text-[10px] font-bold shrink-0">
                    {log.node}
                  </span>
                  <span className={log.message.includes('ERROR') ? 'text-red-400 font-semibold' : log.message.includes('WARN') ? 'text-amber-400' : 'text-zinc-300'}>
                    {log.message}
                  </span>
                </div>
              ))
            )}
            <div ref={logsEndRef} />
          </div>
        </div>
      </div>
    </div>
  );
};
