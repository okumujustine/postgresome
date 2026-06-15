import type { MetricQueryPoint } from '../types/dashboard';

export function MetricSparkline({ points }: { points: MetricQueryPoint[] }) {
  if (points.length === 0) {
    return <div className="h-16 rounded-lg bg-[var(--muted)]" />;
  }

  const width = 240;
  const height = 64;
  const values = points.map((point) => point.value);
  const min = Math.min(...values);
  const max = Math.max(...values);
  const span = max - min || 1;

  const polyline = points
    .map((point, index) => {
      const x = (index / Math.max(1, points.length - 1)) * width;
      const y = height - ((point.value - min) / span) * height;
      return `${x},${y}`;
    })
    .join(' ');

  return (
    <svg viewBox={`0 0 ${width} ${height}`} width="100%" height={height} preserveAspectRatio="none" className="rounded-lg bg-[var(--muted)]">
      <polyline fill="none" stroke="var(--primary)" strokeWidth="2" points={polyline} />
    </svg>
  );
}

