import { ChevronDown, Database, Menu, RefreshCw } from 'lucide-react';
import type { InstanceStatus } from '../types/dashboard';
import type { MetricRange } from '../types/dashboard';
import { StatusBadge } from './StatusBadge';

const RANGE_OPTIONS: { value: MetricRange; label: string }[] = [
  { value: '15m', label: 'Last 15m' },
  { value: '1h', label: 'Last 1h' },
  { value: '6h', label: 'Last 6h' },
  { value: '24h', label: 'Last 24h' },
  { value: '7d', label: 'Last 7d' },
];

export interface HeaderProps {
  title: string;
  databaseName?: string;
  status?: InstanceStatus | string;
  range: MetricRange;
  onRangeChange: (range: MetricRange) => void;
  onRefresh: () => void;
  refreshing: boolean;
  onHamburger: () => void;
}

export function Header({ title, databaseName, status, range, onRangeChange, onRefresh, refreshing, onHamburger }: HeaderProps) {
  return (
    <header
      className="psm-header sticky top-0 z-20 flex shrink-0 items-center gap-[14px] border-b px-5 backdrop-blur-sm"
      style={{ height: 'var(--header-h)', borderColor: 'var(--border-subtle)', background: 'color-mix(in srgb, var(--bg-base) 82%, transparent)' }}
    >
      <button
        onClick={onHamburger}
        aria-label="Menu"
        className="inline-flex cursor-pointer border-none bg-transparent p-1 md:hidden"
        style={{ color: 'var(--text-secondary)' }}
      >
        <Menu size={20} />
      </button>

      <div className="flex min-w-0 items-center gap-[11px]">
        <h1 className="m-0 truncate text-base font-semibold" style={{ color: 'var(--text-primary)', letterSpacing: 'var(--ls-snug)' }}>
          {title}
        </h1>
        {databaseName && (
          <>
            <span className="hidden h-[18px] w-px sm:block" style={{ background: 'var(--border-default)' }} />
            <div
              className="hidden items-center gap-2 rounded-[var(--radius-pill)] border px-[9px] py-1 sm:flex"
              style={{ background: 'var(--surface-card)', borderColor: 'var(--border-subtle)' }}
            >
              <Database size={13} style={{ color: 'var(--text-muted)' }} />
              <span className="text-[12.5px]" style={{ fontFamily: 'var(--font-mono)', color: 'var(--text-secondary)' }}>
                {databaseName}
              </span>
              {status && <StatusBadge status={status} size="sm" />}
            </div>
          </>
        )}
      </div>

      <div className="flex-1" />

      <div className="relative inline-block">
        <select
          value={range}
          onChange={(event) => onRangeChange(event.target.value as MetricRange)}
          className="h-[var(--control-h-sm)] cursor-pointer appearance-none rounded-[var(--radius-md)] border pr-8 pl-3 text-[13px] outline-none"
          style={{ background: 'var(--surface-raised)', color: 'var(--text-primary)', borderColor: 'var(--border-default)', fontFamily: 'var(--font-sans)' }}
        >
          {RANGE_OPTIONS.map((option) => (
            <option key={option.value} value={option.value} style={{ background: 'var(--surface-card)', color: 'var(--text-primary)' }}>
              {option.label}
            </option>
          ))}
        </select>
        <span className="pointer-events-none absolute right-[11px] top-1/2 -translate-y-1/2" style={{ color: 'var(--text-muted)' }}>
          <ChevronDown size={14} />
        </span>
      </div>

      <button
        onClick={onRefresh}
        aria-label="Refresh"
        disabled={refreshing}
        className="inline-flex h-[var(--control-h-sm)] w-[var(--control-h-sm)] shrink-0 cursor-pointer items-center justify-center rounded-[var(--radius-sm)] border"
        style={{ background: 'var(--surface-raised)', color: 'var(--text-secondary)', borderColor: 'var(--border-default)' }}
      >
        <RefreshCw size={15} className={refreshing ? 'animate-spin' : ''} />
      </button>
    </header>
  );
}
