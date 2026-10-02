import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '../api/axios';
import { MapContainer, TileLayer, Marker, Popup, Polyline, useMap, useMapEvents } from 'react-leaflet';
import 'leaflet/dist/leaflet.css';
import L from 'leaflet';
import { Search, MapPin, Car, Navigation2, Inbox, Loader2, Route, Package } from 'lucide-react';

// Fix for default leaflet markers in react
delete (L.Icon.Default.prototype as any)._getIconUrl;
L.Icon.Default.mergeOptions({
  iconRetinaUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-icon-2x.png',
  iconUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-icon.png',
  shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-shadow.png',
});

// Custom Car Icon
const carIcon = new L.Icon({
  iconUrl: 'https://cdn-icons-png.flaticon.com/512/3204/3204121.png',
  iconSize: [32, 32],
  iconAnchor: [16, 16],
  popupAnchor: [0, -16],
});

const originIcon = new L.Icon({
  iconUrl: 'https://raw.githubusercontent.com/pointhi/leaflet-color-markers/master/img/marker-icon-2x-green.png',
  shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-shadow.png',
  iconSize: [25, 41],
  iconAnchor: [12, 41],
  popupAnchor: [1, -34],
  shadowSize: [41, 41]
});

const destIcon = new L.Icon({
  iconUrl: 'https://raw.githubusercontent.com/pointhi/leaflet-color-markers/master/img/marker-icon-2x-red.png',
  shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-shadow.png',
  iconSize: [25, 41],
  iconAnchor: [12, 41],
  popupAnchor: [1, -34],
  shadowSize: [41, 41]
});

function ChangeView({ center, zoom }: { center: [number, number], zoom: number }) {
  const map = useMap();
  map.setView(center, zoom);
  return null;
}

function MapEvents({ onMapClick }: { onMapClick: (lat: number, lon: number) => void }) {
  useMapEvents({
    click(e) {
      onMapClick(e.latlng.lat, e.latlng.lng);
    },
  });
  return null;
}

const RideRoute = ({ ride, selectedId }: { ride: any, selectedId: string | null }) => {
  const { data: routeData } = useQuery({
    queryKey: ['ride-route', ride.id],
    queryFn: async () => {
      try {
        const res = await api.get('/routing/route', {
          params: {
            origin_lat: ride.origin[0],
            origin_lon: ride.origin[1],
            dest_lat: ride.destination[0],
            dest_lon: ride.destination[1]
          }
        });
        if (res.data && res.data.geometry) {
          return res.data.geometry.map((c: any) => [c.latitude, c.longitude]);
        }
        return [ride.origin, ride.destination]; // Fallback
      } catch {
        return [ride.origin, ride.destination]; // Fallback
      }
    },
    staleTime: Infinity, // Routes between two exact points don't change often
  });

  const positions = routeData || [ride.origin, ride.destination];

  return (
    <>
      <Polyline
        positions={positions}
        color={ride.color}
        weight={selectedId === ride.id ? 7 : 4}
        opacity={selectedId === ride.id ? 1 : 0.8}
      >
        <Popup>
          <div className="p-1">
            <div className="font-bold text-sm mb-1">{ride.type === 'parcel' ? 'Посылка' : 'Поездка'} {ride.id.substring(0,8)}</div>
            <div className="text-xs text-muted-foreground mt-1">Статус: {ride.status}</div>
          </div>
        </Popup>
      </Polyline>
      <Marker position={ride.origin} icon={originIcon}>
         <Popup>Старт ({ride.type === 'parcel' ? 'Посылка' : 'Поездка'} {ride.id.substring(0,4)})</Popup>
      </Marker>
      <Marker position={ride.destination} icon={destIcon}>
         <Popup>Финиш ({ride.type === 'parcel' ? 'Посылка' : 'Поездка'} {ride.id.substring(0,4)})</Popup>
      </Marker>
    </>
  );
};

