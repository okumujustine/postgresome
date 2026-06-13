import { apiGet } from './client';
import type { DatabaseInstancesResponse } from '../types/dashboard';

export function listDatabaseInstances(): Promise<DatabaseInstancesResponse> {
  return apiGet<DatabaseInstancesResponse>('/api/database-instances');
}
