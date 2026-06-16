import { Navigate, Route, Routes } from "react-router-dom";
import { AppShell } from "@/components/app-shell";
import { WorkspaceProvider } from "@/lib/workspace-context";
import { DiagnosisPage } from "@/pages/DiagnosisPage";
import { HistoryPage } from "@/pages/HistoryPage";
import { QueryExplorerPage } from "@/pages/QueryExplorerPage";
import { SetupPage } from "@/pages/SetupPage";

export default function App() {
  return (
    <WorkspaceProvider>
      <Routes>
        <Route element={<AppShell />}>
          <Route index element={<Navigate to="/diagnosis" replace />} />
          <Route path="/diagnosis" element={<DiagnosisPage />} />
          <Route path="/history" element={<HistoryPage />} />
          <Route path="/queries" element={<QueryExplorerPage />} />
          <Route path="/setup" element={<SetupPage />} />
        </Route>
      </Routes>
    </WorkspaceProvider>
  );
}
