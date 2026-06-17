import { Navigate, Route, Routes } from "react-router-dom";
import { AppShell } from "@/components/app-shell";
import { WorkspaceProvider } from "@/lib/workspace-context";
import { ConnectionDetailPage } from "@/pages/ConnectionDetailPage";
import { DiagnosisDetailPage } from "@/pages/DiagnosisDetailPage";
import { DiagnosisPage } from "@/pages/DiagnosisPage";
import { HistoryPage } from "@/pages/HistoryPage";
import { QueryDetailPage } from "@/pages/QueryDetailPage";
import { QueryExplorerPage } from "@/pages/QueryExplorerPage";
import { SetupPage } from "@/pages/SetupPage";

export default function App() {
  return (
    <WorkspaceProvider>
      <Routes>
        <Route element={<AppShell />}>
          <Route index element={<Navigate to="/diagnosis" replace />} />
          <Route path="/diagnosis" element={<DiagnosisPage />} />
          <Route path="/diagnosis/:findingId" element={<DiagnosisDetailPage />} />
          <Route path="/history" element={<HistoryPage />} />
          <Route path="/queries" element={<QueryExplorerPage />} />
          <Route path="/queries/:queryId" element={<QueryDetailPage />} />
          <Route path="/setup" element={<SetupPage />} />
          <Route path="/setup/:sourceId" element={<ConnectionDetailPage />} />
        </Route>
      </Routes>
    </WorkspaceProvider>
  );
}
