import React from 'react';

export const MumbaiMapBackground: React.FC = () => {
  return (
    <svg
      viewBox="0 0 1200 1600"
      className="absolute inset-0 h-full w-full opacity-20"
      style={{ pointerEvents: 'none' }}
      xmlns="http://www.w3.org/2000/svg"
    >
      <defs>
        <pattern id="gridPattern" width="80" height="80" patternUnits="userSpaceOnUse">
          <path d="M 80 0 L 0 0 0 80" fill="none" stroke="rgba(34, 211, 238, 0.12)" strokeWidth="0.5" />
        </pattern>
        <linearGradient id="arterialGradient" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stopColor="#22d3ee" />
          <stop offset="55%" stopColor="#64748b" />
          <stop offset="100%" stopColor="#10b981" />
        </linearGradient>
        <radialGradient id="nodeGlow" cx="50%" cy="50%" r="50%">
          <stop offset="0%" stopColor="rgba(34,211,238,0.55)" />
          <stop offset="100%" stopColor="rgba(34,211,238,0)" />
        </radialGradient>
      </defs>

      <rect width="1200" height="1600" fill="url(#gridPattern)" />

      <path d="M 200 50 Q 210 400 220 800 Q 225 1200 230 1600" stroke="url(#arterialGradient)" strokeWidth="8" fill="none" opacity="0.72" />
      <path d="M 400 0 Q 410 300 420 700 Q 425 1100 430 1600" stroke="url(#arterialGradient)" strokeWidth="6" fill="none" opacity="0.62" />
      <path d="M 600 20 L 610 1580" stroke="url(#arterialGradient)" strokeWidth="7" fill="none" opacity="0.7" />
      <path d="M 800 0 Q 815 500 820 1000 Q 825 1300 830 1600" stroke="url(#arterialGradient)" strokeWidth="6" fill="none" opacity="0.62" />
      <path d="M 1000 100 Q 1010 400 1015 900 Q 1020 1200 1025 1600" stroke="url(#arterialGradient)" strokeWidth="7" fill="none" opacity="0.7" />

      <path d="M 50 200 Q 300 210 600 200 Q 900 190 1150 200" stroke="url(#arterialGradient)" strokeWidth="7" fill="none" opacity="0.7" />
      <path d="M 30 400 Q 200 420 500 410 Q 800 400 1170 410" stroke="url(#arterialGradient)" strokeWidth="6" fill="none" opacity="0.62" />
      <path d="M 40 600 L 1160 600" stroke="url(#arterialGradient)" strokeWidth="7" fill="none" opacity="0.7" />
      <path d="M 50 800 Q 280 820 600 810 Q 920 800 1150 815" stroke="url(#arterialGradient)" strokeWidth="6" fill="none" opacity="0.62" />
      <path d="M 30 1000 Q 250 1020 600 1010 Q 950 1000 1170 1020" stroke="url(#arterialGradient)" strokeWidth="7" fill="none" opacity="0.7" />
      <path d="M 40 1200 Q 300 1210 600 1200 Q 900 1190 1160 1210" stroke="url(#arterialGradient)" strokeWidth="6" fill="none" opacity="0.62" />
      <path d="M 50 1400 Q 350 1420 650 1410 Q 950 1400 1150 1420" stroke="url(#arterialGradient)" strokeWidth="7" fill="none" opacity="0.7" />

      <path d="M 100 100 Q 400 600 800 1200 Q 1000 1400 1100 1550" stroke="rgba(34,211,238,0.35)" strokeWidth="5" fill="none" opacity="0.7" />
      <path d="M 1100 100 Q 700 500 350 1000 Q 100 1300 50 1550" stroke="rgba(16,185,129,0.3)" strokeWidth="5" fill="none" opacity="0.7" />
      <path d="M 200 300 Q 600 700 950 1100" stroke="rgba(56,189,248,0.28)" strokeWidth="4" fill="none" opacity="0.65" />
      <path d="M 1050 300 Q 650 700 300 1100" stroke="rgba(20,184,166,0.28)" strokeWidth="4" fill="none" opacity="0.65" />

      {[
        { x: 200, y: 200, r: 12 },
        { x: 600, y: 300, r: 15 },
        { x: 1000, y: 250, r: 10 },
        { x: 150, y: 600, r: 8 },
        { x: 600, y: 600, r: 18 },
        { x: 1050, y: 650, r: 12 },
        { x: 300, y: 1000, r: 10 },
        { x: 700, y: 1050, r: 14 },
        { x: 200, y: 1300, r: 9 },
        { x: 950, y: 1350, r: 11 },
      ].map((node, i) => (
        <circle
          key={i}
          cx={node.x}
          cy={node.y}
          r={node.r}
          fill="url(#nodeGlow)"
          stroke="rgba(34, 211, 238, 0.34)"
          strokeWidth="2"
        />
      ))}

      <path
        d="M 950 150 Q 1050 300 1100 500 Q 1080 700 1000 800 Q 900 750 900 550 Z"
        fill="rgba(34, 211, 238, 0.08)"
        stroke="rgba(34, 211, 238, 0.16)"
        strokeWidth="1.5"
        opacity="0.85"
      />
      <path
        d="M 100 1400 Q 150 1500 250 1550 Q 300 1480 250 1350 Z"
        fill="rgba(16, 185, 129, 0.08)"
        stroke="rgba(16, 185, 129, 0.18)"
        strokeWidth="1.5"
        opacity="0.85"
      />

      <path d="M 300 250 Q 500 450 650 600" stroke="rgba(34, 211, 238, 0.3)" strokeWidth="3" fill="none" />
      <path d="M 850 350 Q 700 500 600 700" stroke="rgba(16, 185, 129, 0.28)" strokeWidth="3" fill="none" />
      <path d="M 200 750 Q 400 850 550 1000" stroke="rgba(34, 211, 238, 0.26)" strokeWidth="3" fill="none" />
      <path d="M 900 900 Q 750 1050 650 1200" stroke="rgba(16, 185, 129, 0.26)" strokeWidth="3" fill="none" />

      <text x="50" y="1550" fontSize="10" fill="rgba(34, 211, 238, 0.48)" fontFamily="monospace">
        MUMBAI ARTERIAL NETWORK
      </text>
    </svg>
  );
};