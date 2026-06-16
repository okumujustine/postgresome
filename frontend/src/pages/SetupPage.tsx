import { useState, type FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { createSource, runSourceCheckup } from "@/lib/api";
import { formatRelativeTime, formatTimestamp, severityLabel, statusClasses } from "@/lib/format";
import { useWorkspace } from "@/lib/workspace-context";
import type { RunCheckupResponse } from "@/types/api";

const providerOptions = [
  { value: "postgres", label: "PostgreSQL" },
  { value: "supabase", label: "Supabase" },
  { value: "neon", label: "Neon" },
  { value: "rds", label: "Amazon RDS" },
  { value: "cloudsql", label: "Google Cloud SQL" },
];

const fieldClassName =
  "flex h-10 w-full rounded border bg-white px-3 py-2 text-sm text-foreground outline-none transition-colors focus-visible:border-[#94aee0] focus-visible:ring-2 focus-visible:ring-[#d5e3fd]";

export function SetupPage() {
  const navigate = useNavigate();
  const { sources, selectedSource, selectSource, refreshSources } = useWorkspace();
  const [form, setForm] = useState({
    provider: "postgres",
    sourceName: "",
    databaseName: "",
    host: "",
    connectionUri: "",
  });
  const [submitting, setSubmitting] = useState(false);
  const [runningCheckup, setRunningCheckup] = useState(false);
  const [result, setResult] = useState<RunCheckupResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function handleCreateSource(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSubmitting(true);
    setError(null);

    try {
      const created = await createSource({
        source: {
          kind: "direct",
          provider: form.provider,
          name: form.sourceName,
        },
        database: {
          name: form.databaseName,
          host: form.host,
        },
        connection: {
          uri: form.connectionUri,
        },
      });

      await refreshSources(created.source.id);
      setForm((current) => ({ ...current, connectionUri: "" }));
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : "Failed to create source");
    } finally {
      setSubmitting(false);
    }
  }

  async function handleRunCheckup() {
    if (!selectedSource) {
      return;
    }

    setRunningCheckup(true);
    setError(null);

    try {
      const response = await runSourceCheckup(selectedSource.source.id);
      setResult(response);
      await refreshSources(selectedSource.source.id);
      navigate("/diagnosis");
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : "Failed to run checkup");
    } finally {
      setRunningCheckup(false);
    }
  }

  return (
    <div className="mx-auto grid max-w-[1600px] gap-4 xl:grid-cols-[360px_minmax(0,1fr)]">
      <div className="technical-sheet overflow-hidden">
        <div className="border-b border-border/80 px-5 py-4">
          <div className="kicker">Connected sources</div>
          <div className="mt-2 font-heading text-[18px] font-semibold text-foreground">
            Ready for diagnosis
          </div>
        </div>
        <div className="max-h-[720px] space-y-2 overflow-y-auto p-3">
          {sources.map((item) => (
            <button
              key={item.source.id}
              type="button"
              onClick={() => selectSource(item.source.id)}
              className={`w-full rounded-md border p-4 text-left transition-colors ${
                selectedSource?.source.id === item.source.id
                  ? "border-slate-900 bg-[#f8f9ff]"
                  : "border-border bg-white hover:bg-[#f8f9ff]"
              }`}
            >
              <div className="flex items-start justify-between gap-3">
                <div>
                  <div className="text-sm font-semibold text-foreground">
                    {item.source.name}
                  </div>
                  <div className="mt-1 text-sm text-slate-600">
                    {item.database.name}
                  </div>
                </div>
                <div
                  className={`rounded border px-2 py-1 font-mono text-[11px] uppercase tracking-[0.06em] ${statusClasses(
                    item.instance.status,
                  )}`}
                >
                  {item.instance.status}
                </div>
              </div>
              <div className="mt-3 meta">{item.database.host}</div>
              <div className="mt-2 meta">
                Last check{" "}
                {item.source.last_check_completed_at
                  ? formatRelativeTime(item.source.last_check_completed_at)
                  : "not run"}
              </div>
            </button>
          ))}

          {sources.length === 0 ? (
            <div className="rounded-lg border border-dashed border-border bg-white p-6 text-sm text-muted-foreground">
              No sources connected yet. Create the first one from the form.
            </div>
          ) : null}
        </div>
      </div>

      <div className="space-y-4">
        <div className="technical-sheet p-6">
          <div className="kicker">Connection</div>
          <h2 className="mt-3 font-heading text-[24px] font-semibold tracking-[-0.01em] text-foreground">
            Add a database source
          </h2>
          <p className="mt-2 max-w-2xl text-[15px] leading-7 text-slate-600">
            Setup only needs enough to connect, identify the source, and run a fresh
            diagnosis. This is intentionally lean.
          </p>

          <form className="mt-6 grid gap-4" onSubmit={handleCreateSource}>
            <div className="grid gap-4 md:grid-cols-2">
              <label className="space-y-2">
                <span className="kicker">Provider</span>
                <select
                  className={fieldClassName}
                  value={form.provider}
                  onChange={(event) =>
                    setForm((current) => ({ ...current, provider: event.target.value }))
                  }
                >
                  {providerOptions.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              </label>

              <label className="space-y-2">
                <span className="kicker">Source name</span>
                <Input
                  value={form.sourceName}
                  onChange={(event) =>
                    setForm((current) => ({ ...current, sourceName: event.target.value }))
                  }
                  placeholder="Production primary"
                  required
                />
              </label>
            </div>

            <div className="grid gap-4 md:grid-cols-2">
              <label className="space-y-2">
                <span className="kicker">Database name</span>
                <Input
                  value={form.databaseName}
                  onChange={(event) =>
                    setForm((current) => ({ ...current, databaseName: event.target.value }))
                  }
                  placeholder="app_production"
                  required
                />
              </label>

              <label className="space-y-2">
                <span className="kicker">Host</span>
                <Input
                  value={form.host}
                  onChange={(event) =>
                    setForm((current) => ({ ...current, host: event.target.value }))
                  }
                  placeholder="db.example.internal"
                  required
                />
              </label>
            </div>

            <label className="space-y-2">
              <span className="kicker">Connection URI</span>
              <Textarea
                value={form.connectionUri}
                onChange={(event) =>
                  setForm((current) => ({
                    ...current,
                    connectionUri: event.target.value,
                  }))
                }
                placeholder="postgresql://user:password@host:5432/database"
                required
              />
            </label>

            <div className="flex flex-wrap gap-3">
              <Button type="submit" disabled={submitting}>
                {submitting ? "Saving source..." : "Save source"}
              </Button>
              {selectedSource ? (
                <Button
                  type="button"
                  variant="outline"
                  onClick={handleRunCheckup}
                  disabled={runningCheckup}
                >
                  {runningCheckup ? "Running diagnosis..." : "Run diagnosis"}
                </Button>
              ) : null}
            </div>
          </form>
        </div>

        {selectedSource ? (
          <div className="technical-sheet p-6">
            <div className="kicker">Selected source</div>
            <div className="mt-3 flex flex-wrap items-center gap-3">
              <div className="font-heading text-[18px] font-semibold text-foreground">
                {selectedSource.source.name}
              </div>
              <Badge variant="default">{selectedSource.source.provider}</Badge>
            </div>
            <div className="mt-3 grid gap-4 sm:grid-cols-2">
              <div className="rounded-md border bg-[#f8f9ff] px-4 py-4">
                <div className="kicker">Status</div>
                <div className="metric-value mt-3">
                  {selectedSource.source.last_check_status}
                </div>
              </div>
              <div className="rounded-md border bg-[#f8f9ff] px-4 py-4">
                <div className="kicker">Last completed</div>
                <div className="metric-value mt-3">
                  {selectedSource.source.last_check_completed_at
                    ? formatTimestamp(selectedSource.source.last_check_completed_at)
                    : "No checkup yet"}
                </div>
              </div>
            </div>
          </div>
        ) : null}

        {result ? (
          <div className="technical-sheet p-6">
            <div className="kicker">Recent result</div>
            <h3 className="mt-3 font-heading text-[18px] font-semibold text-foreground">
              Diagnosis completed {formatRelativeTime(result.detected_at)}
            </h3>
            <div className="mt-4 space-y-3">
              {result.findings.slice(0, 3).map((finding) => (
                <div key={`${finding.title}-${finding.detected_at}`} className="rounded-md border bg-white p-4">
                  <div className="flex items-center gap-3">
                    <Badge variant={finding.severity.toLowerCase() as "critical" | "warning" | "info"}>
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

        {error ? (
          <div className="rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
            {error}
          </div>
        ) : null}
      </div>
    </div>
  );
}
