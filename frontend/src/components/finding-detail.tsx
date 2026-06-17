import { ArrowRight, Database, FileSearch, Gauge, ShieldAlert } from "lucide-react";
import type { ReactNode } from "react";
import { Link } from "react-router-dom";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  formatDurationMs,
  formatNumber,
  formatPercent,
  formatRelativeTime,
  statusClasses,
  formatTimestamp,
  severityLabel,
} from "@/lib/format";
import type { FindingDetailResponse } from "@/types/api";

function DetailSection({
  id,
  label,
  title,
  children,
}: {
  id?: string;
  label: string;
  title: string;
  children: ReactNode;
}) {
  return (
    <section id={id} className="scroll-mt-24 space-y-5 py-7 first:pt-0">
      <div className="space-y-2">
        <div className="kicker">{label}</div>
        <h3 className="font-heading text-[20px] font-semibold tracking-[-0.03em] text-foreground">
          {title}
        </h3>
      </div>
      {children}
    </section>
  );
}

export function FindingDetail({
  detail,
  loading,
}: {
  detail: FindingDetailResponse | null;
  loading: boolean;
}) {
  if (loading) {
    return (
      <div className="technical-sheet flex h-full min-h-[540px] items-center justify-center p-8 text-sm text-muted-foreground">
        Loading diagnosis report...
      </div>
    );
  }

  if (!detail) {
    return (
      <div className="technical-sheet flex h-full min-h-[540px] items-center justify-center p-8 text-sm text-muted-foreground">
        Select a finding to review the diagnosis report.
      </div>
    );
  }

  const { finding, evidence, historical_context: history, related_query, related_table } =
    detail;

  return (
    <div className="technical-sheet min-h-[540px] p-6">
      <div className="grid gap-8 border-b-2 border-border pb-8 xl:grid-cols-[minmax(0,1fr)_360px]">
        <div className="space-y-5">
          <div className="flex flex-wrap items-center gap-3">
            <Badge variant={finding.severity.toLowerCase() as "critical" | "warning" | "info"}>
              {severityLabel(finding.severity)}
            </Badge>
            <div className="meta">Finding {finding.id}</div>
            <div className="meta">{formatRelativeTime(finding.detected_at)}</div>
          </div>

          <div>
            <h2 className="report-title max-w-3xl">{finding.title}</h2>
            <p className="report-lead mt-4">{finding.message}</p>
          </div>
        </div>

        <div className="detail-block grid gap-4 self-start p-5">
          <div className="kicker">At a glance</div>
          <dl className="grid gap-3">
            <div className="grid grid-cols-[110px_minmax(0,1fr)] gap-3 border-b-2 border-border pb-3">
              <dt className="text-sm text-slate-500">Object</dt>
              <dd className="data-mono text-[12px] text-slate-700">{finding.resource_name}</dd>
            </div>
            <div className="grid grid-cols-[110px_minmax(0,1fr)] gap-3 border-b-2 border-border pb-3">
              <dt className="text-sm text-slate-500">Confidence</dt>
              <dd className="metric-value">{finding.confidence_label}</dd>
            </div>
            <div className="grid grid-cols-[110px_minmax(0,1fr)] gap-3 border-b-2 border-border pb-3">
              <dt className="text-sm text-slate-500">Detected</dt>
              <dd className="metric-value">{formatTimestamp(finding.detected_at)}</dd>
            </div>
            <div className="grid grid-cols-[110px_minmax(0,1fr)] gap-3">
              <dt className="text-sm text-slate-500">Status</dt>
              <dd>
                <span
                  className={`inline-flex rounded-md border px-2.5 py-1 text-[12px] font-medium ${statusClasses(
                    finding.status,
                  )}`}
                >
                  {finding.status}
                </span>
              </dd>
            </div>
          </dl>
        </div>
      </div>

      <div className="flex flex-wrap gap-2 border-b-2 border-border py-4">
        {[
          { href: "#problem", label: "Problem" },
          { href: "#evidence", label: "Evidence" },
          { href: "#action", label: "Action" },
          { href: "#verification", label: "Verification" },
        ].map((item) => (
          <a
            key={item.href}
            href={item.href}
            className="rounded-md border-2 border-[#111111] bg-white px-3 py-2 text-[11px] font-semibold uppercase tracking-[0.12em] text-slate-700 shadow-[2px_2px_0_#111111] transition-all hover:-translate-x-[1px] hover:-translate-y-[1px] hover:bg-[#f4f1e6] hover:shadow-[3px_3px_0_#111111]"
          >
            {item.label}
          </a>
        ))}
      </div>

      <DetailSection id="problem" label="Summary" title="What is happening">
        <div className="summary-strip max-w-4xl px-5 py-4">
          <p className="text-[15px] leading-7 text-slate-700">{finding.problem_summary}</p>
        </div>
      </DetailSection>

      <div className="section-rule" />

      <DetailSection id="evidence" label="Evidence" title="Why Postgresome believes this">
        <div className="grid gap-4 lg:grid-cols-[minmax(0,1.6fr)_minmax(260px,0.9fr)]">
          <div className="space-y-4">
            <div className="max-w-4xl">
              <p className="text-[15px] leading-7 text-slate-700">{finding.evidence_summary}</p>

              {history ? (
                <div className="summary-strip mt-5 grid gap-0 sm:grid-cols-3">
                  <div className="p-4 sm:border-r-2 sm:border-border">
                    <div className="kicker">Current</div>
                      <div className="mt-2 font-heading text-[22px] font-semibold tracking-[-0.03em] text-foreground">
                        {formatNumber(history.current_value)}
                      </div>
                    </div>
                  <div className="p-4 sm:border-r-2 sm:border-border">
                    <div className="kicker">Baseline</div>
                      <div className="mt-2 font-heading text-[22px] font-semibold tracking-[-0.03em] text-foreground">
                        {formatNumber(history.baseline_value)}
                      </div>
                    </div>
                  <div className="p-4">
                    <div className="kicker">Change</div>
                      <div className="mt-2 font-heading text-[22px] font-semibold tracking-[-0.03em] text-foreground">
                        {formatPercent(history.change_percent)}
                      </div>
                    </div>
                </div>
              ) : null}
            </div>

            {evidence && evidence.length > 0 ? (
              <div className="overflow-hidden rounded-md border-2 border-[#111111] shadow-[2px_2px_0_#111111]">
                <div className="border-b-2 border-[#111111] bg-[#f4f1e6] px-4 py-3">
                  <div className="kicker">Supporting signals</div>
                </div>
                <div className="divide-y">
                  {evidence.map((item) => (
                    <div key={item.id} className="grid gap-3 px-4 py-3 sm:grid-cols-[1.2fr_1fr_140px]">
                      <div>
                        <div className="text-sm font-medium text-foreground">
                          {item.label}
                        </div>
                        <div className="mt-1 text-sm leading-6 text-slate-600">
                          {item.summary}
                        </div>
                      </div>
                      <div className="meta">
                        {item.metric_key || item.reference_id || item.evidence_type}
                      </div>
                      <div className="meta text-left sm:text-right">
                        {formatTimestamp(item.observed_at)}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            ) : null}

            {related_query ? (
              <div className="detail-block p-4">
                <div className="mb-3 flex items-center justify-between gap-3">
                  <div className="kicker">Related query</div>
                  <Link
                    to={`/queries/${encodeURIComponent(related_query.query_id)}`}
                    className="inline-flex items-center gap-2 text-[12px] font-medium text-slate-700 hover:text-slate-900"
                  >
                    Open in Query Explorer
                    <ArrowRight className="h-4 w-4" />
                  </Link>
                </div>
                <pre className="overflow-x-auto rounded-md border-2 border-[#111111] bg-slate-950 p-4 font-mono text-[12px] leading-6 text-white shadow-[2px_2px_0_#111111]">
                  {related_query.query}
                </pre>
              </div>
            ) : null}
          </div>

          <div className="space-y-4 xl:w-[400px]">
            <div className="detail-block p-4">
              <div className="kicker">Impact</div>
              <div className="mt-3 flex items-start gap-3">
                <ShieldAlert className="mt-0.5 h-4 w-4 text-slate-700" />
                <div>
                  <div className="text-sm font-semibold text-foreground">
                    {finding.primary_impact.label}
                  </div>
                  <p className="mt-1 text-sm leading-6 text-slate-600">
                    {finding.primary_impact.summary}
                  </p>
                </div>
              </div>
              {(finding.secondary_impacts || []).map((impact) => (
                <div key={impact.code} className="mt-4 border-t border-dashed pt-4">
                  <div className="text-sm font-medium text-foreground">{impact.label}</div>
                  <p className="mt-1 text-sm leading-6 text-slate-600">{impact.summary}</p>
                </div>
              ))}
            </div>

            {related_table ? (
              <div className="detail-block p-4">
                <div className="kicker">Affected table</div>
                <div className="mt-3 flex items-start gap-3">
                  <Database className="mt-0.5 h-4 w-4 text-slate-700" />
                  <div className="space-y-2">
                    <div className="text-sm font-semibold text-foreground">
                      {related_table.schema_name}.{related_table.table_name}
                    </div>
                    <div className="grid grid-cols-2 gap-3">
                      <div>
                        <div className="meta">Dead rows</div>
                        <div className="mt-1 text-sm font-medium text-foreground">
                          {formatNumber(related_table.dead_rows)}
                        </div>
                      </div>
                      <div>
                        <div className="meta">Sequential scans</div>
                        <div className="mt-1 text-sm font-medium text-foreground">
                          {formatNumber(related_table.sequential_scans)}
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            ) : null}

            {related_query ? (
              <div className="detail-block p-4">
                <div className="kicker">Query load</div>
                <div className="mt-3 flex items-start gap-3">
                  <Gauge className="mt-0.5 h-4 w-4 text-slate-700" />
                  <div className="grid flex-1 grid-cols-2 gap-3">
                    <div>
                      <div className="meta">Mean latency</div>
                      <div className="mt-1 text-sm font-medium text-foreground">
                        {formatDurationMs(related_query.mean_exec_time_ms)}
                      </div>
                    </div>
                    <div>
                      <div className="meta">Calls</div>
                      <div className="mt-1 text-sm font-medium text-foreground">
                        {formatNumber(related_query.calls)}
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            ) : null}
          </div>
        </div>
      </DetailSection>

      <div className="section-rule" />

      <DetailSection id="action" label="Recommended next step" title={finding.primary_action.label}>
        <div className="detail-block border-slate-950/90 p-5">
          <p className="max-w-3xl text-[15px] leading-7 text-slate-700">
            {finding.primary_action.summary}
          </p>

          {(finding.secondary_actions || []).length > 0 ? (
            <div className="mt-5 grid gap-3 sm:grid-cols-2">
              {finding.secondary_actions?.map((action) => (
                <div key={action.code} className="rounded-md border-2 border-[#111111] bg-white p-4 shadow-[2px_2px_0_#111111]">
                  <div className="text-sm font-semibold text-foreground">{action.label}</div>
                  <p className="mt-2 text-sm leading-6 text-slate-600">{action.summary}</p>
                </div>
              ))}
            </div>
          ) : null}

          <div className="mt-6 flex flex-wrap gap-3">
            {related_query ? (
              <Button asChild>
                <Link to={`/queries/${encodeURIComponent(related_query.query_id)}`}>
                  <FileSearch className="h-4 w-4" />
                  Investigate related query
                </Link>
              </Button>
            ) : null}
            <Button variant="outline">Mark for review</Button>
          </div>
        </div>
      </DetailSection>

      <div className="section-rule" />

      <DetailSection id="verification" label="Verification" title="How to confirm the fix">
        <div className="summary-strip grid gap-4 px-5 py-4 sm:grid-cols-[1fr_auto] sm:items-start">
          <div>
            <p className="text-[15px] leading-7 text-slate-700">
              {finding.verification_summary || finding.verification_hint}
            </p>
            <div className="mt-3 meta">
              Last seen {formatTimestamp(finding.last_seen_at)}
            </div>
          </div>
          <Badge variant="default">{finding.verification_status || "pending"}</Badge>
        </div>
      </DetailSection>
    </div>
  );
}
