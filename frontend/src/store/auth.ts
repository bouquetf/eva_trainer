import { create } from 'zustand'
import { __initClient, getApiClient } from '../api/client'
import { fetchMe, getLoginUrl } from '../api/auth'
import type { User } from '../api/auth'

const STORAGE_KEY_ACCESS = 'eva_access_token'
const STORAGE_KEY_REFRESH = 'eva_refresh_token'

interface AuthState {
  user: User | null
  accessToken: string | null
  refreshToken: string | null
  isAuthenticated: boolean
  isLoading: boolean
  error: string | null

  init: () => void
  login: () => void
  handleCallback: () => void
  refreshAuth: () => Promise<boolean>
  logout: () => void
  setTokens: (access: string, refresh: string) => void
  getAccessToken: () => string | null
  getRefreshToken: () => string | null
  clearError: () => void
}

export const useAuthStore = create<AuthState>()((set, get) => {
  __initClient(() => ({
    getAccessToken: () => get().accessToken,
    getRefreshToken: () => get().refreshToken,
    setTokens: (access, refresh) => {
      sessionStorage.setItem(STORAGE_KEY_ACCESS, access)
      sessionStorage.setItem(STORAGE_KEY_REFRESH, refresh)
      set({ accessToken: access, refreshToken: refresh })
    },
    logout: () => {
      sessionStorage.removeItem(STORAGE_KEY_ACCESS)
      sessionStorage.removeItem(STORAGE_KEY_REFRESH)
      set({
        user: null,
        accessToken: null,
        refreshToken: null,
        isAuthenticated: false,
        isLoading: false,
        error: null,
      })
      window.location.href = '/'
    },
  }))

  return {
    user: null,
    accessToken: null,
    refreshToken: null,
    isAuthenticated: false,
    isLoading: true,
    error: null,

    init: () => {
      const accessToken = sessionStorage.getItem(STORAGE_KEY_ACCESS)
      const refreshToken = sessionStorage.getItem(STORAGE_KEY_REFRESH)

      if (accessToken && refreshToken) {
        set({ accessToken, refreshToken, isAuthenticated: true, isLoading: false })
      } else {
        set({ isLoading: false })
      }
    },

    login: () => {
      window.location.href = getLoginUrl()
    },

    handleCallback: () => {
      const hash = window.location.hash.replace(/^#/, '')
      if (!hash) {
        set({ error: 'Aucun token reçu', isLoading: false })
        return
      }

      const params = new URLSearchParams(hash)
      const accessToken = params.get('access_token')
      const refreshToken = params.get('refresh_token')
      const expiresIn = params.get('expires_in')
      const redirect = params.get('redirect') ?? '/'

      if (!accessToken || !refreshToken) {
        set({ error: 'Tokens manquants dans la réponse', isLoading: false })
        return
      }

      sessionStorage.setItem(STORAGE_KEY_ACCESS, accessToken)
      sessionStorage.setItem(STORAGE_KEY_REFRESH, refreshToken)

      set({
        accessToken,
        refreshToken,
        isAuthenticated: true,
        isLoading: true,
        error: null,
      })

      if (expiresIn) {
        const ms = parseInt(expiresIn, 10) * 1000
        if (ms > 0) {
          setTimeout(() => {
            get().refreshAuth()
          }, ms - 60000)
        }
      }

      fetchMe()
        .then((user) => {
          set({ user, isLoading: false })
          window.location.href = redirect
        })
        .catch(() => {
          sessionStorage.removeItem(STORAGE_KEY_ACCESS)
          sessionStorage.removeItem(STORAGE_KEY_REFRESH)
          set({
            user: null,
            accessToken: null,
            refreshToken: null,
            isAuthenticated: false,
            isLoading: false,
            error: 'Impossible de récupérer vos informations',
          })
        })
    },

    refreshAuth: async () => {
      const refreshToken = get().refreshToken
      if (!refreshToken) {
        get().logout()
        return false
      }

      try {
        const api = getApiClient()
        const data = await api.post<{ access_token: string; refresh_token: string; expires_in: number }>(
          '/api/v1/auth/refresh',
          { refresh_token: refreshToken }
        )

        sessionStorage.setItem(STORAGE_KEY_ACCESS, data.access_token)
        sessionStorage.setItem(STORAGE_KEY_REFRESH, data.refresh_token)

        set({
          accessToken: data.access_token,
          refreshToken: data.refresh_token,
        })

        return true
      } catch {
        get().logout()
        return false
      }
    },

    logout: () => {
      sessionStorage.removeItem(STORAGE_KEY_ACCESS)
      sessionStorage.removeItem(STORAGE_KEY_REFRESH)
      set({
        user: null,
        accessToken: null,
        refreshToken: null,
        isAuthenticated: false,
        isLoading: false,
        error: null,
      })
      window.location.href = '/'
    },

    setTokens: (access: string, refresh: string) => {
      sessionStorage.setItem(STORAGE_KEY_ACCESS, access)
      sessionStorage.setItem(STORAGE_KEY_REFRESH, refresh)
      set({ accessToken: access, refreshToken: refresh })
    },

    getAccessToken: () => get().accessToken,
    getRefreshToken: () => get().refreshToken,

    clearError: () => set({ error: null }),
  }
})
