export interface ApiResponse<T> {
  status: number;
  body: T;
  headers: Headers;
}

const BASE_URL = process.env.E2E_API_URL || 'http://localhost:8080';

async function request<T>(
  method: string,
  path: string,
  body?: unknown,
  token?: string,
): Promise<ApiResponse<T>> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const url = `${BASE_URL}${path}`;
  const response = await fetch(url, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });

  let responseBody: T;
  const contentType = response.headers.get('content-type') || '';
  if (contentType.includes('application/json')) {
    responseBody = await response.json();
  } else {
    responseBody = (await response.text()) as unknown as T;
  }

  return {
    status: response.status,
    body: responseBody,
    headers: response.headers,
  };
}

export async function apiGet<T>(path: string, token?: string): Promise<ApiResponse<T>> {
  return request<T>('GET', path, undefined, token);
}

export async function apiPost<T>(path: string, body?: unknown, token?: string): Promise<ApiResponse<T>> {
  return request<T>('POST', path, body, token);
}

export async function apiPut<T>(path: string, body?: unknown, token?: string): Promise<ApiResponse<T>> {
  return request<T>('PUT', path, body, token);
}

export async function apiDelete<T>(path: string, token?: string): Promise<ApiResponse<T>> {
  return request<T>('DELETE', path, undefined, token);
}
