import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useTeamRole } from './useTeamRole';

const mockFetch = vi.fn();
vi.stubGlobal('fetch', mockFetch);

vi.mock('../api/client', () => {
  class MockApiClient {
    private baseURL = 'http://localhost:8080';

    private async request<T>(path: string, options: RequestInit = {}): Promise<T> {
      const token = 'test-token';
      const headers: Record<string, string> = {
        ...(options.headers as Record<string, string>),
        Authorization: `Bearer ${token}`,
      };
      const res = await fetch(`${this.baseURL}${path}`, { ...options, headers });
      if (!res.ok) throw { status: res.status, body: await res.json().catch(() => ({})) };
      return res.json();
    }

    async get<T>(path: string): Promise<T> {
      return this.request<T>(path, { method: 'GET' });
    }

    async post<T>(path: string, _body?: unknown): Promise<T> {
      return this.request<T>(path, { method: 'POST' });
    }

    async put<T>(path: string, _body?: unknown): Promise<T> {
      return this.request<T>(path, { method: 'PUT' });
    }

    async del<T>(path: string, _body?: unknown): Promise<T> {
      return this.request<T>(path, { method: 'DELETE' });
    }
  }

  return { getApiClient: () => new MockApiClient() };
});

vi.mock('../store/auth', () => ({
  useAuthStore: vi.fn(),
}));

import { useAuthStore } from '../store/auth';
const mockUseAuthStore = vi.mocked(useAuthStore);

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
  };
}

function mockResponse(data: unknown) {
  return { ok: true, json: () => Promise.resolve(data) };
}

beforeEach(() => {
  mockFetch.mockReset();
  mockUseAuthStore.mockImplementation((selector) =>
    selector({ user: { id: 1 } }),
  );
});

describe('useTeamRole', () => {
  it('returns role=editor, canEdit=true for an editor', async () => {
    mockFetch.mockResolvedValue(
      mockResponse({
        team: { id: 1, name: 'Test', captain_id: 2 },
        members: [
          { user_id: 1, role: 'editor', display_name: 'Bob', joined_at: '' },
          { user_id: 2, role: 'captain', display_name: 'Alice', joined_at: '' },
        ],
      }),
    );

    const { result } = renderHook(() => useTeamRole(1), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(result.current.role).toBe('editor');
    expect(result.current.canEdit).toBe(true);
    expect(result.current.isEditor).toBe(true);
    expect(result.current.isCaptain).toBe(false);
    expect(result.current.isCocaptain).toBe(false);
  });

  it('returns role=member, canEdit=false for a member', async () => {
    mockFetch.mockResolvedValue(
      mockResponse({
        team: { id: 1, name: 'Test', captain_id: 2 },
        members: [
          { user_id: 1, role: 'member', display_name: 'Bob', joined_at: '' },
          { user_id: 2, role: 'captain', display_name: 'Alice', joined_at: '' },
        ],
      }),
    );

    const { result } = renderHook(() => useTeamRole(1), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(result.current.role).toBe('member');
    expect(result.current.canEdit).toBe(false);
    expect(result.current.isEditor).toBe(false);
  });

  it('returns role=null, canEdit=false when user is not a team member', async () => {
    mockFetch.mockResolvedValue(
      mockResponse({
        team: { id: 1, name: 'Test', captain_id: 3 },
        members: [
          { user_id: 3, role: 'captain', display_name: 'Alice', joined_at: '' },
        ],
      }),
    );

    const { result } = renderHook(() => useTeamRole(1), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(result.current.role).toBeNull();
    expect(result.current.canEdit).toBe(false);
  });

  it('returns defaults when teamId === 0', () => {
    const { result } = renderHook(() => useTeamRole(0), { wrapper: createWrapper() });

    expect(result.current.role).toBeNull();
    expect(result.current.isLoading).toBe(false);
    expect(result.current.canEdit).toBe(false);
  });

  it('returns role=null, canEdit=false when user is not logged in', () => {
    mockUseAuthStore.mockImplementation((selector) => selector({ user: null }));

    const { result } = renderHook(() => useTeamRole(1), { wrapper: createWrapper() });

    expect(result.current.role).toBeNull();
    expect(result.current.canEdit).toBe(false);
    expect(result.current.isLoading).toBe(false);
    expect(mockFetch).not.toHaveBeenCalled();
  });
});
