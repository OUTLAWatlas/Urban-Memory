type VerificationMetadata = {
  is_verified?: boolean;
  on_chain_hash?: string;
  status_message?: string;
};

type IntegrityBadgeProps = {
  verification?: VerificationMetadata;
};

export default function IntegrityBadge({ verification }: IntegrityBadgeProps) {
  const isVerified = Boolean(verification?.is_verified);

  return (
    <>
      <style jsx>{`
        @keyframes shieldPulse {
          0% { box-shadow: 0 0 0 0 rgba(34, 197, 94, 0.45); }
          70% { box-shadow: 0 0 0 10px rgba(34, 197, 94, 0); }
          100% { box-shadow: 0 0 0 0 rgba(34, 197, 94, 0); }
        }
      `}</style>

      {isVerified ? (
        <div
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: '8px',
            border: '1px solid #166534',
            background: 'rgba(22, 101, 52, 0.2)',
            color: '#bbf7d0',
            borderRadius: '999px',
            padding: '6px 12px',
            fontSize: '12px',
            fontWeight: 700,
          }}
        >
          <span
            aria-hidden="true"
            style={{
              display: 'inline-flex',
              width: '20px',
              height: '20px',
              borderRadius: '999px',
              alignItems: 'center',
              justifyContent: 'center',
              background: '#166534',
              animation: 'shieldPulse 1.6s infinite',
            }}
          >
            🛡
          </span>
          <span>Cryptographically Secured</span>
        </div>
      ) : (
        <div
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: '8px',
            border: '1px solid #991b1b',
            background: 'rgba(127, 29, 29, 0.28)',
            color: '#fecaca',
            borderRadius: '999px',
            padding: '6px 12px',
            fontSize: '12px',
            fontWeight: 700,
          }}
        >
          <span aria-hidden="true">⚠</span>
          <span>Integrity Mismatch</span>
        </div>
      )}
    </>
  );
}
