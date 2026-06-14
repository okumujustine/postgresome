const SEVERITY_STYLES: Record<string, { fg: string; bg: string; border: string }> = {
  critical: { fg: 'var(--danger)', bg: 'var(--danger-tint)', border: 'rgba(207,34,46,0.25)' },
  warning: { fg: 'var(--warning)', bg: 'var(--warning-tint)', border: 'rgba(154,103,0,0.25)' },
  info: { fg: 'var(--blue-600)', bg: 'var(--blue-tint)', border: 'rgba(41,98,224,0.25)' },
};

export function SeverityPill({ severity }: { severity: string }) {
  const style = SEVERITY_STYLES[severity.toLowerCase()] ?? {
    fg: 'var(--text-secondary)',
    bg: 'var(--surface-active)',
    border: 'var(--border-default)',
  };

  return (
    <span
      className="inline-flex h-[20px] items-center rounded-[var(--radius-pill)] px-[8px] text-[11px] font-medium capitalize"
      style={{ background: style.bg, color: style.fg, border: `1px solid ${style.border}` }}
    >
      {severity}
    </span>
  );
}
