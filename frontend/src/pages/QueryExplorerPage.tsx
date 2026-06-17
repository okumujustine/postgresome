import { ArrowRight } from "lucide-react";
import { useEffect, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { Badge } from "@/components/ui/badge";
import { DismissibleAlert } from "@/components/ui/dismissible-alert";
import { Input } from "@/components/ui/input";
import { listQueries } from "@/lib/api";
import { formatDurationMs, formatNumber, formatTimestamp } from "@/lib/format";
import { useWorkspace } from "@/lib/workspace-context";
import type { QueryStatsResponse } from "@/types/api";

export function QueryExplorerPage() {
  const { selectedInstanceId, selectedSource } = useWorkspace();
  const [data, setData] = useState<QueryStatsResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [search, setSearch] = useState("");
  const [searchParams] = useSearchParams();
  const focusedQueryId = searchParams.get("focus");

  useEffect(() => {
    if (!selectedInstanceId) {
      setData(null);
      setLoading(false);
      return;
    }

    setLoading(true);
    setError(null);

    void listQueries(selectedInstanceId)
      .then((response) => {
        setData(response);
      })
      .catch((caught) => {
        setError(caught instanceof Error ? caught.message : "Failed to load queries");
        setData(null);
      })
      .finally(() => {
        setLoading(false);
      });
  }, [selectedInstanceId]);

  const filteredQueries =
    data?.queries.filter((query) => {
      const needle = search.trim().toLowerCase();
      if (!needle) {
        return true;
      }

      return [query.query_id, query.query, query.user_name]
        .join(" ")
        .toLowerCase()
        .includes(needle);
    }) ?? [];

  const currentSourceName = selectedSource?.database.name ?? "No source selected";

  return (
    <div className="mx-auto max-w-[1600px] space-y-4">
      {error ? <DismissibleAlert>{error}</DismissibleAlert> : null}

      <div className="space-y-1">
        <h1 className="font-heading text-[24px] font-semibold tracking-[-0.04em] text-foreground">
          Query listing
        </h1>
      </div>

      <div className="panel p-4">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="w-full max-w-sm">
            <Input
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              placeholder="Search query hash, SQL, or user"
            />
          </div>
          <div className="meta">
            {currentSourceName}
            {data?.collected_at ? ` / ${formatTimestamp(data.collected_at)}` : ""}
          </div>
        </div>
      </div>

      <div className="technical-sheet overflow-hidden">
        <div className="table-header hidden md:grid md:grid-cols-[160px_minmax(0,1fr)_220px_40px] md:items-center">
          <div>Query id</div>
          <div>Normalized SQL</div>
          <div>Stats</div>
          <div />
        </div>

        <div>
          {!selectedInstanceId ? (
            <div className="px-6 py-8 text-sm text-muted-foreground">
              Select a connection in Settings to review query evidence.
            </div>
          ) : loading ? (
            <div className="px-6 py-6 text-sm text-muted-foreground">Loading queries...</div>
          ) : (
            filteredQueries.map((query) => {
              const focused = focusedQueryId === query.query_id;

              return (
                <Link
                  key={query.query_id}
                  to={`/queries/${encodeURIComponent(query.query_id)}`}
                  className={`table-row grid gap-4 px-6 py-5 md:grid-cols-[160px_minmax(0,1fr)_220px_40px] md:items-center ${
                    focused ? "surface-selection" : ""
                  }`}
                >
                  <div className="space-y-2">
                    <div className="kicker md:hidden">Hash</div>
                    <div className="data-mono text-[12px] text-slate-700">
                      {query.query_id.slice(0, 12)}
                    </div>
                  </div>

                  <div className="space-y-2">
                    <div className="kicker md:hidden">Query</div>
                    <div className="line-clamp-2 font-mono text-[13px] leading-6 text-slate-800">
                      {query.query}
                    </div>
                  <div className="flex flex-wrap items-center gap-2">
                      <Badge variant="default">user {query.user_name}</Badge>
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-4 md:grid-cols-1">
                  <div>
                      <div className="kicker md:hidden">Mean latency</div>
                      <div className="metric-value mt-2">
                        {formatDurationMs(query.mean_exec_time_ms)}
                      </div>
                    </div>
                    <div>
                      <div className="kicker md:hidden">Calls</div>
                      <div className="metric-value mt-2">{formatNumber(query.calls)}</div>
                    </div>
                  </div>

                  <div className="hidden justify-self-end md:block">
                    <ArrowRight className="h-4 w-4 text-slate-700" />
                  </div>
                </Link>
              );
            })
          )}

          {!loading && selectedInstanceId && filteredQueries.length === 0 ? (
            <div className="px-6 py-8 text-sm text-muted-foreground">
              {search.trim()
                ? "No queries match the current filter."
                : "No query evidence is available for this database yet."}
            </div>
          ) : null}
        </div>
      </div>
    </div>
  );
}
