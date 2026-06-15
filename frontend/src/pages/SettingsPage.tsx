import { AppShell } from '../components/app-shell';
import { DetailCard } from '../components/ui/card';

export function SettingsPage() {
  return (
    <AppShell title="Settings" subtitle="Tune how Postgresome reports, not how many charts it renders.">
      <div className="grid gap-6 xl:grid-cols-[minmax(0,1.1fr)_minmax(0,0.9fr)]">
        <DetailCard title="Diagnosis preferences" description="Choose how findings should be surfaced to the team.">
          <div className="space-y-4">
            <SettingRow title="Priority notifications" description="Send alerts when a new critical diagnosis is opened." value="Enabled" />
            <SettingRow title="Daily summary" description="Compile recent changes and recommended next actions into one report." value="Enabled" />
            <SettingRow title="Confidence threshold" description="Hide low-confidence findings until supporting evidence improves." value="High only" />
          </div>
        </DetailCard>

        <DetailCard title="Workspace posture" description="The product should feel like a database doctor, not a metrics wall.">
          <div className="space-y-4 text-sm leading-7 text-[var(--muted-foreground)]">
            <p>Evidence stays available, but diagnosis remains the first thing operators see.</p>
            <p>Recommended actions should stay crisp, reversible, and grounded in proof from the collected PostgreSQL signals.</p>
            <p>Historical memory exists to explain what changed, when it changed, and whether a fix worked.</p>
          </div>
        </DetailCard>
      </div>
    </AppShell>
  );
}

function SettingRow({ title, description, value }: { title: string; description: string; value: string }) {
  return (
    <div className="rounded-2xl border border-[var(--border)] bg-[var(--muted)] px-4 py-4">
      <div className="flex items-center justify-between gap-4">
        <div>
          <div className="text-sm font-semibold text-[var(--foreground)]">{title}</div>
          <div className="mt-1 text-sm leading-6 text-[var(--muted-foreground)]">{description}</div>
        </div>
        <div className="rounded-full bg-[var(--panel)] px-3 py-1 text-xs font-semibold text-[var(--foreground)] shadow-[var(--shadow-xs)]">{value}</div>
      </div>
    </div>
  );
}
