import { useState, type ReactNode } from 'react';
import { ChevronDown } from 'lucide-react';
import { cn } from '../../lib/utils';

export function Accordion({ children, className }: { children: ReactNode; className?: string }) {
  return <div className={cn('divide-y divide-[var(--border)] rounded-lg border border-[var(--border)]', className)}>{children}</div>;
}

export function AccordionItem({
  title,
  subtitle,
  children,
  defaultOpen = false,
}: {
  title: string;
  subtitle?: string;
  children: ReactNode;
  defaultOpen?: boolean;
}) {
  const [open, setOpen] = useState(defaultOpen);

  return (
    <div className="bg-[var(--panel)]">
      <button
        onClick={() => setOpen((current) => !current)}
        className="flex w-full items-center justify-between gap-4 px-4 py-3 text-left"
      >
        <div>
          <div className="text-sm font-medium text-[var(--foreground)]">{title}</div>
          {subtitle ? <div className="mt-1 text-sm text-[var(--muted-foreground)]">{subtitle}</div> : null}
        </div>
        <ChevronDown size={16} className={cn('text-[var(--muted-foreground)] transition-transform', open && 'rotate-180')} />
      </button>
      {open ? <div className="px-4 pb-4 text-sm leading-6 text-[var(--muted-foreground)]">{children}</div> : null}
    </div>
  );
}

