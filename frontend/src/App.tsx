import { Route, Routes } from 'react-router-dom';
import { OverviewPage } from './pages/OverviewPage';
import { IssuesPage } from './pages/IssuesPage';
import { QueriesPage } from './pages/QueriesPage';
import { TablesPage } from './pages/TablesPage';
import { MetricsPage } from './pages/MetricsPage';

function App() {
  return (
    <Routes>
      <Route path="/" element={<OverviewPage />} />
      <Route path="/issues" element={<IssuesPage />} />
      <Route path="/queries" element={<QueriesPage />} />
      <Route path="/tables" element={<TablesPage />} />
      <Route path="/metrics" element={<MetricsPage />} />
    </Routes>
  );
}

export default App;
