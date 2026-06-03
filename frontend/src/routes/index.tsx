import { lazy, Suspense } from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import AppLayout from '../components/AppLayout'
import AdminGuard from '../components/AdminGuard'
import ArchiveGuard from '../components/ArchiveGuard'
import Login from '../pages/Login'
import Callback from '../pages/Callback'
import TeamSettings from '../pages/TeamSettings'
import TeamsList from '../pages/TeamsList'
import CreateTeam from '../pages/CreateTeam'
import History from '../pages/History'
import ZoneList from '../pages/ZoneList'
import InviteAccept from '../pages/InviteAccept'
import ProfileSettings from '../pages/ProfileSettings'
import ArchivePage from '../pages/ArchivePage'
import NotFound from '../pages/NotFound'

const Editor = lazy(() => import('../pages/Editor'))
const Revision = lazy(() => import('../pages/Revision'))
const AdminDashboard = lazy(() => import('../pages/AdminDashboard'))

export function AppRoutes() {
  return (
    <Suspense fallback={<div className="spinner" />}>
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
        <Route path="/archives" element={<ArchiveGuard><ArchivePage /></ArchiveGuard>} />
        <Route path="/teams/:id/editor/:mapId" element={<Editor />} />
        <Route path="/teams/:id/maps/:mapId/history" element={<History />} />
        <Route path="/teams/:id/revision" element={<Revision />} />
        <Route path="/teams/:id/revision/:mapId" element={<Revision />} />
        <Route path="/teams/:id/zones/:mapId" element={<ZoneList />} />
      </Route>
      <Route path="/invite/:token" element={<InviteAccept />} />
      <Route path="*" element={<NotFound />} />
    </Routes>
    </Suspense>
  )
}
