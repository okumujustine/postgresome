import { cva, type VariantProps } from 'class-variance-authority';
import type { HTMLAttributes } from 'react';
import { cn } from '../../lib/utils';

const badgeVariants = cva(
  'inline-flex items-center rounded-full border px-2.5 py-1 font-mono text-[11px] font-medium tracking-[0.01em]',
  {
    variants: {
      variant: {
        neutral: 'border-[var(--border)] bg-[var(--muted)] text-[var(--muted-foreground)]',
        critical: 'border-transparent bg-[var(--danger-soft)] text-[var(--danger)]',
        warning: 'border-transparent bg-[var(--warning-soft)] text-[var(--warning)]',
        success: 'border-transparent bg-[var(--success-soft)] text-[var(--success)]',
        info: 'border-transparent bg-[var(--info-soft)] text-[var(--info)]',
      },
    },
    defaultVariants: {
      variant: 'neutral',
    },
  },
);

export interface BadgeProps extends HTMLAttributes<HTMLDivElement>, VariantProps<typeof badgeVariants> {}

export function Badge({ className, variant, ...props }: BadgeProps) {
  return <div className={cn(badgeVariants({ variant }), className)} {...props} />;
}
