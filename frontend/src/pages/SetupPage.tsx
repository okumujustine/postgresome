import { useState, type FormEvent } from "react";
import { Link } from "react-router-dom";
import { DismissibleAlert } from "@/components/ui/dismissible-alert";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Sheet } from "@/components/ui/sheet";
import { Textarea } from "@/components/ui/textarea";
import { createSource } from "@/lib/api";
import { formatRelativeTime } from "@/lib/format";
import { useWorkspace } from "@/lib/workspace-context";

const providerOptions = [
  { value: "postgres", label: "PostgreSQL" },
  { value: "supabase", label: "Supabase" },
  { value: "neon", label: "Neon" },
  { value: "rds", label: "Amazon RDS" },
  { value: "cloudsql", label: "Google Cloud SQL" },
];

const fieldClassName =
  "flex h-11 w-full rounded-md border border-[#111111] bg-white px-3.5 py-2 text-sm text-foreground shadow-[1px_1px_0_#111111] outline-none transition-all focus-visible:-translate-x-[1px] focus-visible:-translate-y-[1px] focus-visible:bg-[#fffdf7] focus-visible:ring-0 focus-visible:shadow-[2px_2px_0_#111111]";

export function SetupPage() {
  const { sources, selectedSource, refreshSources } = useWorkspace();
  const [form, setForm] = useState({
    provider: "postgres",
    sourceName: "",
    databaseName: "",
    host: "",
    connectionUri: "",
  });
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [addConnectionOpen, setAddConnectionOpen] = useState(false);

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
      setAddConnectionOpen(false);
      setForm({
        provider: "postgres",
        sourceName: "",
        databaseName: "",
        host: "",
        connectionUri: "",
      });
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : "Failed to create source");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="mx-auto max-w-[1600px] space-y-4">
      {error ? <DismissibleAlert>{error}</DismissibleAlert> : null}

      <div className="space-y-2">
        <h1 className="font-heading text-[24px] font-semibold tracking-[-0.04em] text-foreground">
          Connections
        </h1>
        <div className="meta">
          {sources.length} {sources.length === 1 ? "connection" : "connections"}
        </div>
      </div>

      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <div className="kicker">Settings</div>
          <div className="mt-2 font-heading text-[18px] font-semibold tracking-[-0.02em] text-foreground">
            Connection list
          </div>
        </div>
        <Sheet
          open={addConnectionOpen}
          onOpenChange={setAddConnectionOpen}
          title="Add connection"
          trigger={<Button variant="outline">Add connection</Button>}
        >
          <form className="grid gap-4" onSubmit={handleCreateSource}>
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

            <Button type="submit" disabled={submitting}>
              {submitting ? "Saving connection..." : "Save connection"}
            </Button>
          </form>
        </Sheet>
      </div>

      <div className="technical-sheet overflow-hidden">
          {sources.length > 0 ? (
            <div className="hidden grid-cols-[minmax(0,1.4fr)_minmax(0,1fr)_180px] gap-4 px-5 py-3 md:grid">
              <div className="kicker">Database</div>
              <div className="kicker">Host</div>
              <div className="kicker text-right">Status</div>
            </div>
          ) : null}

          {sources.map((item) => {
            const isCurrent = selectedSource?.source.id === item.source.id;

            return (
              <Link
                key={item.source.id}
                to={`/setup/${encodeURIComponent(item.source.id)}`}
                className="grid gap-3 border-t border-black/10 px-5 py-4 text-left transition-colors duration-150 first:border-t-0 hover:bg-[#f7f4ea] md:grid-cols-[minmax(0,1.4fr)_minmax(0,1fr)_180px] md:items-center"
              >
                <div className="min-w-0">
                  <div className="truncate text-[15px] font-semibold text-foreground">
                    {item.database.name}
                  </div>
                  <div className="mt-1 truncate text-sm text-slate-600">
                    {item.source.name}
                  </div>
                </div>

                <div className="min-w-0">
                  <div className="kicker md:hidden">Host</div>
                  <div className="truncate text-sm text-slate-700">{item.database.host}</div>
                </div>

                <div className="flex items-center justify-between gap-3 md:justify-end">
                  <div className="md:text-right">
                    <div className="kicker md:hidden">Status</div>
                    <div className="meta">
                      Last check{" "}
                      {item.source.last_check_completed_at
                        ? formatRelativeTime(item.source.last_check_completed_at)
                        : "not run"}
                    </div>
                  </div>
                  {isCurrent ? (
                    <span className="rounded-md bg-[#dce8ff] px-2.5 py-1 text-[11px] font-semibold tracking-[0.01em] text-[#254fd2]">
                      Current
                    </span>
                  ) : null}
                </div>
              </Link>
            );
          })}

          {sources.length === 0 ? (
            <div className="px-5 py-8">
              <div className="text-[15px] font-medium text-foreground">No connections yet</div>
              <div className="mt-2 text-sm text-muted-foreground">
                Add the first one to start diagnosing.
              </div>
            </div>
          ) : null}
      </div>
    </div>
  );
}
