import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import { formatRelativeTime, severityClasses, severityLabel } from "@/lib/format";
import type { Finding, FindingListResponse } from "@/types/api";

interface FindingListProps {
  findings: Finding[];
  selectedFindingId: string | null;
  onSelect: (findingId: string) => void;
  search: string;
  onSearchChange: (value: string) => void;
  counts?: FindingListResponse["severity_counts"];
  total?: number;
}

export function FindingList({
  findings,
  selectedFindingId,
  onSelect,
  search,
  onSearchChange,
  counts,
  total,
}: FindingListProps) {
  return (
    <div className="flex h-full flex-col">
      <div className="space-y-4 border-b border-border/80 px-4 py-4">
        <div className="flex items-end justify-between gap-3">
          <div>
            <div className="kicker">Active queue</div>
            <div className="mt-2 font-heading text-[18px] font-semibold tracking-[-0.02em] text-foreground">
              Findings needing attention
            </div>
          </div>
          <div className="meta">{total ?? findings.length} active</div>
        </div>

        <Input
          value={search}
          onChange={(event) => onSearchChange(event.target.value)}
          placeholder="Filter by title, object, or rule"
        />

        {counts ? (
          <div className="flex flex-wrap gap-2">
            {[
              { label: "Critical", value: counts.critical, tone: "critical" },
              { label: "Warning", value: counts.warning, tone: "warning" },
              { label: "Info", value: counts.info, tone: "info" },
            ].map((item) => {
              const tone = severityClasses(item.tone);

              return (
                <div
                  key={item.label}
                  className={cn("rounded-md border-2 px-3 py-2 shadow-[2px_2px_0_#111111]", tone.badge)}
                >
                  <div className="text-[12px] font-medium">
                    {item.label} {item.value}
                  </div>
                </div>
              );
            })}
          </div>
        ) : null}
      </div>

      <div className="flex-1 overflow-y-auto p-3">
        <div className="space-y-2">
          {findings.map((finding) => {
            const tone = severityClasses(finding.severity);
            const selected = selectedFindingId === finding.id;

            return (
              <button
                key={finding.id}
                type="button"
                onClick={() => onSelect(finding.id)}
                className={cn(
                  "w-full rounded-md border-2 bg-white p-4 text-left shadow-[2px_2px_0_#111111] transition-all duration-150 hover:-translate-x-[1px] hover:-translate-y-[1px] hover:bg-[#f4f1e6] hover:shadow-[3px_3px_0_#111111]",
                  selected ? "surface-selection" : "border-border",
                )}
              >
                <div className="flex items-start justify-between gap-3">
                  <div className="space-y-2">
                    <div className={cn("kicker", tone.text)}>
                      {severityLabel(finding.severity)}
                    </div>
                    <div className="font-heading text-[17px] font-semibold leading-6 tracking-[-0.02em] text-foreground">
                      {finding.title}
                    </div>
                  </div>
                  <div className="meta whitespace-nowrap">
                    {formatRelativeTime(finding.detected_at)}
                  </div>
                </div>

                <div className="mt-3 flex flex-wrap items-center gap-2">
                  <span className="rounded-md border-2 border-[#111111] bg-[#eef2ff] px-2.5 py-1 text-[11px] font-semibold tracking-[0.01em] text-[#254fd2]">
                    {finding.resource_type || "object"}
                  </span>
                  <span className="text-sm text-slate-600">{finding.resource_name}</span>
                </div>
              </button>
            );
          })}

          {findings.length === 0 ? (
            <div className="rounded-md border-2 border-dashed border-[#111111] bg-white p-6 text-sm text-muted-foreground">
              No findings match the current filter.
            </div>
          ) : null}
        </div>
      </div>
    </div>
  );
}
