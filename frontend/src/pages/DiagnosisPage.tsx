import { useEffect, useState } from "react";
import { ListFilter } from "lucide-react";
import { useSearchParams } from "react-router-dom";
import { FindingDetail } from "@/components/finding-detail";
import { FindingList } from "@/components/finding-list";
import { Button } from "@/components/ui/button";
import { Sheet } from "@/components/ui/sheet";
import { getFinding, listFindings } from "@/lib/api";
import { useWorkspace } from "@/lib/workspace-context";
import type { FindingDetailResponse, FindingListResponse } from "@/types/api";

export function DiagnosisPage() {
  const { selectedInstanceId, selectedSource, loading: workspaceLoading } = useWorkspace();
  const [listData, setListData] = useState<FindingListResponse | null>(null);
  const [detail, setDetail] = useState<FindingDetailResponse | null>(null);
  const [loadingList, setLoadingList] = useState(false);
  const [loadingDetail, setLoadingDetail] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState("");
  const [mobileQueueOpen, setMobileQueueOpen] = useState(false);
  const [searchParams, setSearchParams] = useSearchParams();

  const selectedFindingId = searchParams.get("finding");

  useEffect(() => {
    if (!selectedInstanceId) {
      setListData(null);
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
      })
      .finally(() => {
        setLoadingList(false);
      });
  }, [selectedInstanceId]);

  useEffect(() => {
    if (!listData?.findings.length) {
      setDetail(null);
      return;
    }

    const findingExists = selectedFindingId
      ? listData.findings.some((item) => item.id === selectedFindingId)
      : false;

    if (!findingExists) {
      setSearchParams(
        { finding: listData.findings[0].id },
        { replace: true },
      );
    }
  }, [listData, selectedFindingId, setSearchParams]);

  useEffect(() => {
    if (!selectedInstanceId || !selectedFindingId) {
      setDetail(null);
      return;
    }

    setLoadingDetail(true);

    void getFinding(selectedFindingId, selectedInstanceId)
      .then((response) => {
        setDetail(response);
      })
      .catch((caught) => {
        setError(caught instanceof Error ? caught.message : "Failed to load finding detail");
      })
      .finally(() => {
        setLoadingDetail(false);
      });
  }, [selectedFindingId, selectedInstanceId]);

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

  if (!workspaceLoading && !selectedSource) {
    return (
      <div className="technical-sheet p-8">
        <div className="kicker">Start here</div>
        <h2 className="mt-3 font-heading text-[30px] font-semibold tracking-[-0.02em] text-foreground">
          Connect a database before opening the diagnosis workspace.
        </h2>
        <p className="mt-3 max-w-2xl text-[15px] leading-7 text-slate-600">
          Setup creates the source record, validates the connection, and lets you run
          the first checkup. Once a source exists, Diagnosis becomes the default home.
        </p>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-[1600px] space-y-4">
      <div className="technical-sheet flex flex-wrap items-center justify-between gap-4 px-6 py-5">
        <div>
          <div className="kicker">Selected database</div>
          <div className="mt-2 font-heading text-[18px] font-semibold text-foreground">
            {selectedSource?.database.name}
          </div>
          <div className="meta mt-1">{selectedSource?.database.host}</div>
        </div>

        <div className="flex items-center gap-2 lg:hidden">
          <Sheet
            open={mobileQueueOpen}
            onOpenChange={setMobileQueueOpen}
            title="Diagnosis queue"
            trigger={
              <Button variant="outline">
                <ListFilter className="h-4 w-4" />
                Open queue
              </Button>
            }
          >
            <div className="h-[calc(100vh-100px)]">
              <FindingList
                findings={filteredFindings}
                selectedFindingId={selectedFindingId}
                onSelect={(findingId) => {
                  setSearchParams({ finding: findingId });
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

      {error ? (
        <div className="rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      ) : null}

      <div className="grid min-h-[720px] gap-4 lg:grid-cols-[320px_minmax(0,1fr)]">
        <div className="technical-sheet hidden overflow-hidden lg:block">
          {loadingList ? (
            <div className="flex h-full items-center justify-center p-8 text-sm text-muted-foreground">
              Loading findings...
            </div>
          ) : (
            <FindingList
              findings={filteredFindings}
              selectedFindingId={selectedFindingId}
              onSelect={(findingId) => setSearchParams({ finding: findingId })}
              search={search}
              onSearchChange={setSearch}
              counts={listData?.severity_counts}
              total={listData?.total}
            />
          )}
        </div>

        <FindingDetail detail={detail} loading={loadingDetail} />
      </div>
    </div>
  );
}
