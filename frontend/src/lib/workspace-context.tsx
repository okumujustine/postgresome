import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";
import { listSources } from "@/lib/api";
import type { SourceRecord } from "@/types/api";

interface WorkspaceContextValue {
  sources: SourceRecord[];
  selectedSource: SourceRecord | null;
  selectedInstanceId: string | null;
  loading: boolean;
  error: string | null;
  selectSource: (sourceId: string) => void;
  refreshSources: (preferredSourceId?: string) => Promise<void>;
}

const WorkspaceContext = createContext<WorkspaceContextValue | null>(null);

function pickSelectedSourceId(
  sources: SourceRecord[],
  currentId: string | null,
  preferredSourceId?: string,
) {
  if (preferredSourceId && sources.some((item) => item.source.id === preferredSourceId)) {
    return preferredSourceId;
  }

  if (currentId && sources.some((item) => item.source.id === currentId)) {
    return currentId;
  }

  return sources[0]?.source.id ?? null;
}

export function WorkspaceProvider({ children }: { children: ReactNode }) {
  const [sources, setSources] = useState<SourceRecord[]>([]);
  const [selectedSourceId, setSelectedSourceId] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  async function refreshSources(preferredSourceId?: string) {
    setLoading(true);
    setError(null);

    try {
      const response = await listSources();
      setSources(response.sources);
      setSelectedSourceId((current) =>
        pickSelectedSourceId(response.sources, current, preferredSourceId),
      );
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : "Failed to load sources");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void refreshSources();
  }, []);

  const selectedSource =
    sources.find((item) => item.source.id === selectedSourceId) ?? null;

  const value = {
    sources,
    selectedSource,
    selectedInstanceId: selectedSource?.instance.id ?? null,
    loading,
    error,
    selectSource: (sourceId: string) => setSelectedSourceId(sourceId),
    refreshSources,
  };

  return (
    <WorkspaceContext.Provider value={value}>{children}</WorkspaceContext.Provider>
  );
}

export function useWorkspace() {
  const context = useContext(WorkspaceContext);

  if (!context) {
    throw new Error("useWorkspace must be used inside WorkspaceProvider");
  }

  return context;
}