export const LiveMap = () => {
  const [search, setSearch] = useState('');
  const [mapCenter, setMapCenter] = useState<[number, number]>([41.311081, 69.240562]); // Tashkent
  const [mapZoom, setMapZoom] = useState(13);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'drivers' | 'rides' | 'parcels'>('drivers');
  const [statusFilter, setStatusFilter] = useState('active');

  const [addressQuery, setAddressQuery] = useState('');
  const [currentAddress, setCurrentAddress] = useState('Ташкент, Узбекистан');
  const [clickedPos, setClickedPos] = useState<[number, number] | null>(null);
  const [isGeocoding, setIsGeocoding] = useState(false);

  // Fetch real telematics locations
  const { data: locationsData, isLoading: isLoadingLocs } = useQuery({
    queryKey: ['telematics-locations'],
    queryFn: async () => {
      try {
        const res = await api.get('/telematics/admin/locations');
        return res.data?.locations || [];
      } catch {
        return [];
      }
    },
    refetchInterval: 10000 // Poll every 10 seconds
  });

  // Fetch real active rides
  const { data: activeRidesData, isLoading: isLoadingRides } = useQuery({
    queryKey: ['rides-active', statusFilter],
    queryFn: async () => {
      try {
        const res = await api.get(`/rides/admin/active?status=${statusFilter}`);
        return res.data?.rides || [];
      } catch {
        return [];
      }
    },
    refetchInterval: 15000 // Poll every 15 seconds
  });

  // Fetch real active parcels
  const { data: activeParcelsData, isLoading: isLoadingParcels } = useQuery({
    queryKey: ['parcels-active', statusFilter],
    queryFn: async () => {
      try {
        const res = await api.get(`/parcels/admin/list?status=${statusFilter}`);
        return res.data || [];
      } catch {
        return [];
      }
    },
    refetchInterval: 15000 // Poll every 15 seconds
  });

  // Fetch all users from DB
  const { data: usersData, isLoading: isLoadingUsers } = useQuery({
    queryKey: ['admin-users'],
    queryFn: async () => {
      try {
        const res = await api.get('/users/admin/list');
        return res.data || [];
      } catch {
        return [];
      }
    },
    refetchInterval: 60000 // Poll every 1 min
  });

  // Merge users (role driver) with telematics data
  const dbDrivers = (usersData || []).filter((u: any) => u.role === 'driver');
  const drivers = dbDrivers.map((u: any) => {
    const loc = (locationsData || []).find((l: any) => l.driver_id === u.id);
    return {
      id: u.id,
      name: `${u.first_name || ''} ${u.last_name || ''}`.trim() || `Водитель (ID: ${u.id.substring(0,6)})`,
      status: loc ? 'На линии' : 'Офлайн',
      lat: loc ? loc.latitude : null,
      lng: loc ? loc.longitude : null,
      phone: u.phone_number
    };
  });

  const onlineDrivers = drivers.filter((d: any) => d.lat !== null && d.lng !== null);

  const formatRouteItem = (item: any, isParcel: boolean = false) => {
    let displayStatus = 'В ожидании';
    let color = '#64748b';
    let statusBg = 'bg-slate-500/10 text-slate-600 border-slate-500/20';

    switch(item.status) {
      case 'created':
      case 'draft':
      case 'pending':
        displayStatus = 'Черновик';
        color = '#ef4444'; // Red
        statusBg = 'bg-red-500/10 text-red-600 border-red-500/20';
        break;
      case 'active':
      case 'accepted':
      case 'picked_up':
        displayStatus = 'Активен';
        color = '#eab308'; // Yellow
        statusBg = 'bg-yellow-500/10 text-yellow-700 border-yellow-500/20';
        break;
      case 'in_progress':
        displayStatus = 'В пути';
        color = '#3b82f6'; // Blue
        statusBg = 'bg-blue-500/10 text-blue-600 border-blue-500/20';
        break;
      case 'completed':
      case 'delivered':
        displayStatus = 'Завершен';
        color = '#94a3b8'; // Grey
        statusBg = 'bg-slate-500/10 text-slate-600 border-slate-500/20';
        break;
    }

    return {
      id: item.id,
      driverId: item.driver_id,
      origin: [item.origin_lat, item.origin_lon] as [number, number],
      destination: [item.destination_lat, item.destination_lon] as [number, number],
      status: displayStatus,
      color: color,
      statusBg: statusBg,
      rawStatus: item.status,
      type: isParcel ? 'parcel' : 'ride'
    };
  };

  const activeRides = (activeRidesData || []).map((r: any) => formatRouteItem(r, false)).filter((r: any) => typeof r.origin[0] === 'number' && typeof r.destination[0] === 'number');
  const activeParcels = (activeParcelsData || []).map((p: any) => formatRouteItem(p, true)).filter((r: any) => typeof r.origin[0] === 'number' && typeof r.destination[0] === 'number');

  const filteredDrivers = drivers.filter((d: any) => 
    d.name.toLowerCase().includes(search.toLowerCase()) || 
    d.id.includes(search)
  );

  const filteredRides = activeRides.filter((r: any) => 
    r.id.includes(search) || (r.driverId && r.driverId.includes(search))
  );

  const filteredParcels = activeParcels.filter((r: any) => 
    r.id.includes(search) || (r.driverId && r.driverId.includes(search))
  );

  const focusOnMap = (lat: number, lng: number, id: string) => {
    setMapCenter([lat, lng]);
    setMapZoom(15);
    // If clicking already selected, deselect it
    setSelectedId(prev => prev === id ? null : id);
  };
  
  // Prepare items for map rendering
  const driversToRender = selectedId 
    ? onlineDrivers.filter((d: any) => d.id === selectedId)
    : onlineDrivers;
    
  let routesToRender = activeTab === 'rides' ? activeRides : activeTab === 'parcels' ? activeParcels : [...activeRides, ...activeParcels];
  if (selectedId) {
    routesToRender = routesToRender.filter((r: any) => r.id === selectedId || r.driverId === selectedId);
  }

  const handleMapClick = async (lat: number, lon: number) => {
    setClickedPos([lat, lon]);
    setIsGeocoding(true);
    try {
      const res = await api.get(`/routing/reverse-geocode?lat=${lat}&lon=${lon}`);
      if (res.data && res.data.display_name) {
        setCurrentAddress(res.data.display_name);
        setAddressQuery(res.data.display_name);
      } else {
        setCurrentAddress('Неизвестное место');
      }
    } catch (e) {
      setCurrentAddress('Ошибка геокодирования');
    } finally {
      setIsGeocoding(false);
    }
  };

  const handleAddressSearch = async () => {
    if (!addressQuery.trim()) return;
    setIsGeocoding(true);
    try {
      const res = await api.get(`/routing/geocode?query=${encodeURIComponent(addressQuery)}`);
      if (res.data && res.data.length > 0) {
        const first = res.data[0];
        setMapCenter([first.coordinate.latitude, first.coordinate.longitude]);
        setMapZoom(15);
        setCurrentAddress(first.display_name);
        setClickedPos([first.coordinate.latitude, first.coordinate.longitude]);
      } else {
        alert('Адрес не найден');
      }
    } catch (e) {
      alert('Ошибка при поиске адреса');
    } finally {
      setIsGeocoding(false);
    }
  };

  return (
    <div className="flex h-[calc(100vh-8rem)] gap-4">
      {/* Sidebar */}
      <div className="w-80 flex flex-col bg-card border rounded-xl shadow-sm overflow-hidden shrink-0">
        <div className="p-4 border-b bg-muted/20">
          <div className="flex gap-2 mb-4">
            <button 
              className={`flex-1 py-1.5 px-1 text-xs font-medium rounded-md transition-colors ${activeTab === 'drivers' ? 'bg-primary text-primary-foreground shadow' : 'bg-transparent text-muted-foreground hover:bg-muted'}`}
              onClick={() => setActiveTab('drivers')}
            >
              Водители ({drivers.length})
            </button>
            <button 
              className={`flex-1 py-1.5 px-1 text-xs font-medium rounded-md transition-colors ${activeTab === 'rides' ? 'bg-primary text-primary-foreground shadow' : 'bg-transparent text-muted-foreground hover:bg-muted'}`}
              onClick={() => setActiveTab('rides')}
            >
              Поездки ({activeRides.length})
            </button>
            <button 
              className={`flex-1 py-1.5 px-1 text-xs font-medium rounded-md transition-colors ${activeTab === 'parcels' ? 'bg-primary text-primary-foreground shadow' : 'bg-transparent text-muted-foreground hover:bg-muted'}`}
              onClick={() => setActiveTab('parcels')}
            >
              <Package className="h-4 w-4 mr-2" />
              Посылки
            </button>
          </div>
          
          {(activeTab === 'rides' || activeTab === 'parcels') && (
            <div className="flex px-4 pb-2 items-center text-xs">
              <span className="text-muted-foreground mr-2 font-medium">Статус:</span>
              <select 
                className="bg-transparent border border-input rounded-md px-2 py-1 flex-1 focus:outline-none"
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
              >
                <option value="active">Активные</option>
                <option value="completed">Завершенные</option>
                <option value="canceled">Отмененные</option>
                <option value="">Все статусы</option>
              </select>
            </div>
          )}

          <div className="px-4 pb-4 border-b">
            <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
            <input
              type="text"
              placeholder="Поиск по ID..."
              className="w-full pl-9 pr-4 py-2 bg-background border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </div>
        </div>
        
        <div className="flex-1 overflow-y-auto p-2 space-y-2">
          {activeTab === 'drivers' && (
            isLoadingLocs || isLoadingUsers ? (
              <div className="p-8 text-center text-muted-foreground text-sm">
                <div className="inline-block animate-spin rounded-full h-5 w-5 border-b-2 border-primary mb-2"></div>
                <p>Загрузка водителей...</p>
              </div>
            ) : filteredDrivers.length === 0 ? (
              <div className="p-8 text-center text-muted-foreground">
                <Inbox className="h-8 w-8 mx-auto mb-2 opacity-40" />
                <p className="text-sm font-medium">Водители не найдены</p>
                <p className="text-xs mt-1">В базе нет зарегистрированных водителей</p>
              </div>
            ) : (
              filteredDrivers.map((driver: any) => (
                <div 
                  key={driver.id} 
                  className={`p-3 rounded-lg border transition-all ${
                    driver.status === 'На линии' ? 'cursor-pointer hover:bg-muted/50' : 'opacity-70 grayscale'
                  } ${selectedId === driver.id ? 'bg-primary/10 border-primary/50 shadow-sm' : ''}`}
                  onClick={() => {
                    if (driver.status === 'На линии' && driver.lat && driver.lng) {
                      focusOnMap(driver.lat, driver.lng, driver.id);
                    }
                  }}
                >
                  <div className="flex justify-between items-start mb-1">
                    <span className="font-semibold text-sm">{driver.name}</span>
                    <span className={`text-[10px] px-2 py-0.5 rounded-full font-bold border ${
                      driver.status === 'На линии' ? 'bg-green-500/10 text-green-600 border-green-500/20' : 'bg-slate-500/10 text-slate-600 border-slate-500/20'
                    }`}>
                      {driver.status}
                    </span>
                  </div>
                  <div className="text-xs text-muted-foreground font-mono">
                    {driver.status === 'На линии' ? `Lat: ${driver.lat.toFixed(4)}, Lon: ${driver.lng.toFixed(4)}` : 'Координаты недоступны'}
                  </div>
                </div>
              ))
            )
          )}

          {activeTab === 'rides' && (
            isLoadingRides ? (
              <div className="p-8 text-center text-muted-foreground text-sm">
                <div className="inline-block animate-spin rounded-full h-5 w-5 border-b-2 border-primary mb-2"></div>
                <p>Загрузка поездок...</p>
              </div>
            ) : filteredRides.length === 0 ? (
              <div className="p-8 text-center text-muted-foreground">
                <Route className="h-8 w-8 mx-auto mb-2 opacity-40" />
                <p className="text-sm font-medium">Поездки не найдены</p>
                <p className="text-xs mt-1">В данный момент нет активных поездок</p>
              </div>
            ) : (
              filteredRides.map((ride: any) => (
                <div 
                  key={ride.id} 
                  className={`p-3 rounded-lg border cursor-pointer transition-all ${
                    selectedId === ride.id ? 'bg-primary/10 border-primary/50 shadow-sm' : 'hover:bg-muted/50'
                  }`}
                  onClick={() => focusOnMap(ride.origin[0], ride.origin[1], ride.id)}
                >
                  <div className="flex justify-between items-start mb-1">
                    <span className="font-semibold text-sm">ID: {ride.id.substring(0,8)}</span>
                    <span className={`text-[10px] px-2 py-0.5 rounded-full font-bold border ${ride.statusBg}`}>
                      {ride.status}
                    </span>
                  </div>
                  <div className="text-xs text-muted-foreground font-mono truncate">Driver: {ride.driverId || 'Не назначен'}</div>
                </div>
              ))
            )
          )}

          {activeTab === 'parcels' && (
            isLoadingParcels ? (
              <div className="p-8 text-center text-muted-foreground text-sm">
                <div className="inline-block animate-spin rounded-full h-5 w-5 border-b-2 border-primary mb-2"></div>
                <p>Загрузка посылок...</p>
              </div>
            ) : filteredParcels.length === 0 ? (
              <div className="p-8 text-center text-muted-foreground">
                <Inbox className="h-8 w-8 mx-auto mb-2 opacity-40" />
                <p className="text-sm font-medium">Посылки не найдены</p>
                <p className="text-xs mt-1">В данный момент нет активных посылок</p>
              </div>
            ) : (
              filteredParcels.map((parcel: any) => (
                <div 
                  key={parcel.id} 
                  className={`p-3 rounded-lg border cursor-pointer transition-all ${
                    selectedId === parcel.id ? 'bg-primary/10 border-primary/50 shadow-sm' : 'hover:bg-muted/50'
                  }`}
                  onClick={() => focusOnMap(parcel.origin[0], parcel.origin[1], parcel.id)}
                >
                  <div className="flex justify-between items-start mb-1">
                    <span className="font-semibold text-sm">Посылка: {parcel.id.substring(0,8)}</span>
                    <span className={`text-[10px] px-2 py-0.5 rounded-full font-bold border ${parcel.statusBg}`}>
                      {parcel.status}
                    </span>
                  </div>
                  <div className="text-xs text-muted-foreground font-mono truncate">Driver: {parcel.driverId || 'Не назначен'}</div>
                </div>
              ))
            )
          )}
        </div>
      </div>

      {/* Map Container */}
      <div className="flex-1 rounded-xl border overflow-hidden shadow-sm relative z-0">
        
        {/* Address Search Overlay */}
        <div className="absolute top-4 left-1/2 -translate-x-1/2 z-[400] w-[400px] bg-background/95 backdrop-blur border p-2 rounded-xl shadow-md flex items-center gap-2">
          <input 
            type="text" 
            placeholder="Поиск адреса (например, Ташкент сквер)"
            className="flex-1 bg-transparent border-none focus:outline-none text-sm px-2"
            value={addressQuery}
            onChange={(e) => setAddressQuery(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleAddressSearch()}
          />
          <button 
            onClick={handleAddressSearch}
            className="p-1.5 bg-primary/10 text-primary hover:bg-primary hover:text-white rounded-lg transition-colors"
          >
            {isGeocoding ? <Loader2 className="h-4 w-4 animate-spin" /> : <Search className="h-4 w-4" />}
          </button>
        </div>

        <MapContainer 
          center={mapCenter} 
          zoom={mapZoom} 
          style={{ height: '100%', width: '100%' }}
        >
          <ChangeView center={mapCenter} zoom={mapZoom} />
          <MapEvents onMapClick={handleMapClick} />
          
          <TileLayer
            attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
            url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
          />
          
          {/* Driver Markers (only online) */}
          {driversToRender.map((driver: any) => (
            <Marker 
              key={`driver-${driver.id}`} 
              position={[driver.lat, driver.lng]} 
              icon={carIcon}
            >
              <Popup>
                <div className="p-1">
                  <div className="font-bold text-sm mb-1">{driver.name}</div>
                  <div className="text-xs font-mono">{driver.id}</div>
                  <div className="text-xs text-muted-foreground mt-1">
                    <Navigation2 className="h-3 w-3 inline mr-1 text-primary" />
                    {driver.status}
                  </div>
                </div>
              </Popup>
            </Marker>
          ))}

          {/* Active Rides/Parcels Polylines (Real Routes) */}
          {routesToRender.map((item: any) => (
            <RideRoute key={`route-${item.id}`} ride={item} selectedId={selectedId} />
          ))}

          {clickedPos && (
             <Marker position={clickedPos}>
               <Popup>Выбранное место: {currentAddress}</Popup>
             </Marker>
          )}

        </MapContainer>

        <div className="absolute bottom-4 right-4 z-[400] bg-background/95 backdrop-blur border p-3 rounded-xl shadow-md text-sm flex items-center gap-2 pointer-events-none max-w-xs">
          <MapPin className="h-4 w-4 text-primary shrink-0" />
          <span className="font-semibold text-foreground truncate">{currentAddress}</span>
        </div>
      </div>
    </div>
  );
};
