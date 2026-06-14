import type { ReactNode } from 'react';
import { Database } from 'lucide-react';
import { Sidebar } from './Sidebar';
import { Header } from './Header';
import type { HeaderProps } from './Header';
import { Card } from './Card';
import { useDatabaseInstance } from '../lib/databaseInstance';

interface LayoutProps extends HeaderProps {
  children: ReactNode;
}

export function Layout({ children, ...headerProps }: LayoutProps) {
  const { instances, loading } = useDatabaseInstance();

  const noInstances = !loading && instances.length === 0;

  return (
    <div className="flex h-screen overflow-hidden" style={{ background: 'var(--bg-app)' }}>
      <Sidebar />
      <div className="flex h-full min-w-0 flex-1 flex-col">
        <Header {...headerProps} />
        <main className="flex-1 overflow-y-auto" style={{ background: 'var(--bg-app)' }}>
          <div className="mx-auto px-6 py-5 pb-16" style={{ maxWidth: 'var(--content-max)' }}>
            {noInstances ? (
              <Card title="No database connected">
                <div className="flex flex-col items-center gap-3 py-10 text-center">
                  <Database size={24} style={{ color: 'var(--text-muted)' }} />
                  <p className="m-0 max-w-[420px] text-sm" style={{ color: 'var(--text-muted)' }}>
                    No database instances are registered yet. Start a Postgresome agent
                    pointed at this API to begin collecting evidence and recommendations.
                  </p>
                </div>
              </Card>
            ) : (
              children
            )}
          </div>
        </main>
      </div>
    </div>
  );
}
