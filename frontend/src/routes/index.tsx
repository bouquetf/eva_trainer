import { Routes, Route, Navigate } from 'react-router-dom'
import Login from '../pages/Login'
import Callback from '../pages/Callback'
import Editor from '../pages/Editor'
import History from '../pages/History'
import Revision from '../pages/Revision'
import NotFound from '../pages/NotFound'
import AppLayout from '../components/AppLayout'
import AdminGuard from '../components/AdminGuard'
import TeamsList from '../pages/TeamsList'
import CreateTeam from '../pages/CreateTeam'
import TeamSettings from '../pages/TeamSettings'
import ZoneList from '../pages/ZoneList'
import InviteAccept from '../pages/InviteAccept'
import ProfileSettings from '../pages/ProfileSettings'
import AdminDashboard from '../pages/AdminDashboard'

export function AppRoutes() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/callback" element={<Callback />} />
      <Route element={<AppLayout />}>
        <Route path="/" element={<Navigate to="/teams" replace />} />
        <Route path="/settings" element={<ProfileSettings />} />
        <Route path="/teams" element={<TeamsList />} />
        <Route path="/teams/create" element={<CreateTeam />} />
        <Route path="/teams/:id" element={<TeamSettings />} />
        <Route path="/admin" element={<AdminGuard><AdminDashboard /></AdminGuard>} />
        <Route path="/teams/:id/editor/:mapId" element={<Editor />} />
        <Route path="/teams/:id/maps/:mapId/history" element={<History />} />
        <Route path="/teams/:id/revision" element={<Revision />} />
        <Route path="/teams/:id/revision/:mapId" element={<Revision />} />
        <Route path="/teams/:id/zones/:mapId" element={<ZoneList />} />
      </Route>
      <Route path="/invite/:token" element={<InviteAccept />} />
      <Route path="*" element={<NotFound />} />
    </Routes>
  )
}
