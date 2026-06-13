import type { MetricQueryPoint } from '../types/dashboard';

interface LineChartProps {
  data: MetricQueryPoint[];
  color?: string;
  height?: number;
}

const WIDTH = 480;

export function LineChart({ data, color, height = 120 }: LineChartProps) {
  if (data.length === 0) {
    return (
      <div className="flex items-center justify-center text-xs" style={{ height, color: 'var(--text-faint)' }}>
        No data for this range.
      </div>
    );
  }

  const values = data.map((point) => point.value);
  const min = Math.min(...values);
  const max = Math.max(...values);
  const span = max - min || 1;

  const points = data.map((point, i) => {
    const x = data.length === 1 ? 0 : (i / (data.length - 1)) * WIDTH;
    const y = height - ((point.value - min) / span) * height;
    return `${x.toFixed(2)},${y.toFixed(2)}`;
  });

  const strokeColor = color ?? 'var(--viz-1)';
  const areaPoints = `0,${height} ${points.join(' ')} ${WIDTH},${height}`;
  const gradientId = `line-chart-gradient-${data.length}-${min}-${max}`.replace(/[^a-zA-Z0-9-]/g, '');

  return (
    <div className="flex flex-col gap-1">
      <svg viewBox={`0 0 ${WIDTH} ${height}`} width="100%" height={height} preserveAspectRatio="none">
        <defs>
          <linearGradient id={gradientId} x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" style={{ stopColor: strokeColor, stopOpacity: 0.16 }} />
            <stop offset="100%" style={{ stopColor: strokeColor, stopOpacity: 0 }} />
          </linearGradient>
        </defs>
        <polygon points={areaPoints} style={{ fill: `url(#${gradientId})`, stroke: 'none' }} />
        <polyline points={points.join(' ')} style={{ fill: 'none', stroke: strokeColor, strokeWidth: 1.5 }} />
      </svg>
      <div className="flex items-center justify-between text-[11px]" style={{ color: 'var(--text-faint)' }}>
        <span>{min.toFixed(1)}</span>
        <span>{max.toFixed(1)}</span>
      </div>
    </div>
  );
}
