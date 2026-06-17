import { ChevronLeft } from "lucide-react";
import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { FindingDetail } from "@/components/finding-detail";
import { DismissibleAlert } from "@/components/ui/dismissible-alert";
import { getFinding } from "@/lib/api";
import { useWorkspace } from "@/lib/workspace-context";
import type { FindingDetailResponse } from "@/types/api";

export function DiagnosisDetailPage() {
  const { findingId } = useParams<{ findingId: string }>();
  const { selectedInstanceId } = useWorkspace();
  const [detail, setDetail] = useState<FindingDetailResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!findingId) {
      setDetail(null);
      return;
    }

    if (!selectedInstanceId) {
      setDetail(null);
      setLoading(false);
      return;
    }

    setLoading(true);
    setError(null);

    void getFinding(findingId, selectedInstanceId)
      .then((response) => {
        setDetail(response);
      })
      .catch((caught) => {
        setError(caught instanceof Error ? caught.message : "Failed to load diagnosis detail");
        setDetail(null);
      })
      .finally(() => {
        setLoading(false);
      });
  }, [findingId, selectedInstanceId]);

  const backHref = findingId
    ? `/diagnosis?focus=${encodeURIComponent(findingId)}`
    : "/diagnosis";

  return (
    <div className="mx-auto max-w-[1600px] space-y-4">
      <Link
        to={backHref}
        className="inline-flex items-center gap-2 rounded-md border-2 border-[#111111] bg-white px-3 py-2 text-[14px] font-semibold text-slate-700 shadow-[2px_2px_0_#111111] transition-all hover:-translate-x-[1px] hover:-translate-y-[1px] hover:bg-[#f4f1e6] hover:text-slate-950 hover:shadow-[3px_3px_0_#111111]"
      >
        <ChevronLeft className="h-4 w-4" />
        Back to Diagnosis
      </Link>

      {error ? <DismissibleAlert>{error}</DismissibleAlert> : null}

      <FindingDetail detail={detail} loading={loading} />
    </div>
  );
}
