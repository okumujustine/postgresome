import { useEffect, useState } from "react";
import { listFindings } from "@/lib/api";
import { formatRelativeTime, formatTimestamp, severityClasses, severityLabel } from "@/lib/format";
import { useWorkspace } from "@/lib/workspace-context";
import type { Finding } from "@/types/api";

interface HistoryEvent {
  time: string;
  title: string;
  description: string;
  severity: string;
}

function buildEvents(findings: Finding[]) {
  const events: HistoryEvent[] = [];

  findings.forEach((finding) => {
    events.push({
      time: finding.detected_at,
      title: finding.title,
      description: `Finding opened for ${finding.resource_name}.`,
      severity: finding.severity,
    });

    if (finding.last_regressed_at) {
      events.push({
        time: finding.last_regressed_at,
        title: `${finding.title} regressed`,
        description: finding.change_summary || "Evidence moved further away from baseline.",
        severity: finding.severity,
      });
    }

    if (finding.improving_since) {
      events.push({
        time: finding.improving_since,
        title: `${finding.title} began improving`,
        description:
          finding.verification_summary || "Signals started moving back toward a healthier range.",
        severity: "info",
      });
    }

    if (finding.verified_fixed_at) {
      events.push({
        time: finding.verified_fixed_at,
        title: `${finding.title} verified fixed`,
        description: "Postgresome confirmed the finding is no longer active.",
        severity: "info",
      });
    }
  });

  return events.sort((a, b) => +new Date(b.time) - +new Date(a.time));
}

export function HistoryPage() {
  const { selectedInstanceId, selectedSource } = useWorkspace();
  const [events, setEvents] = useState<HistoryEvent[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!selectedInstanceId) {
      setEvents([]);
      return;
    }

    setLoading(true);
    setError(null);

    void listFindings({
      databaseInstanceId: selectedInstanceId,
      status: "all",
      limit: 50,
    })
      .then((response) => {
        setEvents(buildEvents(response.findings));
      })
      .catch((caught) => {
        setError(caught instanceof Error ? caught.message : "Failed to load history");
      })
      .finally(() => {
        setLoading(false);
      });
  }, [selectedInstanceId]);

  return (
    <div className="mx-auto max-w-[1600px] space-y-4">
      <div className="technical-sheet px-6 py-5">
        <div className="kicker">Diagnosis timeline</div>
        <div className="mt-3 font-heading text-[24px] font-semibold tracking-[-0.01em] text-foreground">
          {selectedSource?.database.name || "No database selected"}
        </div>
        <p className="mt-2 max-w-2xl text-[15px] leading-7 text-slate-600">
          Follow when a finding appeared, when it regressed, and whether the evidence
          later improved.
        </p>
      </div>

      {error ? (
        <div className="rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      ) : null}

      <div className="technical-sheet p-6">
        {loading ? (
          <div className="text-sm text-muted-foreground">Loading history...</div>
        ) : (
          <div className="relative">
            <div className="absolute left-[107px] top-1 bottom-1 hidden w-px bg-border sm:block" />
            <div className="space-y-5">
              {events.map((event, index) => {
                const tone = severityClasses(event.severity);

                return (
                  <div
                    key={`${event.title}-${event.time}-${index}`}
                    className="grid gap-3 sm:grid-cols-[96px_22px_minmax(0,1fr)] sm:items-start"
                  >
                    <div className="meta pt-1">{formatTimestamp(event.time)}</div>
                    <div className="hidden pt-1 sm:flex sm:justify-center">
                      <div className={`mt-1 h-3 w-3 rounded-full ${tone.accent}`} />
                    </div>
                    <div className="rounded-md border bg-white p-4">
                      <div className="flex flex-wrap items-center gap-3">
                        <div className={`kicker ${tone.text}`}>
                          {severityLabel(event.severity)}
                        </div>
                        <div className="meta">{formatRelativeTime(event.time)}</div>
                      </div>
                      <div className="mt-2 font-heading text-[18px] font-semibold text-foreground">
                        {event.title}
                      </div>
                      <div className="mt-2 text-sm leading-6 text-slate-600">
                        {event.description}
                      </div>
                    </div>
                  </div>
                );
              })}

              {events.length === 0 ? (
                <div className="rounded-lg border border-dashed border-border bg-white p-6 text-sm text-muted-foreground">
                  No diagnosis history is available yet for this database.
                </div>
              ) : null}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
