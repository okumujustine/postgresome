import type { InstanceStatus } from '../types/dashboard';

interface StatusBadgeProps {
  status: InstanceStatus | string;
  size?: 'sm' | 'md';
}

const STATUS_MAP: Record<string, { color: string; label: string }> = {
  healthy: { color: 'var(--success)', label: 'Healthy' },
  warning: { color: 'var(--warning)', label: 'Warning' },
  critical: { color: 'var(--danger)', label: 'Critical' },
  unknown: { color: 'var(--text-muted)', label: 'Unknown' },
};

export function StatusBadge({ status, size = 'md' }: StatusBadgeProps) {
  const entry = STATUS_MAP[status] ?? STATUS_MAP.unknown;
  const fontSize = size === 'sm' ? 12 : 13;
  const dot = size === 'sm' ? 7 : 8;

  return (
    <span
      className="inline-flex items-center gap-[7px] font-medium"
      style={{ fontFamily: 'var(--font-sans)', fontSize, color: 'var(--text-secondary)' }}
    >
      <span className="relative inline-block shrink-0" style={{ width: dot, height: dot }}>
        <span className="absolute inset-0 rounded-full" style={{ background: entry.color }} />
      </span>
      <span>{entry.label}</span>
    </span>
  );
}
