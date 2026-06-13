import { useCallback, useEffect, useState } from 'react';
import { ChevronDown, ChevronLeft, ChevronRight, Search } from 'lucide-react';
import { listFindings } from '../api/findings';
import { ApiError } from '../api/client';
import type { FindingsListResponse } from '../types/dashboard';
import { Layout } from '../components/Layout';
import { Card } from '../components/Card';
import { IssueRow } from '../components/IssueRow';
import { useDatabaseInstance } from '../lib/databaseInstance';

const PAGE_SIZE = 20;

// Issues are an open-ended tracker, not a time-series view — use the widest
// available range so open/resolved issues aren't hidden by a short window.
const ISSUES_RANGE = '7d';

const SEVERITY_OPTIONS = [
  { value: '', label: 'All severities' },
  { value: 'critical', label: 'Critical' },
  { value: 'warning', label: 'Warning' },
  { value: 'info', label: 'Info' },
];

const CATEGORY_OPTIONS = [
  { value: '', label: 'All categories' },
  { value: 'queries', label: 'Queries' },
  { value: 'connections', label: 'Connections' },
  { value: 'vacuum', label: 'Vacuum' },
  { value: 'locks', label: 'Locks' },
  { value: 'transactions', label: 'Transactions' },
  { value: 'cache', label: 'Cache' },
  { value: 'query_plan', label: 'Query Plan' },
];

const STATUS_TABS: { value: string; label: string }[] = [
  { value: 'open', label: 'Open' },
  { value: 'resolved', label: 'Resolved' },
];

function FilterSelect({
  value,
  onChange,
  options,
}: {
  value: string;
  onChange: (value: string) => void;
  options: { value: string; label: string }[];
}) {
  return (
    <div className="relative inline-block">
      <select
        value={value}
        onChange={(event) => onChange(event.target.value)}
        className="h-[var(--control-h-sm)] cursor-pointer appearance-none rounded-[var(--radius-md)] border pr-8 pl-3 text-[13px] outline-none"
        style={{ background: 'var(--surface-raised)', color: 'var(--text-primary)', borderColor: 'var(--border-default)', fontFamily: 'var(--font-sans)' }}
      >
        {options.map((option) => (
          <option key={option.value} value={option.value} style={{ background: 'var(--surface-card)', color: 'var(--text-primary)' }}>
            {option.label}
          </option>
        ))}
      </select>
      <span className="pointer-events-none absolute right-[11px] top-1/2 -translate-y-1/2" style={{ color: 'var(--text-muted)' }}>
        <ChevronDown size={14} />
      </span>
    </div>
  );
}

