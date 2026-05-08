import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { useAuthStore } from './auth'

beforeEach(() => {
  sessionStorage.clear()
  useAuthStore.setState({
    user: null,
    accessToken: null,
    refreshToken: null,
    isAuthenticated: false,
    isLoading: false,
    error: null,
  })
})

afterEach(() => {
  vi.restoreAllMocks()
})

describe('useAuthStore', () => {
  describe('init', () => {
    it('restores tokens from sessionStorage', () => {
      sessionStorage.setItem('eva_access_token', 'acc-123')
      sessionStorage.setItem('eva_refresh_token', 'ref-456')

      useAuthStore.getState().init()

      const state = useAuthStore.getState()
      expect(state.accessToken).toBe('acc-123')
      expect(state.refreshToken).toBe('ref-456')
      expect(state.isAuthenticated).toBe(true)
      expect(state.isLoading).toBe(false)
    })

    it('sets isLoading false when no tokens', () => {
      useAuthStore.getState().init()

      const state = useAuthStore.getState()
      expect(state.accessToken).toBeNull()
      expect(state.isAuthenticated).toBe(false)
      expect(state.isLoading).toBe(false)
    })
  })

  describe('setTokens', () => {
    it('stores tokens in sessionStorage and state', () => {
      useAuthStore.getState().setTokens('acc-1', 'ref-1')

      expect(sessionStorage.getItem('eva_access_token')).toBe('acc-1')
      expect(sessionStorage.getItem('eva_refresh_token')).toBe('ref-1')
      expect(useAuthStore.getState().accessToken).toBe('acc-1')
      expect(useAuthStore.getState().refreshToken).toBe('ref-1')
    })
  })

  describe('clearError', () => {
    it('clears the error', () => {
      useAuthStore.setState({ error: 'something went wrong' })
      useAuthStore.getState().clearError()
      expect(useAuthStore.getState().error).toBeNull()
    })
  })

  describe('logout', () => {
    it('clears tokens, user, and redirects to /', () => {
      const originalLocation = window.location
      const mockLocation = { href: '' } as Location
      vi.stubGlobal('location', mockLocation)

      sessionStorage.setItem('eva_access_token', 'acc')
      sessionStorage.setItem('eva_refresh_token', 'ref')
      useAuthStore.setState({
        user: { id: 1, discord_id: 'test', discord_username: '', display_name: '', is_admin: false, created_at: '2024-01-01' },
        accessToken: 'acc',
        refreshToken: 'ref',
        isAuthenticated: true,
      })

      useAuthStore.getState().logout()

      expect(sessionStorage.getItem('eva_access_token')).toBeNull()
      expect(sessionStorage.getItem('eva_refresh_token')).toBeNull()
      expect(useAuthStore.getState().user).toBeNull()
      expect(useAuthStore.getState().isAuthenticated).toBe(false)
      expect(mockLocation.href).toBe('/')

      vi.stubGlobal('location', originalLocation)
    })
  })

  describe('getAccessToken / getRefreshToken', () => {
    it('returns current tokens', () => {
      useAuthStore.setState({ accessToken: 'acc-1', refreshToken: 'ref-1' })
      expect(useAuthStore.getState().getAccessToken()).toBe('acc-1')
      expect(useAuthStore.getState().getRefreshToken()).toBe('ref-1')
    })
  })
})
