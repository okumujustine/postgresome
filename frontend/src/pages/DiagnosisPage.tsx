import { useEffect, useState } from "react";
import { ArrowRight, ListFilter } from "lucide-react";
import { Link, useSearchParams } from "react-router-dom";
import { FindingList } from "@/components/finding-list";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { DismissibleAlert } from "@/components/ui/dismissible-alert";
import { Input } from "@/components/ui/input";
import { Sheet } from "@/components/ui/sheet";
import { listFindings } from "@/lib/api";
import { formatRelativeTime, severityLabel } from "@/lib/format";
import { useWorkspace } from "@/lib/workspace-context";
import type { FindingListResponse } from "@/types/api";

export function DiagnosisPage() {
  const { selectedInstanceId } = useWorkspace();
  const [listData, setListData] = useState<FindingListResponse | null>(null);
  const [loadingList, setLoadingList] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState("");
  const [mobileQueueOpen, setMobileQueueOpen] = useState(false);
  const [searchParams, setSearchParams] = useSearchParams();

  const focusedFindingId = searchParams.get("focus");

  useEffect(() => {
    if (!selectedInstanceId) {
      setListData(null);
      setLoadingList(false);
      return;
    }

    setLoadingList(true);
    setError(null);

    void listFindings({
      databaseInstanceId: selectedInstanceId,
      status: "open",
      limit: 40,
    })
      .then((response) => {
        setListData(response);
      })
      .catch((caught) => {
        setError(caught instanceof Error ? caught.message : "Failed to load findings");
        setListData(null);
      })
      .finally(() => {
        setLoadingList(false);
      });
  }, [selectedInstanceId]);

  const filteredFindings =
    listData?.findings.filter((finding) => {
      const needle = search.trim().toLowerCase();
      if (!needle) {
        return true;
      }

      return [finding.title, finding.resource_name, finding.rule_key]
        .join(" ")
        .toLowerCase()
        .includes(needle);
    }) ?? [];

  return (
    <div className="mx-auto max-w-[1600px] space-y-4">
      {error ? <DismissibleAlert>{error}</DismissibleAlert> : null}

      <div className="space-y-1">
        <h1 className="font-heading text-[24px] font-semibold tracking-[-0.04em] text-foreground">
          Diagnosis listing
        </h1>
      </div>

      <div className="flex flex-wrap items-end justify-between gap-4">
        <div className="w-full max-w-xl space-y-2">
          <div className="kicker">Filter list</div>
          <Input
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder="Filter by title, object, or rule"
            className="bg-transparent"
          />
        </div>
        <div className="lg:hidden">
          <Sheet
            open={mobileQueueOpen}
            onOpenChange={setMobileQueueOpen}
            title="Diagnosis queue"
            trigger={
              <Button variant="outline">
                <ListFilter className="h-4 w-4" />
                Queue
              </Button>
            }
          >
            <div className="h-[calc(100vh-100px)]">
              <FindingList
                findings={filteredFindings}
                selectedFindingId={focusedFindingId}
                onSelect={(findingId) => {
                  setSearchParams({ focus: findingId });
                  setMobileQueueOpen(false);
                }}
                search={search}
                onSearchChange={setSearch}
                counts={listData?.severity_counts}
                total={listData?.total}
              />
            </div>
          </Sheet>
        </div>
      </div>

      <div className="technical-sheet overflow-hidden">
        <div className="table-header hidden md:grid md:grid-cols-[150px_minmax(0,1fr)_140px_220px_160px_40px] md:items-center">
          <div>Severity</div>
          <div>Problem</div>
          <div>Type</div>
          <div>Name</div>
          <div>Detected</div>
          <div />
        </div>

        <div>
          {!selectedInstanceId ? (
            <div className="px-6 py-8 text-sm text-muted-foreground">
              Select a connection in Settings to load diagnosis data.
            </div>
          ) : loadingList ? (
            <div className="flex h-full items-center justify-center p-8 text-sm text-muted-foreground">
              Loading findings...
            </div>
          ) : (
            filteredFindings.map((finding) => (
              <Link
                key={finding.id}
                to={`/diagnosis/${encodeURIComponent(finding.id)}`}
                className={`table-row grid gap-4 px-6 py-5 md:grid-cols-[150px_minmax(0,1fr)_140px_220px_160px_40px] md:items-center ${
                  focusedFindingId === finding.id ? "surface-selection" : ""
                }`}
              >
                <div className="space-y-2">
                  <div className="kicker md:hidden">Severity</div>
                  <Badge variant={finding.severity.toLowerCase() as "critical" | "warning" | "info"}>
                    {severityLabel(finding.severity)}
                  </Badge>
                </div>

                <div className="space-y-2">
                  <div className="kicker md:hidden">Problem</div>
                  <div className="font-heading text-[18px] font-semibold leading-6 tracking-[-0.02em] text-foreground">
                    {finding.title}
                  </div>
                </div>

                <div className="space-y-2">
                  <div className="kicker md:hidden">Type</div>
                  <span className="inline-flex rounded-md border-2 border-[#111111] bg-[#eef2ff] px-2.5 py-1 text-[11px] font-semibold tracking-[0.01em] text-[#254fd2]">
                    {finding.resource_type || "object"}
                  </span>
                </div>

                <div className="space-y-2">
                  <div className="kicker md:hidden">Name</div>
                  <div className="text-sm text-slate-600">{finding.resource_name}</div>
                </div>

                <div className="space-y-2">
                  <div className="kicker md:hidden">Detected</div>
                  <div className="text-sm text-slate-600">
                    {formatRelativeTime(finding.detected_at)}
                  </div>
                </div>

                <div className="hidden justify-self-end md:block">
                  <ArrowRight className="h-4 w-4 text-slate-700" />
                </div>
              </Link>
            ))
          )}

          {!loadingList && selectedInstanceId && filteredFindings.length === 0 ? (
            <div className="px-6 py-8 text-sm text-muted-foreground">
              {search.trim()
                ? "No findings match the current filter."
                : "No active findings are available for this database yet."}
            </div>
          ) : null}
        </div>
      </div>
    </div>
  );
}
