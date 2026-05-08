interface ApiConfig {
  baseURL: string
  getAccessToken: () => string | null
  getRefreshToken: () => string | null
  onRefresh: (access: string, refresh: string) => void
  onAuthError: () => void
}

class ApiClient {
  private config: ApiConfig
  private refreshing: Promise<boolean> | null = null

  constructor(config: ApiConfig) {
    this.config = config
  }

  async get<T>(path: string): Promise<T> {
    return this.request<T>(path, { method: 'GET' })
  }

  async post<T>(path: string, body?: unknown): Promise<T> {
    return this.request<T>(path, {
      method: 'POST',
      headers: body ? { 'Content-Type': 'application/json' } : undefined,
      body: body ? JSON.stringify(body) : undefined,
    })
  }

  async put<T>(path: string, body?: unknown): Promise<T> {
    return this.request<T>(path, {
      method: 'PUT',
      headers: body ? { 'Content-Type': 'application/json' } : undefined,
      body: body ? JSON.stringify(body) : undefined,
    })
  }

  async del<T>(path: string, body?: unknown): Promise<T> {
    return this.request<T>(path, {
      method: 'DELETE',
      headers: body ? { 'Content-Type': 'application/json' } : undefined,
      body: body ? JSON.stringify(body) : undefined,
    })
  }

  async upload<T>(path: string, formData: FormData): Promise<T> {
    const token = this.config.getAccessToken()
    const headers: Record<string, string> = {}
    if (token) {
      headers['Authorization'] = `Bearer ${token}`
    }

    const res = await fetch(`${this.config.baseURL}${path}`, {
      method: 'POST',
      headers,
      body: formData,
    })

    if (res.status === 401) {
      const refreshed = await this.tryRefresh()
      if (refreshed) {
        const newToken = this.config.getAccessToken()
        headers['Authorization'] = `Bearer ${newToken}`
        const retryRes = await fetch(`${this.config.baseURL}${path}`, {
          method: 'POST',
          headers,
          body: formData,
        })
        if (!retryRes.ok) {
          throw new ApiError(retryRes.status, await retryRes.json().catch(() => ({})))
        }
        return retryRes.json()
      }

      this.config.onAuthError()
      throw new ApiError(401, { error: 'unauthorized', message: 'Session expirée' })
    }

    if (!res.ok) {
      throw new ApiError(res.status, await res.json().catch(() => ({})))
    }

    return res.json()
  }

  private async request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const token = this.config.getAccessToken()
    const headers: Record<string, string> = {
      ...(options.headers as Record<string, string>),
    }

    if (token) {
      headers['Authorization'] = `Bearer ${token}`
    }

    const res = await fetch(`${this.config.baseURL}${path}`, {
      ...options,
      headers,
    })

    if (res.status === 401) {
      const refreshed = await this.tryRefresh()
      if (refreshed) {
        const newToken = this.config.getAccessToken()
        headers['Authorization'] = `Bearer ${newToken}`
        const retryRes = await fetch(`${this.config.baseURL}${path}`, {
          ...options,
          headers,
        })
        if (!retryRes.ok) {
          throw new ApiError(retryRes.status, await retryRes.json().catch(() => ({})))
        }
        return retryRes.json()
      }

      this.config.onAuthError()
      throw new ApiError(401, { error: 'unauthorized', message: 'Session expirée' })
    }

    if (!res.ok) {
      throw new ApiError(res.status, await res.json().catch(() => ({})))
    }

    return res.json()
  }

  private async tryRefresh(): Promise<boolean> {
    if (this.refreshing) {
      return this.refreshing
    }

    this.refreshing = this.doRefresh()
    try {
      return await this.refreshing
    } finally {
      this.refreshing = null
    }
  }

  private async doRefresh(): Promise<boolean> {
    const refreshToken = this.config.getRefreshToken()
    if (!refreshToken) {
      return false
    }

    try {
      const res = await fetch(`${this.config.baseURL}/api/v1/auth/refresh`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: refreshToken }),
      })

      if (!res.ok) {
        return false
      }

      const data = await res.json()
      this.config.onRefresh(data.access_token, data.refresh_token)
      return true
    } catch {
      return false
    }
  }
}

class ApiError extends Error {
  status: number
  body: Record<string, unknown>

  constructor(status: number, body: Record<string, unknown>) {
    super(String(body.message ?? body.error ?? 'API Error'))
    this.status = status
    this.body = body
  }
}

type StoreAccessor = () => {
  getAccessToken: () => string | null
  getRefreshToken: () => string | null
  setTokens: (access: string, refresh: string) => void
  logout: () => void
}

let apiInstance: ApiClient | null = null

export function __initClient(accessor: StoreAccessor): void {
  if (!apiInstance) {
    apiInstance = new ApiClient({
      baseURL: import.meta.env.VITE_API_URL ?? 'http://localhost:8080',
      getAccessToken: () => accessor().getAccessToken(),
      getRefreshToken: () => accessor().getRefreshToken(),
      onRefresh: (access, refresh) => accessor().setTokens(access, refresh),
      onAuthError: () => accessor().logout(),
    })
  }
}

export function getApiClient(): ApiClient {
  if (!apiInstance) {
    throw new Error('ApiClient not initialized. Call __initClient first.')
  }
  return apiInstance
}

export { ApiClient, ApiError }
