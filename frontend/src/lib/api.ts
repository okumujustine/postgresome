import type {
  CreateSourceInput,
  FindingDetailResponse,
  FindingListResponse,
  ListSourcesResponse,
  QueryStatsResponse,
  RunCheckupResponse,
  SourceRecord,
} from "@/types/api";

const API_BASE = (import.meta.env.VITE_API_BASE_URL || "").replace(/\/$/, "");

class APIError extends Error {
  status: number;

  constructor(message: string, status: number) {
    super(message);
    this.status = status;
  }
}

async function apiRequest<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers || {}),
    },
    ...init,
  });

  if (!response.ok) {
    const message = await response.text();
    throw new APIError(message || "Request failed", response.status);
  }

  return response.json() as Promise<T>;
}

export async function listSources() {
  return apiRequest<ListSourcesResponse>("/api/sources");
}

export async function createSource(input: CreateSourceInput) {
  return apiRequest<SourceRecord>("/api/sources", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function runSourceCheckup(sourceId: string) {
  return apiRequest<RunCheckupResponse>(`/api/sources/${sourceId}/checkup`, {
    method: "POST",
  });
}

export async function listFindings(params: {
  databaseInstanceId: string;
  status?: "open" | "resolved" | "all";
  limit?: number;
  range?: string;
}) {
  const search = new URLSearchParams({
    database_instance_id: params.databaseInstanceId,
  });

  if (params.status) {
    search.set("status", params.status);
  }
  if (params.limit) {
    search.set("limit", String(params.limit));
  }
  if (params.range) {
    search.set("range", params.range);
  }

  return apiRequest<FindingListResponse>(`/api/findings?${search.toString()}`);
}

export async function getFinding(findingId: string, databaseInstanceId: string) {
  const search = new URLSearchParams({
    database_instance_id: databaseInstanceId,
  });

  return apiRequest<FindingDetailResponse>(
    `/api/findings/${findingId}?${search.toString()}`,
  );
}

export async function listQueries(databaseInstanceId: string) {
  const search = new URLSearchParams({
    database_instance_id: databaseInstanceId,
  });

  return apiRequest<QueryStatsResponse>(`/api/queries?${search.toString()}`);
}

export { APIError };

