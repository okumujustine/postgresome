import { useCallback, useEffect, useState } from 'react';
import { Activity, Database, RotateCcw, Search } from 'lucide-react';
import { getDashboardOverview } from '../api/dashboard';
import { queryMetrics } from '../api/metrics';
import { ApiError } from '../api/client';
import type { DashboardOverview, MetricQueryResponse, MetricRange } from '../types/dashboard';
import { Layout } from '../components/Layout';
import { Card } from '../components/Card';
import { MetricCard } from '../components/MetricCard';
import { LineChart } from '../components/LineChart';

function formatMetricValue(value: number, unit: string): string {
  if (unit === 'percent') {
    return value.toFixed(1);
  }
  return Math.round(value).toString();
}

function metricUnitLabel(unit: string): string {
  return unit === 'percent' ? '%' : '';
}

const CHART_METRICS: { key: string; label: string; color: string }[] = [
  { key: 'active_connections', label: 'Active connections', color: 'var(--viz-1)' },
  { key: 'transaction_commits', label: 'Transaction commits', color: 'var(--viz-2)' },
  { key: 'transaction_rollbacks', label: 'Transaction rollbacks', color: 'var(--viz-5)' },
  { key: 'blocks_hit_in_cache', label: 'Blocks hit in cache', color: 'var(--viz-6)' },
  { key: 'blocks_read_from_disk', label: 'Blocks read from disk', color: 'var(--viz-3)' },
];

export function MetricsPage() {
  const [range, setRange] = useState<MetricRange>('1h');
  const [overview, setOverview] = useState<DashboardOverview | null>(null);
  const [charts, setCharts] = useState<Record<string, MetricQueryResponse>>({});
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async (currentRange: MetricRange, isRefresh: boolean) => {
    if (isRefresh) {
      setRefreshing(true);
    }

    try {
      const [overviewResult, ...chartResults] = await Promise.all([
        getDashboardOverview({ range: currentRange }),
        ...CHART_METRICS.map((m) => queryMetrics({ metricKey: m.key, range: currentRange })),
      ]);

      const chartMap: Record<string, MetricQueryResponse> = {};
      CHART_METRICS.forEach((m, i) => {
        chartMap[m.key] = chartResults[i];
      });

      setOverview(overviewResult);
      setCharts(chartMap);
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
  const instance = overview?.database_instance ?? { id: '', database_name: '', host: '', status: 'unknown' };
  const hasInstance = Boolean(overview?.database_instance?.id);

  return (
    <Layout
      title="Metrics"
      databaseName={hasInstance ? instance.database_name : undefined}
      status={instance.status}
      range={range}
      onRangeChange={setRange}
      onRefresh={() => load(range, true)}
      refreshing={refreshing}
    >
      {error && (
        <div
          className="mb-6 rounded-[var(--radius-lg)] border px-4 py-3 text-sm"
          style={{ borderColor: 'rgba(207,34,46,0.25)', background: 'var(--danger-tint)', color: 'var(--danger)' }}
        >
          {error}
        </div>
      )}

      {loading && !overview ? (
        <div className="flex items-center gap-2 text-sm" style={{ color: 'var(--text-muted)' }}>
          <Search size={14} className="animate-pulse" />
          Loading metrics…
        </div>
      ) : overview ? (
        <>
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

          <div className="grid gap-[14px]" style={{ gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))' }}>
            {CHART_METRICS.map((m) => (
              <Card key={m.key} title={m.label}>
                <LineChart data={charts[m.key]?.points ?? []} color={m.color} />
              </Card>
            ))}
          </div>
        </>
      ) : null}
    </Layout>
  );
}
