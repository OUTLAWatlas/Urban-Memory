'use client';

import { usePathname } from 'next/navigation';

export function AppShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const isHome = pathname === '/';

  return <main className={isHome ? 'pt-0' : 'pt-16'}>{children}</main>;
}