import { useEffect, useState } from 'react';
import type { ReactNode } from 'react';
import { listDatabaseInstances } from '../api/databaseInstances';
import type { DashboardInstance } from '../types/dashboard';
import { DATABASE_INSTANCE_STORAGE_KEY, DatabaseInstanceContext } from './databaseInstance';

export function DatabaseInstanceProvider({ children }: { children: ReactNode }) {
  const [instances, setInstances] = useState<DashboardInstance[]>([]);
  const [selectedId, setSelectedIdState] = useState(() => localStorage.getItem(DATABASE_INSTANCE_STORAGE_KEY) ?? '');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    listDatabaseInstances()
      .then((result) => {
        setInstances(result.database_instances);

        if (!selectedId || !result.database_instances.some((instance) => instance.id === selectedId)) {
          const fallback = result.database_instances[0]?.id ?? '';
          setSelectedIdState(fallback);
          if (fallback) {
            localStorage.setItem(DATABASE_INSTANCE_STORAGE_KEY, fallback);
          }
        }
      })
      .catch(() => setInstances([]))
      .finally(() => setLoading(false));
    // Only run once on mount - selectedId is read from localStorage as the
    // initial state and validated against the fetched instance list here.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const setSelectedId = (id: string) => {
    setSelectedIdState(id);
    localStorage.setItem(DATABASE_INSTANCE_STORAGE_KEY, id);
  };

  return (
    <DatabaseInstanceContext.Provider value={{ instances, selectedId, setSelectedId, loading }}>
      {children}
    </DatabaseInstanceContext.Provider>
  );
}
