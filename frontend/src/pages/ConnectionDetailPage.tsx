import { ChevronLeft } from "lucide-react";
import { useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { DismissibleAlert } from "@/components/ui/dismissible-alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { runSourceCheckup } from "@/lib/api";
import {
  formatRelativeTime,
  formatTimestamp,
  severityLabel,
  statusClasses,
} from "@/lib/format";
import { useWorkspace } from "@/lib/workspace-context";
import type { RunCheckupResponse, SourceRecord } from "@/types/api";

export function ConnectionDetailPage() {
  const { sourceId } = useParams<{ sourceId: string }>();
  const navigate = useNavigate();
  const { sources, selectedSource, selectSource, refreshSources } = useWorkspace();
  const [runningCheckup, setRunningCheckup] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<RunCheckupResponse | null>(null);

  const source =
    sources.find((item) => item.source.id === sourceId) ?? null;

  async function handleRunDiagnosis(record: SourceRecord) {
    setRunningCheckup(true);
    setError(null);

    try {
      const response = await runSourceCheckup(record.source.id);
      setResult(response);
      await refreshSources(record.source.id);
      navigate("/diagnosis");
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : "Failed to run checkup");
    } finally {
      setRunningCheckup(false);
    }
  }

  if (!source) {
    return (
      <div className="mx-auto max-w-[1200px] space-y-4">
        <Link
          to="/setup"
          className="inline-flex items-center gap-2 rounded-md border border-[#111111] bg-white px-3 py-2 text-[14px] font-semibold text-slate-700 shadow-[1px_1px_0_#111111] transition-all hover:-translate-x-[1px] hover:-translate-y-[1px] hover:bg-[#f4f1e6] hover:text-slate-950 hover:shadow-[2px_2px_0_#111111]"
        >
          <ChevronLeft className="h-4 w-4" />
          Back to Connections
        </Link>

        <div className="technical-sheet p-6 text-sm text-muted-foreground">
          This connection could not be found.
        </div>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-[1200px] space-y-4">
      <Link
        to="/setup"
        className="inline-flex items-center gap-2 rounded-md border border-[#111111] bg-white px-3 py-2 text-[14px] font-semibold text-slate-700 shadow-[1px_1px_0_#111111] transition-all hover:-translate-x-[1px] hover:-translate-y-[1px] hover:bg-[#f4f1e6] hover:text-slate-950 hover:shadow-[2px_2px_0_#111111]"
      >
        <ChevronLeft className="h-4 w-4" />
        Back to Connections
      </Link>

      {error ? <DismissibleAlert>{error}</DismissibleAlert> : null}

      <div className="space-y-1">
        <h1 className="font-heading text-[24px] font-semibold tracking-[-0.04em] text-foreground">
          Connection details
        </h1>
      </div>

      <div className="technical-sheet p-6">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <div className="kicker">Selected connection</div>
            <div className="mt-3 flex flex-wrap items-center gap-3">
              <div className="font-heading text-[20px] font-semibold tracking-[-0.03em] text-foreground">
                {source.database.name}
              </div>
              <Badge variant="default">{source.source.provider}</Badge>
              {selectedSource?.source.id === source.source.id ? (
                <Badge variant="info">Current source</Badge>
              ) : null}
            </div>
            <div className="meta mt-2">{source.source.name}</div>
          </div>

          <div className="flex flex-wrap gap-3">
            {selectedSource?.source.id !== source.source.id ? (
              <Button
                variant="outline"
                type="button"
                onClick={() => selectSource(source.source.id)}
              >
                Use as current source
              </Button>
            ) : null}
            <Button
              type="button"
              onClick={() => void handleRunDiagnosis(source)}
              disabled={runningCheckup}
            >
              {runningCheckup ? "Running diagnosis..." : "Run diagnosis"}
            </Button>
          </div>
        </div>

        <div className="mt-6 grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          <div className="surface-muted px-4 py-4">
            <div className="kicker">Host</div>
            <div className="mt-3 text-sm font-medium text-foreground">
              {source.database.host}
            </div>
          </div>
          <div className="surface-muted px-4 py-4">
            <div className="kicker">Instance status</div>
            <div className="mt-3">
              <span
                className={`inline-flex rounded-md border px-2.5 py-1 text-[12px] font-medium ${statusClasses(
                  source.instance.status,
                )}`}
              >
                {source.instance.status}
              </span>
            </div>
          </div>
          <div className="surface-muted px-4 py-4">
            <div className="kicker">Setup state</div>
            <div className="mt-3 text-sm font-medium capitalize text-foreground">
              {source.source.setup_state}
            </div>
          </div>
          <div className="surface-muted px-4 py-4">
            <div className="kicker">Last check</div>
            <div className="mt-3 text-sm font-medium text-foreground">
              {source.source.last_check_completed_at
                ? formatTimestamp(source.source.last_check_completed_at)
                : "No checkup yet"}
            </div>
          </div>
          <div className="surface-muted px-4 py-4">
            <div className="kicker">Last result</div>
            <div className="mt-3">
              <span
                className={`inline-flex rounded-md border px-2.5 py-1 text-[12px] font-medium ${statusClasses(
                  source.source.last_check_status,
                )}`}
              >
                {source.source.last_check_status}
              </span>
            </div>
          </div>
        </div>
      </div>

      {result ? (
        <div className="technical-sheet p-6">
          <div className="kicker">Recent result</div>
          <h3 className="mt-3 font-heading text-[18px] font-semibold text-foreground">
            Diagnosis completed {formatRelativeTime(result.detected_at)}
          </h3>
          <div className="mt-4 space-y-3">
            {result.findings.slice(0, 3).map((finding) => (
              <div
                key={`${finding.title}-${finding.detected_at}`}
                className="rounded-md border border-[#111111] bg-white p-4"
              >
                <div className="flex items-center gap-3">
                  <Badge
                    variant={finding.severity.toLowerCase() as "critical" | "warning" | "info"}
                  >
                    {severityLabel(finding.severity)}
                  </Badge>
                  <div className="font-heading text-sm font-semibold text-foreground">
                    {finding.title}
                  </div>
                </div>
                <p className="mt-2 text-sm leading-6 text-slate-600">
                  {finding.problem_summary}
                </p>
              </div>
            ))}
          </div>
        </div>
      ) : null}
    </div>
  );
}
