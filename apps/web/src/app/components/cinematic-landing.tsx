'use client';

import React, { useRef } from 'react';
import Link from 'next/link';
import { motion, useScroll, useTransform } from 'framer-motion';
import { GisNavigationPin } from './gis-navigation-pin';
import { MumbaiMapBackground } from './mumbai-map-background';

export const CinematicLanding: React.FC = () => {
  const containerRef = useRef<HTMLDivElement>(null);
  const pinContainerRef = useRef<HTMLDivElement>(null);

  // Track window scroll
  const { scrollY } = useScroll();

  // Hero section animations
  const titleOpacity = useTransform(scrollY, [0, 300], [1, 0]);
  const subtitleOpacity = useTransform(scrollY, [0, 400], [1, 0]);

  // Pin animations - the match cut (the cinematic transform)
  const pinScale = useTransform(scrollY, [200, 800], [1, 2.5]);
  const pinRotation = useTransform(scrollY, [120, 780, 1100], [0, 720, 720]);
  const pinY = useTransform(scrollY, [120, 800, 1100], [0, 0, -700]);
  const pinX = useTransform(scrollY, [120, 800, 1100], [0, 0, -420]);

  // Dashboard content animations
  const contentY = useTransform(scrollY, [850, 1150], [28, 0]);

  // Map background - stays visible throughout
  const mapOpacity = useTransform(scrollY, [0, 1500], [1, 0.15]);

  return (
    <div
      ref={containerRef}
      className="relative w-full bg-zinc-950"
    >
      {/* === SCROLLABLE CONTENT === */}
      <div className="relative z-10">
        {/* === HERO SECTION === */}
        <section className="relative h-screen w-full flex items-center justify-center overflow-hidden">
          {/* Radial gradient overlays */}
          <div className="absolute inset-0">
            <div className="absolute top-0 left-1/2 transform -translate-x-1/2 w-96 h-96 bg-cyan-400/20 rounded-full filter blur-3xl" />
            <div className="absolute bottom-0 right-1/4 w-80 h-80 bg-emerald-400/10 rounded-full filter blur-3xl" />
          </div>

          {/* Hero content */}
          <div className="relative z-10 text-center px-6 max-w-4xl mx-auto">
            {/* Main title with pin replacement */}
            <motion.div className="mb-8">
              <h1 className="text-7xl lg:text-8xl font-bold tracking-tight mb-4">
                <span className="bg-clip-text text-transparent bg-gradient-to-r from-cyan-300 via-emerald-300 to-cyan-300">
                  Urban
                </span>
                <span className="text-white">Mem</span>
                <span className="inline-block relative">
                  <motion.div
                    ref={pinContainerRef}
                    className="relative inline-flex h-32 w-28 items-center justify-center lg:h-40 lg:w-36"
                    style={{
                      x: pinX,
                      y: pinY,
                      scale: pinScale,
                      perspective: 1200,
                    }}
                  >
                    <div className="absolute bottom-1 h-16 w-24 rounded-[999px] bg-zinc-950/80 ring-1 ring-cyan-400/15 shadow-[0_12px_30px_rgba(2,132,199,0.12)] backdrop-blur-sm lg:h-20 lg:w-28">
                      <div className="absolute inset-2 rounded-[999px] bg-[radial-gradient(circle_at_center,rgba(6,182,212,0.08),rgba(9,9,11,0.2)_70%,transparent_100%)]" />
                      <div className="absolute inset-[10px] rounded-[999px] border border-cyan-300/10" />
                      <div className="absolute left-1/2 top-1/2 h-[1px] w-[72%] -translate-x-1/2 -translate-y-1/2 bg-cyan-300/10" />
                      <div className="absolute left-1/2 top-1/2 h-[72%] w-[1px] -translate-x-1/2 -translate-y-1/2 bg-cyan-300/10" />
                      <svg
                        className="absolute inset-0 h-full w-full"
                        viewBox="0 0 120 80"
                        fill="none"
                        aria-hidden="true"
                      >
                        <path d="M16 55 C 28 48, 40 48, 52 56 S 76 64, 92 52" stroke="rgba(148,163,184,0.18)" strokeWidth="1.2" />
                        <path d="M20 30 H100" stroke="rgba(34,211,238,0.16)" strokeWidth="1" strokeDasharray="3 5" />
                        <path d="M35 18 V64" stroke="rgba(16,185,129,0.12)" strokeWidth="1" strokeDasharray="2 5" />
                        <path d="M60 14 V66" stroke="rgba(148,163,184,0.15)" strokeWidth="1" />
                        <path d="M84 20 V62" stroke="rgba(16,185,129,0.12)" strokeWidth="1" strokeDasharray="2 4" />
                        <circle cx="60" cy="40" r="8" stroke="rgba(34,211,238,0.18)" strokeWidth="1" />
                        <circle cx="60" cy="40" r="2" fill="rgba(34,211,238,0.22)" />
                      </svg>
                    </div>
                    <motion.div
                      className="absolute bottom-14 left-1/2 -translate-x-1/2 origin-center will-change-transform"
                      style={{
                        rotateY: pinRotation,
                        transformStyle: 'preserve-3d',
                        transformOrigin: 'center',
                      }}
                    >
                      <GisNavigationPin className="h-24 w-24 lg:h-32 lg:w-32 drop-shadow-[0_0_30px_rgba(34,211,238,0.42)]" />
                    </motion.div>
                    <motion.div
                      className="absolute bottom-4 left-1/2 h-3 w-20 -translate-x-1/2 rounded-full bg-cyan-300/15 blur-md"
                      animate={{ opacity: [0.24, 0.42, 0.24], scaleX: [1, 1.06, 1] }}
                      transition={{ repeat: Infinity, duration: 3.5, ease: 'easeInOut' }}
                    />
                  </motion.div>
                </span>
                <span className="text-white">ry</span>
              </h1>
            </motion.div>

            {/* Subtitle */}
            <motion.p
              className="text-lg lg:text-xl text-slate-300 max-w-2xl mx-auto leading-relaxed"
              style={{ opacity: subtitleOpacity }}
            >
              Municipal Geospatial Intelligence, Anchored in Cryptographic Integrity
            </motion.p>
          </div>

          {/* Scroll hint */}
          <motion.div
            className="absolute bottom-10 left-1/2 transform -translate-x-1/2 z-20"
            animate={{ y: [0, 12, 0] }}
            transition={{ duration: 2, repeat: Infinity }}
            style={{ opacity: titleOpacity }}
          >
            <div className="w-6 h-10 border-2 border-cyan-300 rounded-full flex justify-center">
              <motion.div
                className="w-1 h-2 bg-cyan-300 rounded-full mt-2"
                animate={{ y: [0, 8, 0] }}
                transition={{ duration: 2, repeat: Infinity }}
              />
            </div>
          </motion.div>
        </section>

        {/* === DASHBOARD CONTENT SECTION === */}
        <motion.section
          className="relative w-full px-6 lg:px-10 py-32 bg-zinc-950/90 backdrop-blur-sm"
          style={{ y: contentY }}
        >
          {/* Background elements */}
          <div className="absolute inset-0 overflow-hidden">
            <div className="absolute top-20 right-1/3 w-96 h-96 bg-emerald-400/5 rounded-full filter blur-3xl" />
            <div className="absolute bottom-40 left-1/4 w-80 h-80 bg-cyan-400/5 rounded-full filter blur-3xl" />
          </div>

          {/* Content */}
          <div className="relative z-10 max-w-7xl mx-auto">
            {/* Dashboard header */}
            <div
              className="mb-16"
            >
              <h2 className="text-5xl lg:text-6xl font-bold text-white mb-4">
                <span className="bg-clip-text text-transparent bg-gradient-to-r from-cyan-300 to-emerald-300">
                  Civic Intelligence
                </span>
              </h2>
              <p className="text-lg text-slate-400 max-w-2xl">
                Explore municipal data with cryptographic verification. Real-time ledger integrity,
                transparent governance workflows, and immutable audit trails.
              </p>
            </div>

            {/* Portal cards */}
            <div id="explorer" className="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-16 scroll-mt-28">
              {/* Public Explorer Portal */}
              <div
                className="group relative rounded-2xl border border-white/10 bg-white/5 p-8 backdrop-blur-xl overflow-hidden hover:border-cyan-400/50 transition-all duration-300"
              >
                {/* Hover glow */}
                <div className="absolute inset-0 bg-gradient-to-br from-cyan-400/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300" />

                <div className="relative z-10">
                  <div className="w-12 h-12 mb-4 rounded-lg bg-gradient-to-br from-cyan-400 to-emerald-400 flex items-center justify-center text-white font-bold text-xl">
                    🗺️
                  </div>
                  <h3 className="text-2xl font-bold text-white mb-2">Public Explorer Portal</h3>
                  <p className="text-slate-300 leading-relaxed mb-6">
                    Navigate municipal geospatial data with full transparency. Access zoning records,
                    electoral boundaries, and development records verified on-chain.
                  </p>
                  <ul className="space-y-3">
                    <li className="flex items-center gap-3 text-sm text-slate-400">
                      <span className="w-2 h-2 bg-cyan-400 rounded-full" />
                      Interactive cartography
                    </li>
                    <li className="flex items-center gap-3 text-sm text-slate-400">
                      <span className="w-2 h-2 bg-emerald-400 rounded-full" />
                      Real-time ledger verification
                    </li>
                    <li className="flex items-center gap-3 text-sm text-slate-400">
                      <span className="w-2 h-2 bg-cyan-400 rounded-full" />
                      Time-bound data sealing
                    </li>
                  </ul>
                </div>
              </div>

              {/* Governance Hub */}
              <div
                id="governance"
                className="group relative rounded-2xl border border-white/10 bg-white/5 p-8 backdrop-blur-xl overflow-hidden hover:border-emerald-400/50 transition-all duration-300"
              >
                {/* Hover glow */}
                <div className="absolute inset-0 bg-gradient-to-br from-emerald-400/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300" />

                <div className="relative z-10">
                  <div className="w-12 h-12 mb-4 rounded-lg bg-gradient-to-br from-emerald-400 to-cyan-400 flex items-center justify-center text-white font-bold text-xl">
                    🔐
                  </div>
                  <h3 className="text-2xl font-bold text-white mb-2">Governance Hub</h3>
                  <p className="text-slate-300 leading-relaxed mb-6">
                    Administrative workflows for notary approval, OTP sealing, and governance record
                    management. Maintain cryptographic integrity across all civic operations.
                  </p>
                  <ul className="space-y-3">
                    <li className="flex items-center gap-3 text-sm text-slate-400">
                      <span className="w-2 h-2 bg-emerald-400 rounded-full" />
                      Secure admin authentication
                    </li>
                    <li className="flex items-center gap-3 text-sm text-slate-400">
                      <span className="w-2 h-2 bg-cyan-400 rounded-full" />
                      Notary OTP sealing
                    </li>
                    <li className="flex items-center gap-3 text-sm text-slate-400">
                      <span className="w-2 h-2 bg-emerald-400 rounded-full" />
                      Audit trail tracking
                    </li>
                  </ul>
                </div>
              </div>
            </div>

            {/* Real-time ledger stats */}
            <div
              className="rounded-2xl border border-white/10 bg-white/5 p-8 backdrop-blur-xl"
            >
              <h3 className="text-2xl font-bold text-white mb-8">Real-Time Ledger Intelligence</h3>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
                {[
                  { label: 'Records Verified', value: '128.4K' },
                  { label: 'Blockchain Commits', value: '45.2K' },
                  { label: 'Data Integrity Score', value: '99.8%' },
                  { label: 'Active Participants', value: '2,847' },
                ].map((stat, idx) => (
                  <div
                    key={idx}
                    className="relative"
                  >
                    <div className="p-4 rounded-lg bg-white/5 border border-white/5 hover:border-cyan-400/30 transition-all">
                      <p className="text-xs uppercase tracking-wider text-slate-400 mb-2">{stat.label}</p>
                      <p className="text-2xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-cyan-300 to-emerald-300">
                        {stat.value}
                      </p>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            {/* CTA Section */}
            <div
              className="mt-16 flex flex-col sm:flex-row gap-6 justify-center"
            >
              <Link
                href="/map"
                className="px-8 py-4 rounded-full bg-gradient-to-r from-cyan-400 to-emerald-400 text-slate-950 font-bold text-lg hover:shadow-[0_0_40px_rgba(34,211,238,0.3)] transition-all duration-300 inline-flex items-center justify-center"
              >
                Enter Public Portal
              </Link>
              <Link
                href="/login"
                className="px-8 py-4 rounded-full border-2 border-cyan-400/50 text-cyan-300 font-bold text-lg hover:bg-cyan-400/10 transition-all duration-300 inline-flex items-center justify-center"
              >
                Admin Access
              </Link>
            </div>
          </div>
        </motion.section>

        {/* Additional scroll padding */}
        <div className="h-96 bg-zinc-950" />
      </div>
    </div>
  );
};
