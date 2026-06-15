import { useEffect, useMemo, useState } from 'react';
import { AlertTriangle, ArrowUpRight } from 'lucide-react';
import { useNavigate, useParams } from 'react-router-dom';
import { getFinding, listFindings } from '../api/findings';
import { ApiError } from '../api/client';
import { AppShell } from '../components/app-shell';
import { FindingDetail } from '../components/finding-detail';
import { SeverityBadge, VerificationBadge } from '../components/status-badges';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { useDatabaseInstance } from '../lib/databaseInstance';
import type { IssueDetailResponse, IssueQueueItem } from '../types/issues';

export function DiagnosisPage() {
  const navigate = useNavigate();
  const params = useParams<{ id?: string }>();
  const { selectedId, loading: instanceLoading } = useDatabaseInstance();
  const [findings, setFindings] = useState<IssueQueueItem[]>([]);
  const [selectedFinding, setSelectedFinding] = useState<IssueDetailResponse | null>(null);
  const [query, setQuery] = useState('');
  const [status, setStatus] = useState('open');
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (instanceLoading || !selectedId) return;

    listFindings({ databaseInstanceId: selectedId, status, range: '7d', limit: 50 })
      .then((result) => {
        setFindings(result.findings);
        setError(null);
      })
      .catch((err) => {
        const message =
          err instanceof ApiError
            ? `The Postgresome API returned an error (${err.status}).`
            : 'Unable to reach the Postgresome API. Is it running?';
        setError(message);
      });
  }, [selectedId, instanceLoading, status]);

  useEffect(() => {
    if (!selectedId) return;
    const nextId = params.id ?? findings[0]?.id;
    if (!nextId) return;

    getFinding(nextId, selectedId)
      .then(setSelectedFinding)
      .catch(() => setSelectedFinding(null));
  }, [selectedId, findings, params.id]);

  const filtered = useMemo(() => {
    const needle = query.trim().toLowerCase();
    if (!needle) return findings;
    return findings.filter((finding) =>
      `${finding.problem_summary} ${finding.resource_name} ${finding.category}`.toLowerCase().includes(needle),
    );
  }, [findings, query]);

  return (
    <AppShell title="Findings" subtitle="The main product experience. Review one database problem at a time and move directly to the next action.">
      <div className="space-y-6">
        {error ? <div className="rounded-xl border border-[var(--danger)] bg-[var(--danger-soft)] px-4 py-3 text-[13px] text-[var(--danger)]">{error}</div> : null}

        <div className="flex flex-wrap items-end justify-between gap-4">
          <div>
            <h2 className="text-[16px] font-semibold text-[var(--foreground)]">Issue summary</h2>
            <p className="mt-1 max-w-2xl text-[14px] leading-6 text-[var(--muted-foreground)]">
              Start from the problem, read the explanation, validate the evidence, then act. The queue keeps the highest-signal diagnoses at the top.
            </p>
          </div>
          <div className="flex items-center gap-2 rounded-[10px] border border-[var(--border)] bg-[var(--muted)] p-1">
            <Button variant={status === 'open' ? 'outline' : 'ghost'} size="sm" onClick={() => setStatus('open')}>
              Open
            </Button>
            <Button variant={status === 'resolved' ? 'outline' : 'ghost'} size="sm" onClick={() => setStatus('resolved')}>
              Resolved
            </Button>
          </div>
        </div>

        <div className="grid gap-6 xl:grid-cols-[320px_minmax(0,1fr)]">
          <div className="rounded-xl border border-[var(--border)] bg-[var(--panel)]">
            <div className="border-b border-[var(--border)] px-4 py-4">
              <Input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Search diagnosis queue" />
            </div>
            <div className="max-h-[72vh] overflow-y-auto p-2">
              {filtered.map((finding) => {
                const active = selectedFinding?.finding.id === finding.id;
                return (
                  <button
                    key={finding.id}
                    onClick={() => navigate(`/findings/${finding.id}`)}
                    className={`mb-2 w-full rounded-xl border p-4 text-left transition-colors ${
                      active ? 'border-[var(--foreground)] bg-[var(--muted)]' : 'border-transparent bg-[var(--panel)] hover:border-[var(--border)] hover:bg-[var(--muted)]'
                    }`}
                  >
                    <div className="flex flex-wrap items-center gap-2">
                      <SeverityBadge severity={finding.severity} />
                      <VerificationBadge status={finding.verification_status} />
                    </div>
                    <div className="mt-3 text-[14px] font-semibold leading-6 text-[var(--foreground)]">{finding.problem_summary || finding.title}</div>
                    <div className="mt-1 text-[13px] leading-6 text-[var(--muted-foreground)]">{finding.impact_summary || finding.evidence_summary}</div>
                    <div className="mt-3 inline-flex items-center gap-2 text-[12px] text-[var(--muted-foreground)]">
                      {finding.resource_type} · {finding.resource_name}
                      <ArrowUpRight size={12} />
                    </div>
                  </button>
                );
              })}
            </div>
          </div>

          <div>
            {selectedFinding ? (
              <FindingDetail detail={selectedFinding} />
            ) : (
              <div className="rounded-xl border border-dashed border-[var(--border)] bg-[var(--muted)] px-6 py-10 text-center">
                <AlertTriangle className="mx-auto mb-3 text-[var(--muted-foreground)]" size={20} />
                <div className="text-[14px] text-[var(--muted-foreground)]">Choose a finding from the queue to inspect the evidence and recommended next step.</div>
              </div>
            )}
          </div>
        </div>
      </div>
    </AppShell>
  );
}
