import { createContext, useContext } from 'react';
import type { DashboardInstance } from '../types/dashboard';

export const DATABASE_INSTANCE_STORAGE_KEY = 'postgresome:databaseInstanceId';

export interface DatabaseInstanceContextValue {
  instances: DashboardInstance[];
  selectedId: string;
  setSelectedId: (id: string) => void;
  loading: boolean;
}

export const DatabaseInstanceContext = createContext<DatabaseInstanceContextValue | null>(null);

export function useDatabaseInstance(): DatabaseInstanceContextValue {
  const ctx = useContext(DatabaseInstanceContext);
  if (!ctx) {
    throw new Error('useDatabaseInstance must be used within a DatabaseInstanceProvider');
  }
  return ctx;
}
