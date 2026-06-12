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
      style={{ background: 'var(--surface-card)', borderColor: 'var(--border-subtle)', boxShadow: 'var(--shadow-xs)' }}
    >
      {title && (
        <div
          className="flex items-center gap-[10px] border-b px-5 py-4"
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
              <div className="mt-[2px] text-xs" style={{ color: 'var(--text-muted)' }}>
                {subtitle}
              </div>
            )}
          </div>
          {actions && <div className="flex shrink-0 items-center gap-[6px]">{actions}</div>}
        </div>
      )}
      <div className="p-5">{children}</div>
    </div>
  );
}
