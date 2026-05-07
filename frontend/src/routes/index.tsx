import { Routes, Route } from 'react-router-dom'
import Home from '../pages/Home'
import Editor from '../pages/Editor'
import NotFound from '../pages/NotFound'

export function AppRoutes() {
  return (
    <Routes>
      <Route path="/" element={<Home />} />
      <Route path="/editor" element={<Editor />} />
      <Route path="*" element={<NotFound />} />
    </Routes>
  )
}
