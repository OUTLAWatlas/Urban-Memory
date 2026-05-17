'use client';

import { useEffect, useMemo, useRef, useState } from 'react';
import type { AdminSession } from './session-provider';

type OtpModalProps = {
  open: boolean;
  adminSession: AdminSession | null;
  layerType: string;
  year: number;
  city?: string;
  onClose: () => void;
  onSuccess: (result: { message?: string; tx_hash?: string; sha256_hash?: string }) => void;
};

type RequestState = 'idle' | 'requesting' | 'code' | 'success' | 'error' | 'expired';

const OTP_LENGTH = 6;
const OTP_WINDOW_SECONDS = 5 * 60;

function formatSeconds(totalSeconds: number) {
  const minutes = Math.floor(totalSeconds / 60)
    .toString()
    .padStart(2, '0');
  const seconds = Math.max(totalSeconds % 60, 0).toString().padStart(2, '0');
  return `${minutes}:${seconds}`;
}

export default function OtpModal({ open, adminSession, layerType, year, city = 'Mumbai', onClose, onSuccess }: OtpModalProps) {
  const [step, setStep] = useState<RequestState>('idle');
  const [digits, setDigits] = useState<string[]>(() => Array.from({ length: OTP_LENGTH }, () => ''));
  const [countdown, setCountdown] = useState(OTP_WINDOW_SECONDS);
  const [statusMessage, setStatusMessage] = useState('');
  const [errorMessage, setErrorMessage] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const inputsRef = useRef<Array<HTMLInputElement | null>>([]);

  const otpCode = useMemo(() => digits.join(''), [digits]);
  const isComplete = digits.every((digit) => digit !== '');

  useEffect(() => {
    if (!open) {
      setStep('idle');
      setDigits(Array.from({ length: OTP_LENGTH }, () => ''));
      setCountdown(OTP_WINDOW_SECONDS);
      setStatusMessage('');
      setErrorMessage('');
      setIsSubmitting(false);
      return;
    }

    let cancelled = false;

    const requestNotary = async () => {
      setStep('requesting');
      setStatusMessage('A security challenge code has been dispatched to your verified channel.');
      setErrorMessage('');
      setDigits(Array.from({ length: OTP_LENGTH }, () => ''));
      setCountdown(OTP_WINDOW_SECONDS);

      if (!adminSession?.token || !adminSession?.adminId) {
        setStep('error');
        setErrorMessage('Admin session is required to request a notary challenge.');
        return;
      }

      try {
        const response = await fetch('/api/v1/admin/request-notary', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${adminSession.token}`,
          },
          body: JSON.stringify({
            layer_type: layerType,
            year,
          }),
        });

        const payload = await response.json().catch(() => ({}));

        if (cancelled) {
          return;
        }

        if (!response.ok) {
          setStep('error');
          setErrorMessage(String(payload?.error ?? payload?.message ?? 'Unable to dispatch OTP challenge'));
          return;
        }

        setStep('code');
        setStatusMessage(String(payload?.message ?? 'A security challenge code has been dispatched to your verified channel.'));
      } catch (error) {
        if (cancelled) {
          return;
        }

        setStep('error');
        setErrorMessage(String(error));
      }
    };

    void requestNotary();

    return () => {
      cancelled = true;
    };
  }, [adminSession?.adminId, adminSession?.token, layerType, open, year]);

  useEffect(() => {
    if (!open || step !== 'code') {
      return;
    }

    const timer = window.setInterval(() => {
      setCountdown((current) => {
        if (current <= 1) {
          window.clearInterval(timer);
          setStep('expired');
          setErrorMessage('The OTP window has expired. Request a fresh challenge.');
          return 0;
        }

        return current - 1;
      });
    }, 1000);

    return () => window.clearInterval(timer);
  }, [open, step]);

  useEffect(() => {
    if (step !== 'success') {
      return;
    }

    const timer = window.setTimeout(() => {
      onClose();
    }, 900);

    return () => window.clearTimeout(timer);
  }, [onClose, step]);

  const updateDigit = (index: number, value: string) => {
    const nextValue = value.replace(/\D/g, '').slice(-1);
    setDigits((current) => {
      const next = [...current];
      next[index] = nextValue;
      return next;
    });

    if (nextValue && index < OTP_LENGTH - 1) {
      inputsRef.current[index + 1]?.focus();
    }
  };

  const handleKeyDown = (index: number, event: React.KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'Backspace' && !digits[index] && index > 0) {
      inputsRef.current[index - 1]?.focus();
    }
  };

  const handleSubmit = async () => {
    if (!adminSession?.token || !adminSession?.adminId) {
      setErrorMessage('Admin session is required to confirm notarization.');
      return;
    }

    if (!isComplete || step !== 'code' || countdown <= 0) {
      return;
    }

    setIsSubmitting(true);
    setErrorMessage('');

    try {
      const response = await fetch('/api/v1/admin/confirm-notary', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${adminSession.token}`,
        },
        body: JSON.stringify({
          layer_type: layerType,
          year,
          admin_id: adminSession.adminId,
          otp_code: otpCode,
          city,
        }),
      });

      const payload = await response.json().catch(() => ({}));

      if (!response.ok) {
        setErrorMessage(String(payload?.error ?? payload?.message ?? 'Seal confirmation failed'));
        return;
      }

      setStep('success');
      setStatusMessage(String(payload?.message ?? 'Layer history sealed successfully.'));
      onSuccess({
        message: String(payload?.message ?? 'Layer history sealed successfully.'),
        tx_hash: payload?.tx_hash,
        sha256_hash: payload?.sha256_hash,
      });
    } catch (error) {
      setErrorMessage(String(error));
    } finally {
      setIsSubmitting(false);
    }
  };

  if (!open) {
    return null;
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center px-4 py-8">
      <div className="absolute inset-0 bg-slate-950/80 backdrop-blur-2xl" onClick={step === 'success' ? onClose : undefined} />

      <div className="relative z-10 w-full max-w-xl rounded-[2rem] border border-white/10 bg-slate-950/80 p-6 text-white shadow-[0_30px_120px_rgba(0,0,0,0.6)] ring-1 ring-cyan-400/10 md:p-8">
        <div className="flex items-start justify-between gap-4">
          <div>
            <p className="text-xs uppercase tracking-[0.35em] text-cyan-300/80">Notary Challenge</p>
            <h2 className="mt-2 text-2xl font-semibold">Seal layer history</h2>
            <p className="mt-2 max-w-lg text-sm text-slate-300">
              {step === 'requesting'
                ? 'Dispatching a time-bound challenge to your verified admin channel.'
                : step === 'code'
                  ? 'Enter the 6-digit code to authorize the notarization seal.'
                  : step === 'expired'
                    ? 'The challenge expired before confirmation was completed.'
                    : step === 'success'
                      ? 'Seal confirmed. The session will close automatically.'
                      : 'A secure administrative step is required before the selected layer can be sealed.'}
            </p>
          </div>

          <button
            type="button"
            onClick={onClose}
            className="rounded-full border border-white/10 px-3 py-1 text-sm text-slate-300 transition hover:border-white/20 hover:text-white"
          >
            Close
          </button>
        </div>

        <div className="mt-6 rounded-3xl border border-white/10 bg-white/5 p-5">
          <div className="flex flex-wrap items-center gap-3 text-sm text-slate-300">
            <span className="rounded-full border border-cyan-400/20 bg-cyan-400/10 px-3 py-1 text-cyan-200">
              {layerType.replace(/_/g, ' ')}
            </span>
            <span className="rounded-full border border-white/10 bg-white/5 px-3 py-1">Year {year}</span>
            <span className="rounded-full border border-white/10 bg-white/5 px-3 py-1">OTP {formatSeconds(countdown)}</span>
          </div>

          {step === 'requesting' && (
            <div className="mt-6 flex items-center gap-4 rounded-2xl border border-cyan-400/20 bg-cyan-400/10 px-4 py-4 text-cyan-100">
              <div className="h-10 w-10 animate-pulse rounded-full bg-cyan-400/25" />
              <div>
                <p className="font-medium">Dispatching challenge</p>
                <p className="text-sm text-cyan-100/80">A security challenge code has been dispatched to your verified channel.</p>
              </div>
            </div>
          )}

          {step === 'code' && (
            <div className="mt-6 space-y-5">
              <div>
                <p className="text-sm font-medium text-emerald-200">A security challenge code has been dispatched to your verified channel.</p>
                <p className="mt-1 text-sm text-slate-400">The code expires in {formatSeconds(countdown)}.</p>
              </div>

              <div className="grid grid-cols-6 gap-2 sm:gap-3">
                {digits.map((digit, index) => (
                  <input
                    key={index}
                    ref={(element) => {
                      inputsRef.current[index] = element;
                    }}
                    inputMode="numeric"
                    autoComplete="one-time-code"
                    maxLength={1}
                    value={digit}
                    onChange={(event) => updateDigit(index, event.target.value)}
                    onKeyDown={(event) => handleKeyDown(index, event)}
                    className="h-14 rounded-2xl border border-white/10 bg-slate-900/90 text-center text-2xl font-semibold text-white outline-none transition placeholder:text-slate-600 focus:border-cyan-400/60 focus:bg-slate-900 focus:ring-2 focus:ring-cyan-400/20"
                  />
                ))}
              </div>

              <div className="flex flex-wrap items-center justify-between gap-3">
                <p className="text-sm text-slate-400">This action locks the visible layer history and updates the integrity badge in real time.</p>
                <button
                  type="button"
                  onClick={handleSubmit}
                  disabled={!isComplete || isSubmitting || countdown <= 0}
                  className="rounded-full bg-gradient-to-r from-emerald-400 to-cyan-400 px-5 py-3 text-sm font-semibold text-slate-950 transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-40"
                >
                  {isSubmitting ? 'Confirming...' : 'Confirm Seal'}
                </button>
              </div>
            </div>
          )}

          {step === 'expired' && (
            <div className="mt-6 rounded-2xl border border-amber-400/20 bg-amber-400/10 px-4 py-4 text-amber-100">
              <p className="font-medium">Challenge expired</p>
              <p className="mt-1 text-sm text-amber-100/80">Request a new OTP challenge to continue the notarization flow.</p>
            </div>
          )}

          {step === 'error' && (
            <div className="mt-6 rounded-2xl border border-rose-400/20 bg-rose-400/10 px-4 py-4 text-rose-100">
              <p className="font-medium">Unable to continue</p>
              <p className="mt-1 text-sm text-rose-100/80">{errorMessage}</p>
            </div>
          )}

          {step === 'success' && (
            <div className="mt-6 rounded-2xl border border-emerald-400/20 bg-emerald-400/10 px-4 py-4 text-emerald-100">
              <p className="font-medium">Seal confirmed</p>
              <p className="mt-1 text-sm text-emerald-100/80">{statusMessage}</p>
            </div>
          )}

          {step !== 'code' && step !== 'requesting' && statusMessage && step !== 'success' && (
            <p className="mt-4 text-sm text-slate-400">{statusMessage}</p>
          )}
        </div>
        <div className="mt-5 flex items-center justify-between text-xs uppercase tracking-[0.3em] text-slate-500">
          <span>Verified session required</span>
          <span>5 minute window</span>
        </div>
      </div>
    </div>
  );
}
