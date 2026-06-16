import {
  Bell,
  ChevronRight,
  Clock3,
  Database,
  FileSearch,
  Menu,
  Search,
  Settings,
  Stethoscope,
  Wrench,
} from "lucide-react";
import { NavLink, Outlet, useLocation } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Sheet } from "@/components/ui/sheet";
import { cn } from "@/lib/utils";
import { useWorkspace } from "@/lib/workspace-context";
import { statusClasses } from "@/lib/format";
import { useState } from "react";

const navigation = [
  { to: "/diagnosis", label: "Diagnosis", icon: Stethoscope, primary: true },
  { to: "/history", label: "History", icon: Clock3 },
  { to: "/queries", label: "Query Explorer", icon: FileSearch },
  { to: "/setup", label: "Setup", icon: Wrench },
];

const titles: Record<string, { title: string; subtitle: string }> = {
  "/diagnosis": {
    title: "Diagnosis",
    subtitle: "Work from active findings into evidence and the next action.",
  },
  "/history": {
    title: "History",
    subtitle: "Reconstruct how findings appeared, regressed, or improved over time.",
  },
  "/queries": {
    title: "Query Explorer",
    subtitle: "Drill into query behavior when a finding points toward SQL.",
  },
  "/setup": {
    title: "Setup",
    subtitle: "Connect a source, validate access, and start a fresh diagnosis run.",
  },
};

function SidebarContent({ onNavigate }: { onNavigate?: () => void }) {
  const { selectedSource } = useWorkspace();

  return (
    <div className="flex h-full flex-col">
      <div className="border-b border-border px-5 pb-5 pt-5">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-lg border bg-white">
            <Database className="h-5 w-5 text-slate-700" />
          </div>
          <div>
            <div className="font-heading text-[24px] font-semibold tracking-[-0.02em] text-foreground">
              Postgresome
            </div>
            <div className="meta">Diagnosis-first PostgreSQL assistant</div>
          </div>
        </div>

        <div className="technical-sheet mt-5 space-y-3 p-4">
          <div className="kicker">Current source</div>
          {selectedSource ? (
            <>
              <div className="text-sm font-semibold text-foreground">
                {selectedSource.database.name}
              </div>
              <div className="meta">{selectedSource.database.host}</div>
              <div
                className={cn(
                  "inline-flex w-fit rounded-md border px-2 py-1 font-mono text-[11px] uppercase tracking-[0.18em]",
                  statusClasses(selectedSource.instance.status),
                )}
              >
                {selectedSource.instance.status}
              </div>
            </>
          ) : (
            <div className="text-sm text-muted-foreground">
              No database connected yet.
            </div>
          )}
        </div>
      </div>

      <nav className="flex-1 space-y-1 px-3 py-4">
        {navigation.map((item) => {
          const Icon = item.icon;

          return (
            <NavLink
              key={item.to}
              to={item.to}
              onClick={onNavigate}
              className={({ isActive }) =>
                cn(
                  "flex items-center justify-between rounded-md px-3 py-2.5 text-sm transition-colors",
                  isActive
                    ? "border border-border bg-white text-foreground"
                    : "text-slate-600 hover:bg-white hover:text-foreground",
                )
              }
            >
              {({ isActive }) => (
                <>
                  <div className="flex items-center gap-3">
                    <Icon
                      className={cn(
                        "h-4 w-4",
                        isActive ? "text-slate-900" : "text-slate-500",
                      )}
                    />
                    <span className={item.primary ? "font-medium" : ""}>
                      {item.label}
                    </span>
                  </div>
                  {isActive ? <ChevronRight className="h-4 w-4 text-slate-400" /> : null}
                </>
              )}
            </NavLink>
          );
        })}
      </nav>

      <div className="border-t border-border px-3 py-4">
        <div className="rounded-md px-3 py-2">
          <div className="kicker">Flow</div>
          <div className="mt-2 text-sm text-slate-600">
            Setup <span className="px-1 text-slate-400">/</span> Diagnosis
            <span className="px-1 text-slate-400">/</span> Evidence
            <span className="px-1 text-slate-400">/</span> Action
          </div>
        </div>
      </div>
    </div>
  );
}

export function AppShell() {
  const [mobileNavOpen, setMobileNavOpen] = useState(false);
  const location = useLocation();
  const pageMeta = titles[location.pathname] ?? titles["/diagnosis"];

  return (
    <div className="app-frame">
      <div className="flex min-h-screen">
        <aside className="hidden w-[240px] border-r border-border/80 bg-[#f8f9ff] lg:block">
          <SidebarContent />
        </aside>

        <div className="flex min-w-0 flex-1 flex-col">
          <header className="sticky top-0 z-20 border-b border-border/80 bg-background/95 backdrop-blur">
            <div className="flex h-16 items-center justify-between gap-4 px-6">
              <div className="flex min-w-0 items-center gap-3">
                <div className="lg:hidden">
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
                </div>

                <div className="min-w-0">
                  <h1 className="truncate font-heading text-[24px] font-semibold tracking-[-0.01em] text-foreground">
                    {pageMeta.title}
                  </h1>
                  <p className="hidden text-sm text-muted-foreground sm:block">
                    {pageMeta.subtitle}
                  </p>
                </div>
              </div>

              <div className="flex items-center gap-2">
                <div className="hidden items-center gap-2 rounded-md border bg-white px-3 py-2 text-sm text-muted-foreground md:flex">
                  <Search className="h-4 w-4" />
                  <span className="meta">Search diagnosis or query id</span>
                </div>

                <Button variant="ghost" size="icon" aria-label="Notifications">
                  <Bell className="h-4 w-4" />
                </Button>
                <Button variant="ghost" size="icon" aria-label="Settings">
                  <Settings className="h-4 w-4" />
                </Button>
              </div>
            </div>
          </header>

          <main className="flex-1 p-6">
            <Outlet />
          </main>
        </div>
      </div>
    </div>
  );
}
