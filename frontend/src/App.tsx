import { Route, Routes } from 'react-router-dom';
import { HealthPage } from './pages/HealthPage';
import { IssuesPage } from './pages/IssuesPage';
import { IssueDetailPage } from './pages/IssueDetailPage';
import { QueriesPage } from './pages/QueriesPage';
import { TablesPage } from './pages/TablesPage';
import { MetricsPage } from './pages/MetricsPage';

function App() {
  return (
    <Routes>
      <Route path="/" element={<HealthPage />} />
      <Route path="/issues" element={<IssuesPage />} />
      <Route path="/issues/:id" element={<IssueDetailPage />} />
      <Route path="/queries" element={<QueriesPage />} />
      <Route path="/tables" element={<TablesPage />} />
      <Route path="/metrics" element={<MetricsPage />} />
    </Routes>
  );
}

export default App;
