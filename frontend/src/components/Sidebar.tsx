import { AlertTriangle, LayoutDashboard, LineChart, Table2, Terminal } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import { NavLink } from 'react-router-dom';

interface NavItem {
  to: string;
  label: string;
  icon: LucideIcon;
  end: boolean;
}

const MONITORING_ITEMS: NavItem[] = [
  { to: '/', label: 'Overview', icon: LayoutDashboard, end: true },
  { to: '/issues', label: 'Issues', icon: AlertTriangle, end: false },
  { to: '/queries', label: 'Queries', icon: Terminal, end: false },
  { to: '/tables', label: 'Tables', icon: Table2, end: false },
];

const ADVANCED_ITEMS: NavItem[] = [
  { to: '/metrics', label: 'Metrics', icon: LineChart, end: false },
];

function NavSection({ title, items }: { title: string; items: NavItem[] }) {
  return (
    <div>
      <div
        className="px-[11px] pb-[7px] text-[10.5px] font-semibold uppercase"
        style={{ color: 'var(--text-faint)', letterSpacing: 'var(--ls-label)' }}
      >
        {title}
      </div>
      <div className="flex flex-col gap-[3px]">
        {items.map(({ to, label, icon: Icon, end }) => (
          <NavLink
            key={to}
            to={to}
            end={end}
            className={({ isActive }) =>
              `relative flex h-[38px] items-center gap-[11px] rounded-[var(--radius-md)] px-[11px] text-[13.5px] font-semibold no-underline transition-colors ${
                isActive ? '' : 'hover:bg-[var(--surface-hover)]'
              }`
            }
            style={({ isActive }) =>
              isActive
                ? {
                    color: 'var(--text-primary)',
                    background: 'var(--blue-tint)',
                    boxShadow: 'inset 0 0 0 1px rgba(41,98,224,0.25)',
                  }
                : { color: 'var(--text-secondary)' }
            }
          >
            {({ isActive }) => (
              <>
                {isActive && (
                  <span
                    className="absolute left-0 rounded-[2px]"
                    style={{ top: 8, bottom: 8, width: 2.5, background: 'var(--accent)' }}
                  />
                )}
                <Icon size={16} style={{ color: isActive ? 'var(--accent)' : 'var(--text-muted)' }} />
                <span>{label}</span>
              </>
            )}
          </NavLink>
        ))}
      </div>
    </div>
  );
}

export function Sidebar() {
  return (
    <aside
      className="relative flex h-full w-[var(--sidebar-w)] shrink-0 flex-col border-r"
      style={{ background: 'var(--bg-base)', borderColor: 'var(--border-subtle)' }}
    >
      <div
        className="flex items-center gap-[9px] border-b px-4"
        style={{ height: 'var(--header-h)', borderColor: 'var(--border-subtle)' }}
      >
        <div
          className="flex h-[26px] w-[26px] shrink-0 items-center justify-center rounded-[var(--radius-sm)]"
          style={{ background: 'var(--blue-tint)', color: 'var(--blue-600)' }}
        >
          <LayoutDashboard size={15} strokeWidth={2} />
        </div>
        <span className="text-[15.5px] font-semibold" style={{ color: 'var(--text-primary)', letterSpacing: 'var(--ls-tight)' }}>
          Postgres<span style={{ color: 'var(--blue-500)' }}>ome</span>
        </span>
      </div>

      <nav className="flex flex-1 flex-col gap-[18px] overflow-y-auto p-3">
        <NavSection title="Monitoring" items={MONITORING_ITEMS} />
        <NavSection title="Advanced" items={ADVANCED_ITEMS} />
      </nav>
    </aside>
  );
}
