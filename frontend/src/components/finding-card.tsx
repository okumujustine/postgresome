import { ArrowRight, TriangleAlert } from 'lucide-react';
import { Link } from 'react-router-dom';
import type { IssueQueueItem } from '../types/issues';
import { formatRelativeTime } from '../lib/format';
import { SeverityBadge, VerificationBadge } from './status-badges';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card';

export function FindingCard({ finding, linkTo = `/findings/${finding.id}` }: { finding: IssueQueueItem; linkTo?: string }) {
  return (
    <Card className="h-full">
      <CardHeader className="gap-3">
        <div className="flex flex-wrap items-center gap-2">
          <SeverityBadge severity={finding.severity} />
          <VerificationBadge status={finding.verification_status} />
        </div>
        <div className="space-y-2">
          <CardTitle className="text-base leading-6">{finding.problem_summary || finding.title}</CardTitle>
          <CardDescription className="leading-6">{finding.impact_summary || finding.message}</CardDescription>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="rounded-lg border border-[var(--border)] bg-[var(--muted)] p-3">
          <div className="mb-2 flex items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.08em] text-[var(--muted-foreground)]">
            <TriangleAlert size={13} />
            Evidence
          </div>
          <p className="text-sm leading-6 text-[var(--foreground)]">{finding.evidence_summary}</p>
        </div>
        <div className="space-y-2 text-sm text-[var(--muted-foreground)]">
          <div>
            <span className="font-medium text-[var(--foreground)]">Affected object:</span> {finding.resource_type} · {finding.resource_name}
          </div>
          <div>
            <span className="font-medium text-[var(--foreground)]">Recommendation:</span> {finding.suggested_action || finding.recommendation}
          </div>
          <div>Last seen {formatRelativeTime(finding.last_seen_at)}</div>
        </div>
        <Link
          to={linkTo}
          className="inline-flex h-9 w-full items-center justify-between rounded-md border border-[var(--border)] bg-[var(--panel)] px-4 text-sm font-medium text-[var(--foreground)] no-underline transition-colors hover:bg-[var(--muted)]"
        >
          Open diagnosis
          <ArrowRight size={14} />
        </Link>
      </CardContent>
    </Card>
  );
}
