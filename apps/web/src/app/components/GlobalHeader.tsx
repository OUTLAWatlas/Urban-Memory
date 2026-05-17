'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { motion, useScroll, useTransform } from 'framer-motion';
import { GisNavigationPin } from './gis-navigation-pin';
import ProfileControl from './ProfileControl';

export function GlobalHeader() {
  const pathname = usePathname();
  const isHome = pathname === '/';
  const { scrollY } = useScroll();

  const homeOpacity = useTransform(scrollY, [900, 1150], [0, 1]);
  const homeY = useTransform(scrollY, [900, 1150], [12, 0]);

  const headerStyle = isHome
    ? {
        opacity: homeOpacity,
        y: homeY,
      }
    : {
        opacity: 1,
        y: 0,
      };

  return (
    <motion.header
      className="fixed top-0 left-0 right-0 z-[100] h-16 flex items-center justify-between px-6 backdrop-blur-md bg-zinc-950/70 border-b border-white/10"
      style={headerStyle}
    >
      <div className="flex items-center gap-3">
        <Link href="/" className="flex items-center gap-3 rounded-full transition hover:bg-white/5">
          <span className="flex h-9 w-9 items-center justify-center">
            <GisNavigationPin className="h-8 w-8 drop-shadow-[0_0_18px_rgba(34,211,238,0.35)]" />
          </span>
          <span className="bg-clip-text text-sm font-semibold tracking-[0.35em] text-transparent bg-gradient-to-r from-cyan-300 to-emerald-300">
            URBANMEMORY
          </span>
        </Link>
      </div>

      <div className="flex items-center gap-3">
        <ProfileControl inline />
      </div>
    </motion.header>
  );
}