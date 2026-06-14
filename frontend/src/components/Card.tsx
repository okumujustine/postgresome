import type { ReactNode } from 'react';

interface CardProps {
  title?: string;
  subtitle?: string;
  actions?: ReactNode;
  children: ReactNode;
}

export function Card({ title, subtitle, actions, children }: CardProps) {
  return (
    <div
      className="min-w-0 overflow-hidden rounded-[var(--radius-lg)] border"
      style={{ background: 'var(--surface-card)', borderColor: 'var(--border-subtle)' }}
    >
      {title && (
        <div
          className="flex items-center gap-[10px] border-b px-4 py-3"
          style={{ borderColor: 'var(--border-subtle)' }}
        >
          <div className="min-w-0 flex-1">
            <div
              className="font-semibold"
              style={{ fontSize: 'var(--fs-title)', color: 'var(--text-primary)', letterSpacing: 'var(--ls-snug)' }}
            >
              {title}
            </div>
            {subtitle && (
              <div className="mt-[3px] text-[12.5px]" style={{ color: 'var(--text-muted)' }}>
                {subtitle}
              </div>
            )}
          </div>
          {actions && <div className="flex shrink-0 items-center gap-[6px]">{actions}</div>}
        </div>
      )}
      <div className="p-4">{children}</div>
    </div>
  );
}
