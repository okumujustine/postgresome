import { useEffect, useState } from "react";
import { DismissibleAlert } from "@/components/ui/dismissible-alert";
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
      {error ? <DismissibleAlert>{error}</DismissibleAlert> : null}

      <div className="technical-sheet p-6">
        <div className="mb-6">
          <div className="kicker">Diagnosis timeline</div>
          <div className="meta mt-2">{selectedSource?.database.name || "No database selected"}</div>
        </div>
        {loading ? (
          <div className="text-sm text-muted-foreground">Loading history...</div>
        ) : (
          <div className="relative">
            <div className="absolute bottom-1 left-[107px] top-1 hidden w-[2px] bg-[#111111] sm:block" />
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
                    <div className="rounded-md border-2 border-[#111111] bg-white p-4 shadow-[2px_2px_0_#111111]">
                      <div className="flex flex-wrap items-center gap-3">
                        <div className={`kicker ${tone.text}`}>
                          {severityLabel(event.severity)}
                        </div>
                        <div className="meta">{formatRelativeTime(event.time)}</div>
                      </div>
                      <div className="mt-2 font-heading text-[18px] font-semibold tracking-[-0.02em] text-foreground">
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
                <div className="rounded-md border-2 border-dashed border-border bg-white p-6 text-sm text-muted-foreground">
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
