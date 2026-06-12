import { useState } from 'react';
import type { ReactNode } from 'react';
import { Sidebar } from './Sidebar';
import { Header } from './Header';
import type { HeaderProps } from './Header';

interface LayoutProps extends Omit<HeaderProps, 'onHamburger'> {
  children: ReactNode;
}

export function Layout({ children, ...headerProps }: LayoutProps) {
  const [mobileOpen, setMobileOpen] = useState(false);

  return (
    <div className="flex h-screen overflow-hidden" style={{ background: 'var(--bg-app)' }}>
      <Sidebar mobileOpen={mobileOpen} onClose={() => setMobileOpen(false)} />
      <div className="flex h-full min-w-0 flex-1 flex-col">
        <Header {...headerProps} onHamburger={() => setMobileOpen(true)} />
        <main className="flex-1 overflow-y-auto" style={{ background: 'var(--bg-app)' }}>
          <div className="mx-auto px-6 py-6 pb-16" style={{ maxWidth: 'var(--content-max)' }}>
            {children}
          </div>
        </main>
      </div>
    </div>
  );
}
