import { useEffect, useState } from 'react';
import { queryMetric } from '../api/metrics';
import { ApiError } from '../api/client';
import { AppShell } from '../components/app-shell';
import { MetricSparkline } from '../components/metric-sparkline';
import { Accordion, AccordionItem } from '../components/ui/accordion';
import { DetailCard } from '../components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../components/ui/tabs';
import { useDatabaseInstance } from '../lib/databaseInstance';
import type { MetricQueryResponse } from '../types/dashboard';

const METRICS = [
  { key: 'active_connections', label: 'Active connections' },
  { key: 'blocks_hit_in_cache', label: 'Cache hits' },
  { key: 'blocks_read_from_disk', label: 'Disk reads' },
];

export function MetricsPage() {
  const { selectedId, loading: instanceLoading } = useDatabaseInstance();
  const [metrics, setMetrics] = useState<Record<string, MetricQueryResponse>>({});
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (instanceLoading || !selectedId) return;

    Promise.all(METRICS.map((metric) => queryMetric(metric.key, selectedId, '24h')))
      .then((responses) => {
        const next: Record<string, MetricQueryResponse> = {};
        responses.forEach((response) => {
          next[response.metric_key] = response;
        });
        setMetrics(next);
        setError(null);
      })
      .catch((err) => {
        const message =
          err instanceof ApiError
            ? `The Postgresome API returned an error (${err.status}).`
            : 'Unable to reach the Postgresome API. Is it running?';
        setError(message);
      });
  }, [selectedId, instanceLoading]);

  return (
    <AppShell title="Raw Metrics" subtitle="Advanced evidence only. Diagnosis remains the primary interface.">
      <div className="space-y-6">
        {error ? <div className="rounded-lg border border-[rgba(201,55,44,0.18)] bg-[rgba(201,55,44,0.08)] px-4 py-3 text-sm text-[var(--danger)]">{error}</div> : null}

        <DetailCard title="Advanced evidence desk" description="Use this area only when you need raw time-series support for a diagnosis or to verify recovery after a change.">
          <Tabs defaultValue="signals">
            <TabsList>
              <TabsTrigger value="signals">Signals</TabsTrigger>
              <TabsTrigger value="explain">How to use</TabsTrigger>
            </TabsList>

            <TabsContent value="signals" className="mt-4">
              <div className="grid gap-4 lg:grid-cols-3">
                {METRICS.map((metric) => (
                  <div key={metric.key} className="rounded-lg border border-[var(--border)] bg-[var(--muted)] p-4">
                    <div className="mb-3 text-sm font-medium text-[var(--foreground)]">{metric.label}</div>
                    <MetricSparkline points={metrics[metric.key]?.points ?? []} />
                  </div>
                ))}
              </div>
            </TabsContent>

            <TabsContent value="explain" className="mt-4">
              <Accordion>
                <AccordionItem title="When to use raw metrics" subtitle="Only when the diagnosis detail needs supporting proof.">
                  Use raw metrics when you are validating trend direction, checking whether a fix improved a signal, or comparing current behavior against historical memory.
                </AccordionItem>
                <AccordionItem title="When not to use them" subtitle="Do not begin your investigation here.">
                  Do not start from charts. Start from the diagnosis queue, then come here only if the issue detail points you to a signal that needs a deeper look.
                </AccordionItem>
              </Accordion>
            </TabsContent>
          </Tabs>
        </DetailCard>
      </div>
    </AppShell>
  );
}
