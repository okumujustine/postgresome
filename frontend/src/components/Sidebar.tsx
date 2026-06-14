import { AlertTriangle, Database, HeartPulse, Table2, Terminal } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import { NavLink } from 'react-router-dom';

interface NavItem {
  to: string;
  label: string;
  icon: LucideIcon;
  end: boolean;
}

const DATABASE_ITEMS: NavItem[] = [
  { to: '/', label: 'Health', icon: HeartPulse, end: true },
  { to: '/issues', label: 'Issues', icon: AlertTriangle, end: false },
  { to: '/queries', label: 'Queries', icon: Terminal, end: false },
  { to: '/tables', label: 'Tables', icon: Table2, end: false },
];

function NavList({ items }: { items: NavItem[] }) {
  return (
    <div className="flex flex-col gap-[3px]">
      {items.map(({ to, label, icon: Icon, end }) => (
        <NavLink
          key={to}
          to={to}
          end={end}
          className={({ isActive }) =>
            `relative flex h-[34px] items-center gap-[10px] rounded-[var(--radius-md)] px-[11px] text-[13px] font-medium no-underline transition-colors ${
              isActive ? '' : 'hover:bg-[var(--surface-hover)]'
            }`
          }
          style={({ isActive }) =>
            isActive
              ? {
                  color: 'var(--text-primary)',
                  background: 'var(--surface-hover)',
                  boxShadow: 'inset 0 0 0 1px var(--border-subtle)',
                }
              : { color: 'var(--text-secondary)' }
          }
        >
          {({ isActive }) => (
            <>
              {isActive && (
                <span
                  className="absolute left-0 rounded-[2px]"
                  style={{ top: 7, bottom: 7, width: 2, background: 'var(--accent)' }}
                />
              )}
              <Icon size={16} style={{ color: isActive ? 'var(--accent)' : 'var(--text-muted)' }} />
              <span>{label}</span>
            </>
          )}
        </NavLink>
      ))}
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
          style={{ background: 'var(--surface-hover)', color: 'var(--blue-600)', border: '1px solid var(--border-subtle)' }}
        >
          <Database size={15} strokeWidth={2} />
        </div>
        <div className="min-w-0">
          <div className="text-[15px] font-semibold" style={{ color: 'var(--text-primary)', letterSpacing: 'var(--ls-tight)' }}>
            Postgresome
          </div>
          <div className="truncate text-[11px]" style={{ color: 'var(--text-muted)' }}>
            PostgreSQL diagnosis
          </div>
        </div>
      </div>

      <nav className="flex flex-1 flex-col gap-[18px] overflow-y-auto p-3">
        <NavList items={DATABASE_ITEMS} />
      </nav>
    </aside>
  );
}