export function IssuesPage() {
  const [severity, setSeverity] = useState('');
  const [category, setCategory] = useState('');
  const [status, setStatus] = useState('open');
  const [offset, setOffset] = useState(0);
  const [data, setData] = useState<FindingsListResponse | null>(null);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { selectedId, loading: instanceLoading } = useDatabaseInstance();

  const load = useCallback(
    async (
      currentSeverity: string,
      currentCategory: string,
      currentStatus: string,
      currentOffset: number,
      databaseInstanceId: string,
      isRefresh: boolean,
    ) => {
      if (isRefresh) {
        setRefreshing(true);
      }

      try {
        const result = await listFindings({
          range: ISSUES_RANGE,
          severity: currentSeverity || undefined,
          category: currentCategory || undefined,
          status: currentStatus,
          limit: PAGE_SIZE,
          offset: currentOffset,
          databaseInstanceId,
        });
        setData(result);
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
    },
    [],
  );

  useEffect(() => {
    if (instanceLoading || !selectedId) return;
    // load() only updates state after its internal await, not synchronously.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    load(severity, category, status, offset, selectedId, false);
  }, [severity, category, status, offset, selectedId, instanceLoading, load]);

  const loading = data === null && error === null;

  const counts = data?.severity_counts ?? { critical: 0, warning: 0, info: 0 };
  const findings = data?.findings ?? [];
  const total = data?.total ?? 0;

  const rangeStart = total === 0 ? 0 : offset + 1;
  const rangeEnd = Math.min(offset + PAGE_SIZE, total);

  return (
    <Layout title="Issues" onRefresh={() => load(severity, category, status, offset, selectedId, true)} refreshing={refreshing}>
      {error && (
        <div
          className="mb-6 rounded-[var(--radius-lg)] border px-4 py-3 text-sm"
          style={{ borderColor: 'rgba(207,34,46,0.25)', background: 'var(--danger-tint)', color: 'var(--danger)' }}
        >
          {error}
        </div>
      )}

      <div className="mb-4 flex flex-wrap items-center justify-between gap-3 border-b" style={{ borderColor: 'var(--border-subtle)' }}>
        <div className="flex items-center gap-1">
          {STATUS_TABS.map((tab) => (
            <button
              key={tab.value}
              onClick={() => {
                setStatus(tab.value);
                setOffset(0);
              }}
              className="relative cursor-pointer border-0 bg-transparent px-3 py-2 text-[13.5px] font-semibold"
              style={{ color: status === tab.value ? 'var(--text-primary)' : 'var(--text-muted)' }}
            >
              {tab.label}
              {status === tab.value && (
                <span className="absolute bottom-[-1px] left-0 right-0 h-[2px] rounded-[1px]" style={{ background: 'var(--accent)' }} />
              )}
            </button>
          ))}
        </div>

        <div className="mb-2 flex flex-wrap items-center gap-[10px]">
          <FilterSelect
            value={severity}
            onChange={(value) => {
              setSeverity(value);
              setOffset(0);
            }}
            options={SEVERITY_OPTIONS}
          />
          <FilterSelect
            value={category}
            onChange={(value) => {
              setCategory(value);
              setOffset(0);
            }}
            options={CATEGORY_OPTIONS}
          />
        </div>
      </div>

      {loading && !data ? (
        <div className="flex items-center gap-2 text-sm" style={{ color: 'var(--text-muted)' }}>
          <Search size={14} className="animate-pulse" />
          Loading issues…
        </div>
      ) : data ? (
        <Card
          title={status === 'open' ? 'Open issues' : 'Resolved issues'}
          subtitle={`${total} issue${total === 1 ? '' : 's'} · ${counts.critical} critical · ${counts.warning} warning · ${counts.info} info`}
        >
          {findings.length === 0 ? (
            <div className="text-sm" style={{ color: 'var(--text-muted)' }}>
              {status === 'open' ? 'No open issues match these filters — nice work.' : 'No resolved issues match these filters.'}
            </div>
          ) : (
            <>
              <div className="flex flex-col">
                {findings.map((finding, index) => (
                  <div key={finding.id} style={{ borderTop: index === 0 ? 'none' : '1px solid var(--border-subtle)' }}>
                    <IssueRow finding={finding} />
                  </div>
                ))}
              </div>

              <div
                className="mt-4 flex items-center justify-between gap-3 border-t pt-4"
                style={{ borderColor: 'var(--border-subtle)' }}
              >
                <span className="text-xs" style={{ color: 'var(--text-muted)' }}>
                  Showing {rangeStart}–{rangeEnd} of {total}
                </span>
                <div className="flex items-center gap-2">
                  <button
                    onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}
                    disabled={offset === 0}
                    className="inline-flex h-[var(--control-h-sm)] items-center gap-1 rounded-[var(--radius-md)] border px-3 text-[13px] disabled:cursor-not-allowed disabled:opacity-40"
                    style={{ background: 'var(--surface-raised)', color: 'var(--text-secondary)', borderColor: 'var(--border-default)' }}
                  >
                    <ChevronLeft size={14} />
                    Previous
                  </button>
                  <button
                    onClick={() => setOffset(offset + PAGE_SIZE)}
                    disabled={offset + PAGE_SIZE >= total}
                    className="inline-flex h-[var(--control-h-sm)] items-center gap-1 rounded-[var(--radius-md)] border px-3 text-[13px] disabled:cursor-not-allowed disabled:opacity-40"
                    style={{ background: 'var(--surface-raised)', color: 'var(--text-secondary)', borderColor: 'var(--border-default)' }}
                  >
                    Next
                    <ChevronRight size={14} />
                  </button>
                </div>
              </div>
            </>
          )}
        </Card>
      ) : null}
    </Layout>
  );
}
