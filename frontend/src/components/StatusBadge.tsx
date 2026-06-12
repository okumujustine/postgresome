import type { InstanceStatus } from '../types/dashboard';

interface StatusBadgeProps {
  status: InstanceStatus | string;
  size?: 'sm' | 'md';
}

const STATUS_MAP: Record<string, { color: string; label: string; pulse: boolean }> = {
  healthy: { color: 'var(--success)', label: 'Healthy', pulse: true },
  warning: { color: 'var(--warning)', label: 'Warning', pulse: false },
  critical: { color: 'var(--danger)', label: 'Critical', pulse: false },
  unknown: { color: 'var(--text-muted)', label: 'Unknown', pulse: false },
};

export function StatusBadge({ status, size = 'md' }: StatusBadgeProps) {
  const entry = STATUS_MAP[status] ?? STATUS_MAP.unknown;
  const fontSize = size === 'sm' ? 12 : 13;
  const dot = size === 'sm' ? 7 : 8;
  const animate = entry.pulse;

  return (
    <span
      className="inline-flex items-center gap-[7px] font-medium"
      style={{ fontFamily: 'var(--font-sans)', fontSize, color: 'var(--text-secondary)' }}
    >
      <span className="relative inline-block shrink-0" style={{ width: dot, height: dot }}>
        {animate && (
          <span
            className="absolute inset-0 rounded-full"
            style={{ background: entry.color, opacity: 0.55, animation: 'psm-ping 1.8s var(--ease-out) infinite' }}
          />
        )}
        <span
          className="absolute inset-0 rounded-full"
          style={{ background: entry.color, boxShadow: entry.pulse ? 'var(--glow-success)' : 'none' }}
        />
      </span>
      <span>{entry.label}</span>
    </span>
  );
}
