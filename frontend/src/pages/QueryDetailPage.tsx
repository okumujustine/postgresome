import { ChevronLeft } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { DismissibleAlert } from "@/components/ui/dismissible-alert";
import { listQueries } from "@/lib/api";
import { formatDurationMs, formatNumber, formatTimestamp } from "@/lib/format";
import { useWorkspace } from "@/lib/workspace-context";
import type { QueryStat, QueryStatsResponse } from "@/types/api";

export function QueryDetailPage() {
  const { queryId } = useParams<{ queryId: string }>();
  const { selectedInstanceId, selectedSource } = useWorkspace();
  const [data, setData] = useState<QueryStatsResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

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
        setError(caught instanceof Error ? caught.message : "Failed to load query detail");
        setData(null);
      })
      .finally(() => {
        setLoading(false);
      });
  }, [selectedInstanceId]);

  const selectedQuery = useMemo<QueryStat | null>(() => {
    if (!queryId || !data) {
      return null;
    }

    return data.queries.find((query) => query.query_id === queryId) ?? null;
  }, [data, queryId]);

  const summaryStats = selectedQuery
    ? [
        { label: "Mean latency", value: formatDurationMs(selectedQuery.mean_exec_time_ms) },
        { label: "Total time", value: formatDurationMs(selectedQuery.total_exec_time_ms) },
        { label: "Calls", value: formatNumber(selectedQuery.calls) },
        { label: "Rows returned", value: formatNumber(selectedQuery.rows_returned) },
      ]
    : [];

  const secondaryStats = selectedQuery
    ? [
        { label: "Shared blocks read", value: formatNumber(selectedQuery.shared_blocks_read) },
        { label: "Shared blocks hit", value: formatNumber(selectedQuery.shared_blocks_hit) },
        { label: "Min latency", value: formatDurationMs(selectedQuery.min_exec_time_ms) },
        { label: "Max latency", value: formatDurationMs(selectedQuery.max_exec_time_ms) },
      ]
    : [];

  const backHref = queryId ? `/queries?focus=${encodeURIComponent(queryId)}` : "/queries";

  return (
    <div className="mx-auto max-w-[1200px] space-y-4">
      <Link
        to={backHref}
        className="inline-flex items-center gap-2 rounded-md border border-[#111111] bg-white px-3 py-2 text-[14px] font-semibold text-slate-700 shadow-[1px_1px_0_#111111] transition-all hover:-translate-x-[1px] hover:-translate-y-[1px] hover:bg-[#f4f1e6] hover:text-slate-950 hover:shadow-[2px_2px_0_#111111]"
      >
        <ChevronLeft className="h-4 w-4" />
        Back to Query Explorer
      </Link>

      {error ? <DismissibleAlert>{error}</DismissibleAlert> : null}

      <div className="technical-sheet p-6">
        {loading ? (
          <div className="text-sm text-muted-foreground">Loading query detail...</div>
        ) : !selectedInstanceId ? (
          <div className="text-sm text-muted-foreground">
            Select a connection in Settings to inspect query evidence.
          </div>
        ) : !selectedQuery ? (
          <div className="text-sm text-muted-foreground">
            This query could not be found in the latest snapshot.
          </div>
        ) : (
          <div className="space-y-8">
            <div className="border-b border-border pb-8">
              <div className="kicker">Investigation detail</div>
              <h3 className="report-title mt-3">Query {selectedQuery.query_id.slice(0, 12)}</h3>
              <p className="report-lead mt-4">
                This query is currently averaging {formatDurationMs(selectedQuery.mean_exec_time_ms)} across{" "}
                {formatNumber(selectedQuery.calls)} calls, with {formatDurationMs(selectedQuery.total_exec_time_ms)} of
                total execution time in the latest snapshot.
              </p>
              <div className="meta mt-4">
                User {selectedQuery.user_name} / snapshot{" "}
                {data?.collected_at ? formatTimestamp(data.collected_at) : "not available"}
              </div>
            </div>

            <div className="flex flex-wrap gap-2 border-b border-border pb-4">
              {[
                { href: "#query-overview", label: "Overview" },
                { href: "#query-sql", label: "SQL" },
                { href: "#query-evidence", label: "Evidence" },
              ].map((item) => (
                <a
                  key={item.href}
                  href={item.href}
                  className="rounded-md border border-[#111111] bg-white px-3 py-2 text-[11px] font-semibold tracking-[0.01em] text-slate-700 shadow-[1px_1px_0_#111111] transition-all hover:-translate-x-[1px] hover:-translate-y-[1px] hover:bg-[#f4f1e6] hover:shadow-[2px_2px_0_#111111]"
                >
                  {item.label}
                </a>
              ))}
            </div>

            <section id="query-overview" className="scroll-mt-24 space-y-4">
              <div className="kicker">What stands out</div>
              <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
                {summaryStats.map((item) => (
                  <div key={item.label} className="summary-strip px-4 py-4">
                    <div className="kicker">{item.label}</div>
                    <div className="mt-3 font-heading text-[24px] font-semibold tracking-[-0.03em] text-foreground">
                      {item.value}
                    </div>
                  </div>
                ))}
              </div>
            </section>

            <section id="query-sql" className="scroll-mt-24 space-y-4">
              <div className="kicker">Normalized SQL</div>
              <pre className="overflow-x-auto rounded-md border border-[#111111] bg-slate-950 p-5 font-mono text-[12px] leading-6 text-white shadow-[1px_1px_0_#111111]">
                {selectedQuery.query}
              </pre>
            </section>

            <section id="query-evidence" className="detail-block scroll-mt-24 p-5">
              <div className="kicker">Secondary evidence</div>
              <div className="mt-4 grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
                {secondaryStats.map((item) => (
                  <div key={item.label} className="space-y-2">
                    <div className="kicker">{item.label}</div>
                    <div className="metric-value text-[15px]">{item.value}</div>
                  </div>
                ))}
              </div>
            </section>
          </div>
        )}
      </div>
    </div>
  );
}
