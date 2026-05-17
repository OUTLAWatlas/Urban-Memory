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
  const statusMessage = verification?.status_message || 'Unknown';
  const isSecured = isVerified || statusMessage === 'Cryptographically Secured' || statusMessage === 'Verified';
  const isLoading = statusMessage === 'Verification Pending' || statusMessage === 'Verifying...';
  const isNotarized = statusMessage === 'Record Not Found' && !isVerified;

  return (
    <>
      <style>{`
        @keyframes shieldPulse {
          0% { box-shadow: 0 0 0 0 rgba(34, 197, 94, 0.45); }
          70% { box-shadow: 0 0 0 10px rgba(34, 197, 94, 0); }
          100% { box-shadow: 0 0 0 0 rgba(34, 197, 94, 0); }
        }
        @keyframes spin {
          0% { transform: rotate(0deg); }
          100% { transform: rotate(360deg); }
        }
      `}</style>

      {isSecured ? (
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
          title="Layer data is cryptographically verified and sealed on the blockchain"
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
      ) : isLoading ? (
        <div
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: '8px',
            border: '1px solid #7c2d12',
            background: 'rgba(120, 53, 15, 0.2)',
            color: '#fdba74',
            borderRadius: '999px',
            padding: '6px 12px',
            fontSize: '12px',
            fontWeight: 700,
          }}
          title="Verifying layer against blockchain ledger..."
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
              background: '#7c2d12',
              animation: 'spin 1s linear infinite',
            }}
          >
            ⟳
          </span>
          <span>Verifying...</span>
        </div>
      ) : isNotarized ? (
        <div
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: '8px',
            border: '1px solid #6b7280',
            background: 'rgba(107, 114, 128, 0.1)',
            color: '#d1d5db',
            borderRadius: '999px',
            padding: '6px 12px',
            fontSize: '12px',
            fontWeight: 700,
          }}
          title="Layer data exists but has not been notarized yet"
        >
          <span aria-hidden="true">◯</span>
          <span>Not Notarized</span>
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
          title="Layer data integrity issue or verification failed"
        >
          <span aria-hidden="true">⚠</span>
          <span>Integrity Issue</span>
        </div>
      )}
    </>
  );
}
