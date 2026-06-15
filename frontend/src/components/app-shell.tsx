import {
  Bell,
  ChevronDown,
  Database,
  HeartPulse,
  MessageCircleMore,
  RefreshCw,
  Settings,
  Search,
  FileBarChart2,
} from 'lucide-react';
import { Link, NavLink, useNavigate } from 'react-router-dom';
import { useMemo, useState, type ReactNode } from 'react';
import { useDatabaseInstance } from '../lib/databaseInstance';
import { Button } from './ui/button';
import { CommandMenu } from './ui/command';

const NAV_ITEMS = [
  { to: '/', label: 'Overview', icon: HeartPulse },
  { to: '/findings', label: 'Findings', icon: FileBarChart2 },
  { to: '/settings', label: 'Settings', icon: Settings },
];

export function AppShell({ title, subtitle, children }: { title: string; subtitle: string; children: ReactNode }) {
  const navigate = useNavigate();
  const [commandOpen, setCommandOpen] = useState(false);
  const [range, setRange] = useState('6h');
  const { instances, selectedId, setSelectedId, loading } = useDatabaseInstance();
  const selectedInstance = instances.find((instance) => instance.id === selectedId);

  const commandItems = useMemo(
    () =>
      NAV_ITEMS.map((item) => ({
        id: item.to,
        title: item.label,
        description: `Open the ${item.label.toLowerCase()} workspace`,
        onSelect: () => navigate(item.to),
      })),
    [navigate],
  );

  return (
    <div className="min-h-screen bg-[var(--background)] text-[var(--foreground)]">
      <div className="grid min-h-screen lg:grid-cols-[252px_minmax(0,1fr)]">
        <aside className="hidden border-r border-[var(--border)] bg-[var(--panel)] lg:flex lg:flex-col">
          <div className="border-b border-[var(--border)] px-6 py-6">
            <Link to="/" className="inline-flex items-center gap-3 no-underline">
              <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-[var(--muted)] text-[var(--foreground)]">
                <Database size={17} />
              </div>
              <div>
                <div className="font-mono text-[20px] font-semibold tracking-[-0.03em] text-[var(--foreground)]">Postgresome</div>
                <div className="font-mono text-[12px] text-[var(--muted-foreground)]">Database doctor</div>
              </div>
            </Link>
          </div>

          <div className="px-4 py-5">
            <button
              type="button"
              onClick={() => setCommandOpen(true)}
              className="w-full rounded-xl border border-[var(--border)] bg-[var(--muted)] px-4 py-4 text-left transition-colors hover:bg-[var(--muted-strong)]"
            >
              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0">
                  <div className="truncate font-mono text-[13px] font-semibold text-[var(--foreground)]">
                    {selectedInstance?.database_name ?? 'postgresome_secondary'}
                  </div>
                  <div className="mt-2 flex items-center gap-2 font-mono text-[11px] text-[var(--muted-foreground)]">
                    <span className="h-2.5 w-2.5 rounded-full bg-[var(--success)]" />
                    {loading ? 'Loading database instance...' : 'PostgreSQL 16.2'}
                  </div>
                </div>
                <ChevronDown size={16} className="mt-1 shrink-0 text-[var(--muted-foreground)]" />
              </div>
            </button>
          </div>

          <nav className="space-y-1 px-4">
            {NAV_ITEMS.map((item) => (
              <NavLink
                key={item.to}
                to={item.to}
                className={({ isActive }) =>
                  `flex items-center gap-3 rounded-[10px] px-4 py-2.5 font-mono text-[13px] font-medium no-underline transition-colors ${
                    isActive ? 'bg-[var(--muted)] text-[var(--foreground)]' : 'text-[var(--muted-foreground)] hover:bg-[var(--muted)] hover:text-[var(--foreground)]'
                  }`
                }
              >
                <item.icon size={16} />
                {item.label}
              </NavLink>
            ))}
          </nav>

          <div className="mt-auto px-4 pb-5">
            <button
              type="button"
              className="flex w-full items-center gap-3 rounded-[10px] px-4 py-2.5 text-left text-[13px] text-[var(--muted-foreground)] transition-colors hover:bg-[var(--muted)]"
            >
              <MessageCircleMore size={16} />
              Give feedback
            </button>
          </div>
        </aside>

        <div className="min-w-0">
          <header className="sticky top-0 z-30 border-b border-[var(--border)] bg-[rgba(250,250,250,0.96)] backdrop-blur">
            <div className="flex items-center gap-4 px-8 py-6">
              <div className="min-w-0 flex-1">
                <div className="font-mono text-[11px] uppercase tracking-[0.08em] text-[var(--muted-foreground)]">Postgresome Console</div>
                <h1 className="mt-2 truncate font-mono text-[22px] font-semibold tracking-[-0.03em] text-[var(--foreground)]">{title}</h1>
                <p className="mt-1 max-w-2xl text-[14px] leading-6 text-[var(--muted-foreground)]">{subtitle}</p>
              </div>

              <div className="flex items-center gap-3">
                <div className="relative hidden md:block">
                  <select
                    value={range}
                    onChange={(event) => setRange(event.target.value)}
                    className="h-10 min-w-[156px] appearance-none rounded-[10px] border border-[var(--border)] bg-[var(--panel)] px-3 pr-9 font-mono text-[12px] text-[var(--foreground)] outline-none"
                  >
                    <option value="1h">Last 1 hour</option>
                    <option value="6h">Last 6 hours</option>
                    <option value="24h">Last 24 hours</option>
                    <option value="7d">Last 7 days</option>
                  </select>
                  <ChevronDown size={14} className="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-[var(--muted-foreground)]" />
                </div>

                <Button variant="outline" size="icon" onClick={() => setCommandOpen(true)} className="rounded-full">
                  <Search size={15} />
                </Button>
                <Button variant="outline" size="icon" onClick={() => setSelectedId(selectedId)} className="rounded-full">
                  <RefreshCw size={15} />
                </Button>
                <Button variant="outline" size="icon" className="rounded-full">
                  <Bell size={15} />
                </Button>
                <div className="flex h-10 w-10 items-center justify-center rounded-full bg-[var(--muted)] text-[12px] font-semibold text-[var(--foreground)]">
                  JD
                </div>
              </div>
            </div>
          </header>

          <main className="px-8 py-8">{children}</main>
        </div>
      </div>

      <CommandMenu open={commandOpen} onOpenChange={setCommandOpen} title="Find anything in Postgresome" items={commandItems} />
    </div>
  );
}
