import {
  ChevronLeft,
  ChevronRight,
  ChevronsUpDown,
  Clock3,
  FileSearch,
  Menu,
  Settings,
  Stethoscope,
} from "lucide-react";
import { NavLink, Outlet } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Sheet } from "@/components/ui/sheet";
import { cn } from "@/lib/utils";
import { useWorkspace } from "@/lib/workspace-context";
import { useState } from "react";

const navigation = [
  { to: "/diagnosis", label: "Diagnosis", icon: Stethoscope, primary: true },
  { to: "/history", label: "History", icon: Clock3 },
  { to: "/queries", label: "Query Explorer", icon: FileSearch },
  { to: "/setup", label: "Settings", icon: Settings },
];

function SidebarContent({
  collapsed = false,
  onNavigate,
  onToggleCollapse,
}: {
  collapsed?: boolean;
  onNavigate?: () => void;
  onToggleCollapse?: () => void;
}) {
  const { selectedSource, sources, selectSource } = useWorkspace();

  return (
    <div className="flex h-full min-h-0 flex-col">
      <div className={cn("border-b border-border pb-5 pt-5", collapsed ? "px-3" : "px-5")}>
        <div className={cn("flex items-center", collapsed ? "justify-center" : "")}>
          {!collapsed ? (
            <div>
              <div className="font-heading text-[24px] font-semibold tracking-[-0.05em] text-white">
                Postgresome
              </div>
            </div>
          ) : (
            <div className="text-[16px] font-semibold text-white">
              PS
            </div>
          )}
        </div>
      </div>

      <nav
        className={cn(
          "flex-1 overflow-y-auto py-4",
          collapsed ? "space-y-2 px-2" : "space-y-1 px-3",
        )}
      >
        {navigation.map((item) => {
          const Icon = item.icon;

          return (
            <NavLink
              key={item.to}
              to={item.to}
              onClick={onNavigate}
              className={({ isActive }) =>
                cn(
                  "flex rounded-md border text-sm transition-all duration-150",
                  collapsed
                    ? "justify-center px-2 py-2.5"
                    : "items-center justify-between px-3 py-2.5",
                  isActive
                    ? "border-white bg-[#fff3a6] text-black shadow-[2px_2px_0_#ffffff]"
                    : "border-transparent text-slate-300 hover:border-white hover:bg-white hover:text-black hover:shadow-[2px_2px_0_#ffffff]",
                )
              }
            >
              {({ isActive }) => (
                <>
                  <div className={cn("flex items-center", collapsed ? "justify-center" : "gap-3")}>
                    <Icon
                      className={cn(
                        "h-4 w-4",
                        isActive ? "text-black" : "text-slate-400",
                      )}
                    />
                    {!collapsed ? (
                      <span
                        className={cn(
                          "text-[15px] font-medium",
                          item.primary && isActive ? "text-black" : "",
                        )}
                      >
                        {item.label}
                      </span>
                    ) : null}
                  </div>
                  {!collapsed && isActive ? (
                    <ChevronRight className="h-4 w-4 text-black" />
                  ) : null}
                </>
              )}
            </NavLink>
          );
        })}
      </nav>

      {onToggleCollapse ? (
        <div className={cn("border-t border-white/10 pb-4 pt-3", collapsed ? "px-2" : "px-3")}>
          {!collapsed ? (
            <div className="mb-3 rounded-md border border-white/25 bg-white/[0.04] p-2.5">
              <div className="mb-2 text-[10px] font-semibold uppercase tracking-[0.18em] text-slate-400">
                Current source
              </div>
              {sources.length > 0 ? (
                <label className="relative block">
                  <select
                    value={selectedSource?.source.id ?? ""}
                    onChange={(event) => selectSource(event.target.value)}
                    className="h-10 w-full appearance-none rounded-md border border-white/25 bg-white px-3 pr-9 text-[13px] font-medium text-black outline-none transition-colors hover:bg-[#fff7cf] focus:border-white"
                  >
                    {sources.map((source) => (
                      <option key={source.source.id} value={source.source.id}>
                        {source.database.name}
                      </option>
                    ))}
                  </select>
                  <ChevronsUpDown className="pointer-events-none absolute right-3 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-black/70" />
                </label>
              ) : (
                <div className="rounded-md border border-dashed border-white/20 px-3 py-2 text-[12px] text-slate-400">
                  No source selected
                </div>
              )}
            </div>
          ) : null}
          <Button
            variant="ghost"
            className={cn(
              "w-full justify-start text-slate-300 hover:text-black",
              collapsed ? "px-2" : "px-3",
            )}
            onClick={onToggleCollapse}
            aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
          >
            {collapsed ? (
              <ChevronRight className="h-4 w-4" />
            ) : (
              <>
                <ChevronLeft className="h-4 w-4" />
                <span>Collapse</span>
              </>
            )}
          </Button>
        </div>
      ) : null}
    </div>
  );
}

export function AppShell() {
  const [mobileNavOpen, setMobileNavOpen] = useState(false);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const { selectedSource, sources, selectSource } = useWorkspace();

  return (
    <div className="app-frame">
      <div className="flex h-screen overflow-hidden">
        <aside
          className={cn(
            "hidden border-r-2 border-r-[#111111] bg-[#111111] transition-[width] duration-150 lg:block",
            sidebarCollapsed ? "w-[76px]" : "w-[240px]",
          )}
        >
          <div className="h-full overflow-hidden">
            <SidebarContent
              collapsed={sidebarCollapsed}
              onToggleCollapse={() => setSidebarCollapsed((current) => !current)}
            />
          </div>
        </aside>

        <div className="flex min-w-0 flex-1 flex-col overflow-hidden">
          <main className="min-h-0 flex-1 overflow-hidden">
            <div className="h-full overflow-y-auto p-6">
              <div className="mb-4 flex items-center justify-between lg:hidden">
                <Sheet
                  open={mobileNavOpen}
                  onOpenChange={setMobileNavOpen}
                  trigger={
                    <Button variant="outline" size="icon" aria-label="Open navigation">
                      <Menu className="h-4 w-4" />
                    </Button>
                  }
                  title="Postgresome"
                >
                  <SidebarContent onNavigate={() => setMobileNavOpen(false)} />
                </Sheet>
                <div className="flex items-center gap-2">
                  {sources.length > 0 ? (
                    <label className="relative block">
                      <select
                        value={selectedSource?.source.id ?? ""}
                        onChange={(event) => selectSource(event.target.value)}
                        className="h-10 min-w-[180px] appearance-none rounded-md border border-[#111111] bg-[#fff3a6] px-3 pr-9 text-[13px] font-medium text-black shadow-[1px_1px_0_#111111] outline-none"
                      >
                        {sources.map((source) => (
                          <option key={source.source.id} value={source.source.id}>
                            {source.database.name}
                          </option>
                        ))}
                      </select>
                      <ChevronsUpDown className="pointer-events-none absolute right-3 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-black/70" />
                    </label>
                  ) : null}
                  <Button variant="ghost" size="icon" aria-label="Settings">
                    <Settings className="h-4 w-4" />
                  </Button>
                </div>
              </div>
              <Outlet />
            </div>
          </main>
        </div>
      </div>
    </div>
  );
}
