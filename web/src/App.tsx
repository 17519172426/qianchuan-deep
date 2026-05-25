import { Routes, Route, Navigate } from 'react-router-dom'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import Ads from './pages/Ads'
import Rules from './pages/Rules'
import Recommendations from './pages/Recommendations'
import Creatives from './pages/Creatives'
import Reports from './pages/Reports'
import Accounts from './pages/Accounts'
import OAuthCallback from './pages/OAuthCallback'
import Layout from './components/Layout'

function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/oauth/callback" element={<OAuthCallback />} />
      <Route element={<Layout />}>
        <Route path="/" element={<Navigate to="/dashboard" replace />} />
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/ads" element={<Ads />} />
        <Route path="/rules" element={<Rules />} />
        <Route path="/recommendations" element={<Recommendations />} />
        <Route path="/creatives" element={<Creatives />} />
        <Route path="/reports" element={<Reports />} />
        <Route path="/accounts" element={<Accounts />} />
      </Route>
    </Routes>
  )
}

export default App
