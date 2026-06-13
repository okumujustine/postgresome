import { useCallback, useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { ArrowLeft, Check, Copy, Search, Sparkles } from 'lucide-react';
import { listFindings } from '../api/findings';
import { ApiError } from '../api/client';
import type { DashboardFinding, MetricRange } from '../types/dashboard';
import { Layout } from '../components/Layout';
import { Card } from '../components/Card';
import { SeverityPill } from '../components/SeverityPill';
import { formatRelativeTime } from '../lib/format';
import { useDatabaseInstance } from '../lib/databaseInstance';

const MAX_FINDINGS_LIMIT = 100;
const ISSUE_DETAIL_RANGE: MetricRange = '7d';

function StatusPill({ status }: { status: string }) {
  const resolved = status === 'resolved';
  return (
    <span
      className="inline-flex h-[22px] items-center rounded-[var(--radius-pill)] px-[9px] text-xs font-medium capitalize"
      style={
        resolved
          ? { background: 'var(--success-tint)', color: 'var(--success)', border: '1px solid rgba(26,127,55,0.25)', letterSpacing: 'var(--ls-snug)' }
          : { background: 'var(--blue-tint)', color: 'var(--blue-600)', border: '1px solid rgba(41,98,224,0.25)', letterSpacing: 'var(--ls-snug)' }
      }
    >
      {resolved ? 'Resolved' : 'Open'}
    </span>
  );
}

function SectionLabel({ children }: { children: string }) {
  return (
    <div className="mb-2 text-xs font-semibold uppercase" style={{ color: 'var(--text-faint)', letterSpacing: 'var(--ls-label)' }}>
      {children}
    </div>
  );
}

export function IssueDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [findings, setFindings] = useState<DashboardFinding[] | null>(null);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const { selectedId, loading: instanceLoading } = useDatabaseInstance();

  const load = useCallback(async (databaseInstanceId: string, isRefresh: boolean) => {
    if (isRefresh) {
      setRefreshing(true);
    }

    try {
      const result = await listFindings({
        status: 'all',
        range: ISSUE_DETAIL_RANGE,
        limit: MAX_FINDINGS_LIMIT,
        databaseInstanceId,
      });
      setFindings(result.findings);
      setError(null);
    } catch (err) {
      const message =
        err instanceof ApiError
          ? `The Postgresome API returned an error (${err.status}).`
          : 'Unable to reach the Postgresome API. Is it running?';
      setError(message);
    } finally {
      setRefreshing(false);
    }
  }, []);

  useEffect(() => {
    if (instanceLoading || !selectedId) return;
    // load() only updates state after its internal await, not synchronously.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    load(selectedId, false);
  }, [selectedId, instanceLoading, load]);

  const loading = findings === null && error === null;
  const finding = findings?.find((f) => f.id === id) ?? null;

  const handleCopySummary = useCallback(() => {
    if (!finding) return;

    const lines = [
      finding.title,
      '',
      `Problem: ${finding.message}`,
      '',
      `Evidence: current value ${finding.current_value.toLocaleString()}, threshold ${finding.threshold_value.toLocaleString()} (rule: ${finding.rule_key})`,
    ];
    if (finding.recommendation) {
      lines.push('', `Recommendation: ${finding.recommendation}`);
    }

    void navigator.clipboard.writeText(lines.join('\n')).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    });
  }, [finding]);

  return (
    <Layout title="Issue" onRefresh={() => selectedId && load(selectedId, true)} refreshing={refreshing}>
      {error && (
        <div
          className="mb-6 rounded-[var(--radius-lg)] border px-4 py-3 text-sm"
          style={{ borderColor: 'rgba(207,34,46,0.25)', background: 'var(--danger-tint)', color: 'var(--danger)' }}
        >
          {error}
        </div>
      )}

      <Link
        to="/issues"
        className="mb-4 inline-flex items-center gap-[6px] text-[13px] font-medium no-underline"
        style={{ color: 'var(--text-secondary)' }}
      >
        <ArrowLeft size={14} />
        Back to Issues
      </Link>

      {loading && !findings ? (
        <div className="flex items-center gap-2 text-sm" style={{ color: 'var(--text-muted)' }}>
          <Search size={14} className="animate-pulse" />
          Loading issue…
        </div>
      ) : !finding ? (
        <Card title="Issue not found">
          <div className="text-sm" style={{ color: 'var(--text-muted)' }}>
            We couldn&apos;t find this issue. It may have been resolved long enough ago to fall outside the lookback
            window, or it no longer exists.{' '}
            <Link to="/issues" style={{ color: 'var(--text-link)' }}>
              Back to Issues
            </Link>
          </div>
        </Card>
      ) : (
        <div className="flex flex-col gap-5">
          <Card>
            <div className="flex flex-col gap-3">
              <div className="flex flex-wrap items-center gap-2">
                <StatusPill status={finding.status} />
                <SeverityPill severity={finding.severity} />
                <span className="text-xs" style={{ fontFamily: 'var(--font-mono)', color: 'var(--text-muted)' }}>
                  {finding.resource_type}: {finding.resource_name}
                </span>
              </div>
              <h1 className="m-0 text-[var(--fs-h1)] font-semibold" style={{ color: 'var(--text-primary)', letterSpacing: 'var(--ls-tight)' }}>
                {finding.title}
              </h1>
            </div>
          </Card>

          <Card title="Problem">
            <p className="m-0 text-[13.5px]" style={{ color: 'var(--text-secondary)' }}>
              {finding.message}
            </p>
          </Card>

          <Card title="Evidence">
            <div className="grid gap-4" style={{ gridTemplateColumns: 'repeat(auto-fit, minmax(160px, 1fr))' }}>
              <div>
                <SectionLabel>Current value</SectionLabel>
                <div className="tabular text-[var(--fs-h3)] font-semibold" style={{ color: 'var(--text-primary)' }}>
                  {finding.current_value.toLocaleString()}
                </div>
              </div>
              <div>
                <SectionLabel>Threshold</SectionLabel>
                <div className="tabular text-[var(--fs-h3)] font-semibold" style={{ color: 'var(--text-primary)' }}>
                  {finding.threshold_value.toLocaleString()}
                </div>
              </div>
              <div>
                <SectionLabel>Rule</SectionLabel>
                <div className="text-[13px]" style={{ fontFamily: 'var(--font-mono)', color: 'var(--text-secondary)' }}>
                  {finding.rule_key}
                </div>
              </div>
            </div>
          </Card>

          <Card title="Timeline">
            <div className="grid gap-4" style={{ gridTemplateColumns: 'repeat(auto-fit, minmax(160px, 1fr))' }}>
              <div>
                <SectionLabel>First detected</SectionLabel>
                <div className="text-[13px]" style={{ color: 'var(--text-secondary)' }} title={new Date(finding.first_seen_at).toLocaleString()}>
                  {formatRelativeTime(finding.first_seen_at)}
                </div>
              </div>
              <div>
                <SectionLabel>Last detected</SectionLabel>
                <div className="text-[13px]" style={{ color: 'var(--text-secondary)' }} title={new Date(finding.last_seen_at).toLocaleString()}>
                  {formatRelativeTime(finding.last_seen_at)}
                </div>
              </div>
              <div>
                <SectionLabel>Occurrences</SectionLabel>
                <div className="tabular text-[13px]" style={{ color: 'var(--text-secondary)' }}>
                  {finding.occurrence_count}
                </div>
              </div>
            </div>
          </Card>

          <Card title="Recommendation">
            <p className="m-0 text-[13.5px]" style={{ color: 'var(--text-secondary)' }}>
              {finding.recommendation || 'No recommendation available.'}
            </p>
          </Card>

          <Card title="AI recommendation">
            <div className="flex items-center gap-2 text-[13.5px]" style={{ color: 'var(--text-muted)' }}>
              <Sparkles size={15} />
              AI-powered recommendations are coming soon.
            </div>
          </Card>

          <div className="flex flex-wrap items-center gap-2">
            <button
              onClick={handleCopySummary}
              className="inline-flex h-[var(--control-h-md)] cursor-pointer items-center gap-[6px] rounded-[var(--radius-md)] border px-4 text-[13px] font-medium"
              style={{ background: 'var(--surface-raised)', color: 'var(--text-primary)', borderColor: 'var(--border-default)' }}
            >
              {copied ? <Check size={14} /> : <Copy size={14} />}
              {copied ? 'Copied' : 'Copy summary'}
            </button>
            <button
              disabled
              title="Coming soon"
              className="inline-flex h-[var(--control-h-md)] cursor-not-allowed items-center gap-[6px] rounded-[var(--radius-md)] border px-4 text-[13px] font-medium opacity-40"
              style={{ background: 'var(--surface-raised)', color: 'var(--text-secondary)', borderColor: 'var(--border-default)' }}
            >
              Ignore
            </button>
            <button
              disabled
              title="Coming soon"
              className="inline-flex h-[var(--control-h-md)] cursor-not-allowed items-center gap-[6px] rounded-[var(--radius-md)] border px-4 text-[13px] font-medium opacity-40"
              style={{ background: 'var(--surface-raised)', color: 'var(--text-secondary)', borderColor: 'var(--border-default)' }}
            >
              Mark resolved
            </button>
          </div>
        </div>
      )}
    </Layout>
  );
}
