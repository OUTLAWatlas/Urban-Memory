import React from 'react';

interface GisNavigationPinProps {
  className?: string;
}

export const GisNavigationPin: React.FC<GisNavigationPinProps> = ({ className = '' }) => {
  return (
    <svg
      viewBox="0 0 100 140"
      className={`${className}`}
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <defs>
        <linearGradient id="pinGradient" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stopColor="#06B6D4" stopOpacity="1" />
          <stop offset="100%" stopColor="#10B981" stopOpacity="0.8" />
        </linearGradient>
        <filter id="pinGlow">
          <feGaussianBlur stdDeviation="3" result="coloredBlur" />
          <feMerge>
            <feMergeNode in="coloredBlur" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>
        <filter id="innerGlow">
          <feGaussianBlur stdDeviation="2" result="coloredBlur" />
          <feMerge>
            <feMergeNode in="coloredBlur" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>
      </defs>

      {/* Outer glow ring */}
      <circle
        cx="50"
        cy="50"
        r="48"
        fill="none"
        stroke="url(#pinGradient)"
        strokeWidth="1"
        opacity="0.4"
        filter="url(#pinGlow)"
      />

      {/* Main pin body */}
      <path
        d="M 50 20 C 64.32 20 75.5 31.18 75.5 45.5 C 75.5 60.5 50 110 50 110 C 50 110 24.5 60.5 24.5 45.5 C 24.5 31.18 35.68 20 50 20 Z"
        fill="url(#pinGradient)"
        filter="url(#pinGlow)"
      />

      {/* Inner highlight */}
      <path
        d="M 50 20 C 64.32 20 75.5 31.18 75.5 45.5 C 75.5 60.5 50 110 50 110 C 50 110 24.5 60.5 24.5 45.5 C 24.5 31.18 35.68 20 50 20 Z"
        fill="none"
        stroke="#ffffff"
        strokeWidth="0.8"
        opacity="0.5"
      />

      {/* Center circle - the "o" replacement */}
      <circle
        cx="50"
        cy="45"
        r="16"
        fill="rgba(255, 255, 255, 0.1)"
        stroke="url(#pinGradient)"
        strokeWidth="2"
        filter="url(#innerGlow)"
      />

      {/* Crosshair - GIS element */}
      <line
        x1="42"
        y1="45"
        x2="58"
        y2="45"
        stroke="url(#pinGradient)"
        strokeWidth="1.5"
        opacity="0.7"
      />
      <line
        x1="50"
        y1="37"
        x2="50"
        y2="53"
        stroke="url(#pinGradient)"
        strokeWidth="1.5"
        opacity="0.7"
      />

      {/* Decorative rings */}
      <circle
        cx="50"
        cy="45"
        r="22"
        fill="none"
        stroke="url(#pinGradient)"
        strokeWidth="0.5"
        opacity="0.3"
      />
    </svg>
  );
};
