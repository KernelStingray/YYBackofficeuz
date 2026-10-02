import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AppLayout } from './components/layout/AppLayout';
import { ProtectedRoute } from './components/layout/ProtectedRoute';
import { Login } from './pages/Login';
import { Dashboard } from './pages/Dashboard';
import { Users } from './pages/Users';
import { Billing } from './pages/Billing';
import { Infrastructure } from './pages/Infrastructure';
import { LiveMap } from './pages/LiveMap';
import { Parcels } from './pages/Parcels';
import { Settings } from './pages/Settings';

const queryClient = new QueryClient();

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<Login />} />
          
          <Route element={<ProtectedRoute />}>
            <Route element={<AppLayout />}>
              <Route path="/" element={<Dashboard />} />
              <Route path="/users" element={<Users />} />
              <Route path="/billing" element={<Billing />} />
              <Route path="/infrastructure" element={<Infrastructure />} />
              {/* Other routes will be added here in subsequent stages */}
              <Route path="/map" element={<LiveMap />} />
              <Route path="/parcels" element={<Parcels />} />
              <Route path="/settings" element={<Settings />} />
            </Route>
          </Route>
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
}

export default App;
