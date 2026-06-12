const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:9090';

export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

export async function apiGet<T>(path: string, params?: Record<string, string | undefined>): Promise<T> {
  const url = new URL(path, API_BASE_URL);

  if (params) {
    for (const [key, value] of Object.entries(params)) {
      if (value) {
        url.searchParams.set(key, value);
      }
    }
  }

  const response = await fetch(url.toString());

  if (!response.ok) {
    throw new ApiError(response.status, `request to ${path} failed with status ${response.status}`);
  }

  return response.json() as Promise<T>;
}
