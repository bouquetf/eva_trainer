import { Routes, Route } from 'react-router-dom'
import Home from '../pages/Home'
import Login from '../pages/Login'
import Callback from '../pages/Callback'
import Editor from '../pages/Editor'
import NotFound from '../pages/NotFound'

export function AppRoutes() {
  return (
    <Routes>
      <Route path="/" element={<Home />} />
      <Route path="/login" element={<Login />} />
      <Route path="/callback" element={<Callback />} />
      <Route path="/editor" element={<Editor />} />
      <Route path="*" element={<NotFound />} />
    </Routes>
  )
}
