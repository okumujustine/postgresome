import { Navigate, Route, Routes } from 'react-router-dom';
import { OverviewPage } from './pages/OverviewPage';
import { DiagnosisPage } from './pages/DiagnosisPage';
import { QueriesPage } from './pages/QueriesPage';
import { ObjectsPage } from './pages/ObjectsPage';
import { HistoryPage } from './pages/HistoryPage';
import { MetricsPage } from './pages/MetricsPage';
import { SettingsPage } from './pages/SettingsPage';

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<OverviewPage />} />
      <Route path="/findings" element={<DiagnosisPage />} />
      <Route path="/findings/:id" element={<DiagnosisPage />} />
      <Route path="/diagnosis" element={<DiagnosisPage />} />
      <Route path="/diagnosis/:id" element={<DiagnosisPage />} />
      <Route path="/issues" element={<Navigate to="/findings" replace />} />
      <Route path="/issues/:id" element={<DiagnosisPage />} />
      <Route path="/queries" element={<QueriesPage />} />
      <Route path="/database" element={<ObjectsPage />} />
      <Route path="/objects" element={<Navigate to="/database" replace />} />
      <Route path="/tables" element={<Navigate to="/database" replace />} />
      <Route path="/history" element={<HistoryPage />} />
      <Route path="/metrics" element={<MetricsPage />} />
      <Route path="/settings" element={<SettingsPage />} />
    </Routes>
  );
}
