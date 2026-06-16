import { useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { Input } from "@/components/ui/input";
import { listQueries } from "@/lib/api";
import { formatDurationMs, formatNumber, formatTimestamp } from "@/lib/format";
import { useWorkspace } from "@/lib/workspace-context";
import type { QueryStat, QueryStatsResponse } from "@/types/api";

function selectInitialQuery(
  queries: QueryStat[],
  focusQueryId: string | null,
  currentQueryId: string | null,
) {
  if (focusQueryId && queries.some((item) => item.query_id === focusQueryId)) {
    return focusQueryId;
  }
  if (currentQueryId && queries.some((item) => item.query_id === currentQueryId)) {
    return currentQueryId;
  }
  return queries[0]?.query_id ?? null;
}

export function QueryExplorerPage() {
  const { selectedInstanceId } = useWorkspace();
  const [data, setData] = useState<QueryStatsResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [search, setSearch] = useState("");
  const [searchParams, setSearchParams] = useSearchParams();

  const selectedQueryId = searchParams.get("focus");

  useEffect(() => {
    if (!selectedInstanceId) {
      setData(null);
      return;
    }

    setLoading(true);
    setError(null);

    void listQueries(selectedInstanceId)
      .then((response) => {
        setData(response);
        const nextQueryId = selectInitialQuery(
          response.queries,
          selectedQueryId,
          selectedQueryId,
        );
        if (nextQueryId) {
          setSearchParams({ focus: nextQueryId }, { replace: true });
        }
      })
      .catch((caught) => {
        setError(caught instanceof Error ? caught.message : "Failed to load queries");
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

  const selectedQuery =
    filteredQueries.find((query) => query.query_id === selectedQueryId) ||
    data?.queries.find((query) => query.query_id === selectedQueryId) ||
    null;

  return (
    <div className="mx-auto max-w-[1600px] space-y-4">
      <div className="technical-sheet px-6 py-5">
        <div className="kicker">Evidence drill-down</div>
        <h2 className="mt-3 font-heading text-[24px] font-semibold tracking-[-0.01em] text-foreground">
          Query Explorer
        </h2>
        <p className="mt-2 max-w-3xl text-[15px] leading-7 text-slate-600">
          Use this only when a diagnosis points toward SQL behavior. This is not a raw
          metrics page; it is supporting evidence for a selected issue.
        </p>
      </div>

      {error ? (
        <div className="rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      ) : null}

      <div className="grid min-h-[680px] gap-4 lg:grid-cols-[360px_minmax(0,1fr)]">
        <div className="technical-sheet overflow-hidden">
          <div className="space-y-4 border-b border-border/80 px-4 py-4">
            <div className="kicker">Latest query snapshot</div>
            <Input
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              placeholder="Search query hash, SQL, or user"
            />
            <div className="meta">
              {data?.collected_at
                ? `Collected ${formatTimestamp(data.collected_at)}`
                : "No query snapshot available"}
            </div>
          </div>

          <div className="h-[560px] overflow-y-auto p-3">
            {loading ? (
              <div className="p-3 text-sm text-muted-foreground">Loading queries...</div>
            ) : (
              <div className="space-y-2">
                {filteredQueries.map((query) => {
                  const selected = selectedQueryId === query.query_id;

                  return (
                    <button
                      key={query.query_id}
                      type="button"
                      onClick={() => setSearchParams({ focus: query.query_id })}
                      className={`w-full rounded-md border p-4 text-left transition-colors ${
                        selected
                          ? "border-slate-900 bg-[#f8f9ff]"
                          : "border-border bg-white hover:bg-[#f8f9ff]"
                      }`}
                    >
                      <div className="flex items-start justify-between gap-3">
                        <div className="kicker">Hash {query.query_id.slice(0, 8)}</div>
                        <div className="meta">{formatDurationMs(query.mean_exec_time_ms)}</div>
                      </div>
                      <div className="mt-3 line-clamp-2 font-mono text-[12px] leading-6 text-slate-700">
                        {query.query}
                      </div>
                      <div className="mt-3 flex items-center gap-3 text-sm text-slate-600">
                        <span>{formatNumber(query.calls)} calls</span>
                        <span>{formatDurationMs(query.total_exec_time_ms)} total</span>
                      </div>
                    </button>
                  );
                })}
              </div>
            )}
          </div>
        </div>

        <div className="technical-sheet p-6">
          {selectedQuery ? (
            <div className="space-y-6">
              <div className="border-b border-border/80 pb-6">
                <div className="kicker">Selected query</div>
                <h3 className="mt-3 font-heading text-[30px] font-semibold tracking-[-0.02em] text-foreground">
                  Investigation detail
                </h3>
                <div className="meta mt-2">
                  Hash {selectedQuery.query_id} / user {selectedQuery.user_name}
                </div>
              </div>

              <section className="space-y-4">
                <div className="kicker">Normalized SQL</div>
                <pre className="overflow-x-auto rounded-md border bg-[#eff4ff] p-5 font-mono text-[12px] leading-6 text-[#0b1c30]">
                  {selectedQuery.query}
                </pre>
              </section>

              <section className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
                {[
                  {
                    label: "Mean latency",
                    value: formatDurationMs(selectedQuery.mean_exec_time_ms),
                  },
                  {
                    label: "Total time",
                    value: formatDurationMs(selectedQuery.total_exec_time_ms),
                  },
                  {
                    label: "Shared blocks read",
                    value: formatNumber(selectedQuery.shared_blocks_read),
                  },
                  {
                    label: "Shared blocks hit",
                    value: formatNumber(selectedQuery.shared_blocks_hit),
                  },
                ].map((item) => (
                  <div key={item.label} className="rounded-md border bg-[#f8f9ff] px-4 py-4">
                    <div className="kicker">{item.label}</div>
                    <div className="mt-3 font-heading text-[24px] font-semibold text-foreground">
                      {item.value}
                    </div>
                  </div>
                ))}
              </section>
            </div>
          ) : (
            <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
              Select a query to inspect its evidence.
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
