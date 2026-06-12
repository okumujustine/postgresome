import { useCallback, useEffect, useState } from 'react';
import { Activity, Database, RotateCcw, Search, Server } from 'lucide-react';
import { getDashboardOverview } from '../api/dashboard';
import { ApiError } from '../api/client';
import type { DashboardInstance, DashboardOverview, MetricRange } from '../types/dashboard';
import { Layout } from '../components/Layout';
import { MetricCard } from '../components/MetricCard';
import { FindingsList } from '../components/FindingsList';

function formatMetricValue(value: number, unit: string): string {
  if (unit === 'percent') {
    return value.toFixed(1);
  }
  return Math.round(value).toString();
}

function metricUnitLabel(unit: string): string {
  return unit === 'percent' ? '%' : '';
}

function InstanceHeader({ instance }: { instance: DashboardInstance }) {
  const hasInstance = Boolean(instance.id);

  return (
    <div className="mb-[22px] flex flex-wrap items-start justify-between gap-4">
      <div>
        <div className="mb-[9px] flex items-center gap-3">
          <h2
            className="m-0 text-[var(--fs-h1)] font-semibold"
            style={{ color: 'var(--text-primary)', letterSpacing: 'var(--ls-tight)', fontFamily: 'var(--font-mono)' }}
          >
            {hasInstance ? instance.database_name : 'No database instance'}
          </h2>
        </div>
        {hasInstance && (
          <div className="flex flex-wrap items-center gap-4">
            <span className="inline-flex items-center gap-[6px] text-[12.5px]" style={{ color: 'var(--text-muted)' }}>
              <Server size={13} />
              <span style={{ fontFamily: 'var(--font-mono)' }}>{instance.host}</span>
            </span>
            <span className="inline-flex items-center gap-[6px] text-[12.5px]" style={{ color: 'var(--text-muted)' }}>
              <Database size={13} />
              <span style={{ fontFamily: 'var(--font-mono)' }}>{instance.id}</span>
            </span>
          </div>
        )}
      </div>
    </div>
  );
}

export function DashboardPage() {
  const [range, setRange] = useState<MetricRange>('1h');
  const [overview, setOverview] = useState<DashboardOverview | null>(null);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async (currentRange: MetricRange, isRefresh: boolean) => {
    if (isRefresh) {
      setRefreshing(true);
    }

    try {
      const data = await getDashboardOverview({ range: currentRange });
      setOverview(data);
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
    // load() only updates state after its internal await, not synchronously.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    load(range, false);
  }, [range, load]);

  const loading = overview === null && error === null;

  const summary = overview?.summary;
  const findings = overview?.findings;
  const instance = overview?.database_instance ?? { id: '', database_name: '', host: '', status: 'unknown' };

  return (
    <Layout
      title="Overview"
      databaseName={overview?.database_instance?.id ? instance.database_name : undefined}
      status={instance.status}
      range={range}
      onRangeChange={setRange}
      onRefresh={() => load(range, true)}
      refreshing={refreshing}
    >
      {error && (
        <div
          className="mb-6 rounded-[var(--radius-lg)] border px-4 py-3 text-sm"
          style={{ borderColor: 'rgba(248,81,73,0.32)', background: 'var(--danger-tint)', color: '#FF9892' }}
        >
          {error}
        </div>
      )}

      {loading && !overview ? (
        <div className="flex items-center gap-2 text-sm" style={{ color: 'var(--text-muted)' }}>
          <Search size={14} className="animate-pulse" />
          Loading dashboard…
        </div>
      ) : overview ? (
        <>
          <InstanceHeader instance={instance} />

          <div className="mb-6 grid gap-[14px]" style={{ gridTemplateColumns: 'repeat(auto-fit, minmax(210px, 1fr))' }}>
            <MetricCard
              label="Active connections"
              value={formatMetricValue(summary!.active_connections.value, summary!.active_connections.unit)}
              unit={metricUnitLabel(summary!.active_connections.unit)}
              trendPercent={summary!.active_connections.trend_percent}
              invertTrend
              icon={<Activity size={14} />}
            />
            <MetricCard
              label="Cache hit ratio"
              value={formatMetricValue(summary!.cache_hit_ratio.value, summary!.cache_hit_ratio.unit)}
              unit={metricUnitLabel(summary!.cache_hit_ratio.unit)}
              trendPercent={summary!.cache_hit_ratio.trend_percent}
              icon={<Database size={14} />}
            />
            <MetricCard
              label="Rollback rate"
              value={formatMetricValue(summary!.rollback_rate.value, summary!.rollback_rate.unit)}
              unit={metricUnitLabel(summary!.rollback_rate.unit)}
              trendPercent={summary!.rollback_rate.trend_percent}
              invertTrend
              icon={<RotateCcw size={14} />}
            />
            <MetricCard
              label="Slow queries"
              value={formatMetricValue(summary!.slow_queries.value, summary!.slow_queries.unit)}
              unit={metricUnitLabel(summary!.slow_queries.unit)}
              trendPercent={summary!.slow_queries.trend_percent}
              invertTrend
              icon={<Search size={14} />}
            />
          </div>

          <FindingsList findings={findings!} />
        </>
      ) : null}
    </Layout>
  );
}
