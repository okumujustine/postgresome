import { LayoutDashboard } from 'lucide-react';

interface SidebarProps {
  mobileOpen: boolean;
  onClose: () => void;
}

export function Sidebar({ mobileOpen, onClose }: SidebarProps) {
  return (
    <>
      <div
        onClick={onClose}
        className="fixed inset-0 z-40 bg-black/55 transition-opacity duration-200 md:hidden"
        style={{ opacity: mobileOpen ? 1 : 0, pointerEvents: mobileOpen ? 'auto' : 'none' }}
      />
      <aside
        className="fixed inset-y-0 left-0 z-50 flex h-full w-[var(--sidebar-w)] shrink-0 -translate-x-full flex-col border-r transition-transform duration-200 ease-out data-[open=true]:translate-x-0 data-[open=true]:shadow-[var(--shadow-xl)] md:relative md:translate-x-0 md:shadow-none"
        data-open={mobileOpen ? 'true' : 'false'}
        style={{ background: 'var(--bg-base)', borderColor: 'var(--border-subtle)' }}
      >
        <div
          className="flex items-center gap-[9px] border-b px-4"
          style={{ height: 'var(--header-h)', borderColor: 'var(--border-subtle)' }}
        >
          <div
            className="flex h-[26px] w-[26px] shrink-0 items-center justify-center rounded-[var(--radius-sm)]"
            style={{ background: 'var(--blue-tint)', color: 'var(--blue-400)' }}
          >
            <LayoutDashboard size={15} strokeWidth={2} />
          </div>
          <span className="text-[15.5px] font-semibold" style={{ color: 'var(--text-primary)', letterSpacing: 'var(--ls-tight)' }}>
            Postgres<span style={{ color: 'var(--blue-500)' }}>ome</span>
          </span>
        </div>

        <nav className="flex flex-1 flex-col gap-[18px] overflow-y-auto p-3">
          <div>
            <div
              className="px-[11px] pb-[7px] text-[10.5px] font-semibold uppercase"
              style={{ color: 'var(--text-faint)', letterSpacing: 'var(--ls-label)' }}
            >
              Monitoring
            </div>
            <div className="flex flex-col gap-[3px]">
              <span
                className="relative flex h-[38px] items-center gap-[11px] rounded-[var(--radius-md)] px-[11px] text-[13.5px] font-semibold"
                style={{
                  color: 'var(--text-primary)',
                  background: 'var(--blue-tint)',
                  boxShadow: 'inset 0 0 0 1px rgba(77,141,255,0.25)',
                }}
              >
                <span
                  className="absolute left-0 rounded-[2px]"
                  style={{ top: 8, bottom: 8, width: 2.5, background: 'var(--accent)' }}
                />
                <LayoutDashboard size={16} style={{ color: 'var(--accent)' }} />
                <span>Overview</span>
              </span>
            </div>
          </div>
        </nav>
      </aside>
    </>
  );
}
