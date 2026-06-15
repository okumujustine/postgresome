import type { InputHTMLAttributes } from 'react';
import { cn } from '../../lib/utils';

export function Input({ className, ...props }: InputHTMLAttributes<HTMLInputElement>) {
  return (
    <input
      className={cn(
        'flex h-10 w-full rounded-[10px] border border-[var(--border)] bg-[var(--panel)] px-3 py-2 text-[14px] text-[var(--foreground)] outline-none placeholder:text-[var(--muted-helper)] focus-visible:ring-2 focus-visible:ring-[var(--ring)]',
        className,
      )}
      {...props}
    />
  );
}
