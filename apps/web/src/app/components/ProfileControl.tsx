'use client';

import { useEffect, useMemo, useState } from 'react';
import { createPortal } from 'react-dom';
import { useSession } from './session-provider';

type RequestResponse = {
  message?: string;
  expires_at?: string;
  error?: string;
};

type PendingAdmin = {
  id: number;
  email: string;
  role: string;
  created_at: string;
};

type PendingAdminsResponse = {
  pending_admins?: PendingAdmin[];
  error?: string;
};

function useCountdown(expiresAt: string | null) {
  const [now, setNow] = useState(() => Date.now());

  useEffect(() => {
    if (!expiresAt) {
      return;
    }

    const timer = window.setInterval(() => {
      setNow(Date.now());
    }, 1000);

    return () => window.clearInterval(timer);
  }, [expiresAt]);

  return useMemo(() => {
    if (!expiresAt) {
      return 0;
    }

    const milliseconds = new Date(expiresAt).getTime() - now;
    return Math.max(0, Math.floor(milliseconds / 1000));
  }, [expiresAt, now]);
}

function formatCountdown(totalSeconds: number) {
  const minutes = Math.floor(totalSeconds / 60)
    .toString()
    .padStart(2, '0');
  const seconds = Math.max(totalSeconds % 60, 0)
    .toString()
    .padStart(2, '0');
  return `${minutes}:${seconds}`;
}

export default function ProfileControl({ inline = false }: { inline?: boolean } = {}) {
  const { session, clearSession } = useSession();
  const admin = session.admin;
  const isAdmin = session.mode === 'admin' && Boolean(admin?.token);
  const isSuperAdmin = admin?.role === 'super_admin';

  const [isOpen, setIsOpen] = useState(false);
  const [activeTab, setActiveTab] = useState<'profile' | 'governance'>('profile');

  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [profileOtpCode, setProfileOtpCode] = useState('');
  const [profileExpiresAt, setProfileExpiresAt] = useState<string | null>(null);
  const [profileBusy, setProfileBusy] = useState(false);
  const [profileNotice, setProfileNotice] = useState('');
  const [profileError, setProfileError] = useState('');

  const [pendingAdmins, setPendingAdmins] = useState<PendingAdmin[]>([]);
  const [pendingBusy, setPendingBusy] = useState(false);
  const [governanceBusy, setGovernanceBusy] = useState(false);
  const [governanceNotice, setGovernanceNotice] = useState('');
  const [governanceError, setGovernanceError] = useState('');

  const profileSecondsLeft = useCountdown(profileExpiresAt);

  useEffect(() => {
    if (!profileExpiresAt || profileSecondsLeft > 0) {
      return;
    }

    setProfileExpiresAt(null);
    setProfileOtpCode('');
    setProfileError('Password rotation window expired. Request a new verification code.');
  }, [profileExpiresAt, profileSecondsLeft]);

  useEffect(() => {
    if (!isOpen || !isSuperAdmin) {
      return;
    }

    const fetchPendingAdmins = async () => {
      setPendingBusy(true);
      setGovernanceError('');

      try {
        console.log('[ADMIN_QUEUE] 📡 Fetching pending administrative applications...');

        const baseUrl = (process.env.NEXT_PUBLIC_API_URL || 'http://localhost:4000').replace(/\/$/, '');
        const headers: HeadersInit = {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${admin?.token ?? ''}`,
        };
        if (process.env.NEXT_PUBLIC_ADMIN_API_KEY) {
          headers['X-Admin-Key'] = process.env.NEXT_PUBLIC_ADMIN_API_KEY;
        }

        const response = await fetch(`${baseUrl}/api/v1/admin/pending-users`, {
          headers,
        });
        const payload = (await response.json().catch(() => ({}))) as PendingAdminsResponse;

        if (!response.ok) {
          setGovernanceError(String(payload.error ?? 'Unable to load pending admins.'));
          return;
        }

        console.log('[ADMIN_DASHBOARD] Raw API payload received:', payload);

        // Multi-layered data format guard
        if (Array.isArray(payload)) {
          // Direct array response
          setPendingAdmins(payload);
        } else if (Array.isArray(payload.pending_admins)) {
          // Wrapped in pending_admins property
          setPendingAdmins(payload.pending_admins);
        } else {
          console.error('[ADMIN_DASHBOARD] 🔴 Invalid data type returned. Expected array, received:', payload);
          setPendingAdmins([]);
        }
      } catch (error) {
        console.error('[ADMIN_QUEUE] 🔴 Failed to fetch registrations:', error);
        setGovernanceError(String(error));
      } finally {
        setPendingBusy(false);
      }
    };

    void fetchPendingAdmins();
  }, [admin?.token, isOpen, isSuperAdmin]);

  if (!isAdmin || !admin) {
    return null;
  }

  const closeModal = () => {
    setIsOpen(false);
    setActiveTab('profile');
    setProfileError('');
    setGovernanceError('');
  };

  const resetProfileFlow = () => {
    setCurrentPassword('');
    setNewPassword('');
    setConfirmPassword('');
    setProfileOtpCode('');
    setProfileExpiresAt(null);
  };

  const fetchPendingAdmins = async () => {
    if (!isSuperAdmin) {
      return;
    }

    setPendingBusy(true);
    try {
      console.log('[ADMIN_QUEUE] 📡 Fetching pending administrative applications...');

      const baseUrl = (process.env.NEXT_PUBLIC_API_URL || 'http://localhost:4000').replace(/\/$/, '');
      const headers: HeadersInit = {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${admin?.token ?? ''}`,
      };
      if (process.env.NEXT_PUBLIC_ADMIN_API_KEY) {
        headers['X-Admin-Key'] = process.env.NEXT_PUBLIC_ADMIN_API_KEY;
      }

      const response = await fetch(`${baseUrl}/api/v1/admin/pending-users`, {
        headers,
      });
      const payload = (await response.json().catch(() => ({}))) as PendingAdminsResponse;
      if (!response.ok) {
        setGovernanceError(String(payload.error ?? 'Unable to load pending admins.'));
        return;
      }

      console.log('[ADMIN_DASHBOARD] Raw API payload received:', payload);

      // Multi-layered data format guard
      if (Array.isArray(payload)) {
        // Direct array response
        setPendingAdmins(payload);
      } else if (Array.isArray(payload.pending_admins)) {
        // Wrapped in pending_admins property
        setPendingAdmins(payload.pending_admins);
      } else {
        console.error('[ADMIN_DASHBOARD] 🔴 Invalid data type returned. Expected array, received:', payload);
        setPendingAdmins([]);
      }
    } catch (error) {
      console.error('[ADMIN_QUEUE] 🔴 Failed to fetch registrations:', error);
      setGovernanceError(String(error));
    } finally {
      setPendingBusy(false);
    }
  };

  const handlePasswordChallenge = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setProfileBusy(true);
    setProfileError('');
    setProfileNotice('');

    try {
      const response = await fetch('/api/v1/admin/request-password-change', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${admin.token}`,
        },
        body: JSON.stringify({
          current_password: currentPassword,
          new_password: newPassword,
          confirm_new_password: confirmPassword,
        }),
      });

      const payload = (await response.json().catch(() => ({}))) as RequestResponse;
      if (!response.ok) {
        setProfileError(String(payload.error ?? 'Unable to start password rotation.'));
        return;
      }

      setProfileExpiresAt(String(payload.expires_at ?? ''));
      setProfileOtpCode('');
      setProfileNotice('Verification code dispatched. Confirm the rotation before the timer expires.');
    } catch (error) {
      setProfileError(String(error));
    } finally {
      setProfileBusy(false);
    }
  };

  const handlePasswordConfirmation = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setProfileBusy(true);
    setProfileError('');

    try {
      const response = await fetch('/api/v1/admin/confirm-password-change', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${admin.token}`,
        },
        body: JSON.stringify({
          otp_code: profileOtpCode,
        }),
      });

      const payload = (await response.json().catch(() => ({}))) as RequestResponse;
      if (!response.ok) {
        setProfileError(String(payload.error ?? 'Unable to confirm password rotation.'));
        return;
      }

      setProfileNotice('Password updated successfully.');
      resetProfileFlow();
    } catch (error) {
      setProfileError(String(error));
    } finally {
      setProfileBusy(false);
    }
  };

  const handleApprovePendingUser = async (e?: React.MouseEvent | React.FormEvent, email?: string) => {
    if (e && typeof e.preventDefault === 'function') {
      e.preventDefault();
    }
    
    if (!email) {
      setGovernanceError('Invalid approval request: missing email.');
      return;
    }

    setGovernanceBusy(true);
    setGovernanceError('');
    setGovernanceNotice('');

    try {
      const baseUrl = (process.env.NEXT_PUBLIC_API_URL || 'http://localhost:4000').replace(/\/$/, '');
      const response = await fetch(`${baseUrl}/api/v1/admin/approve-user`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${admin.token}`,
        },
        body: JSON.stringify({ email }),
      });

      const payload = (await response.json().catch(() => ({}))) as RequestResponse;
      if (!response.ok) {
        setGovernanceError(String(payload.error ?? 'Unable to approve pending admin.'));
        return;
      }

      setPendingAdmins((current) => current.filter((pendingAdmin) => pendingAdmin.email !== email));
      setGovernanceNotice('Admin account approved. Activation email dispatched.');
    } catch (error) {
      setGovernanceError(String(error));
    } finally {
      setGovernanceBusy(false);
    }
  };

  const modal = isOpen
    ? createPortal(
        <div className="fixed inset-0 z-[200] grid place-items-center p-4 sm:p-6">
          <div className="absolute inset-0 bg-slate-950/80 backdrop-blur-xl" onClick={closeModal} />

          <div className="relative z-10 flex max-h-[calc(100vh-2rem)] w-[min(94vw,68rem)] flex-col overflow-hidden rounded-[2rem] border border-white/10 bg-slate-950/88 text-white shadow-[0_40px_180px_rgba(0,0,0,0.58)] backdrop-blur-3xl sm:max-h-[calc(100vh-4rem)]">
            <div className="flex items-start justify-between gap-4 border-b border-white/10 px-6 py-5">
              <div>
                <p className="text-xs uppercase tracking-[0.35em] text-cyan-300/70">Administrative Profile Hub</p>
                <h2 className="mt-2 text-2xl font-semibold">Identity, credentials, and governance approvals</h2>
                <p className="mt-2 text-sm text-slate-400">{admin.email} · {admin.role}</p>
              </div>

              <div className="flex items-center gap-3">
                <button
                  type="button"
                  onClick={() => {
                    clearSession();
                    closeModal();
                  }}
                  className="rounded-full border border-white/10 px-4 py-2 text-sm text-slate-300 transition hover:border-rose-400/40 hover:text-white"
                >
                  End Session
                </button>
                <button
                  type="button"
                  onClick={closeModal}
                  className="rounded-full border border-white/10 px-4 py-2 text-sm text-slate-300 transition hover:border-white/30 hover:text-white"
                >
                  Close
                </button>
              </div>
            </div>

            <div className="flex border-b border-white/10 px-6">
              <button
                type="button"
                onClick={() => setActiveTab('profile')}
                className={`border-b-2 px-4 py-4 text-sm font-medium transition ${activeTab === 'profile' ? 'border-cyan-400 text-white' : 'border-transparent text-slate-400 hover:text-slate-200'}`}
              >
                Profile Settings
              </button>
              {isSuperAdmin && (
                <button
                  type="button"
                  onClick={() => setActiveTab('governance')}
                  className={`border-b-2 px-4 py-4 text-sm font-medium transition ${activeTab === 'governance' ? 'border-emerald-400 text-white' : 'border-transparent text-slate-400 hover:text-slate-200'}`}
                >
                  Governance Control Center
                </button>
              )}
            </div>

            <div className="flex-1 overflow-y-auto px-6 py-6">
              {activeTab === 'profile' && (
                <div className="grid gap-6 lg:grid-cols-[0.92fr_1.08fr]">
                  <div className="rounded-[1.5rem] border border-white/10 bg-white/5 p-5">
                    <p className="text-xs uppercase tracking-[0.3em] text-slate-500">Profile Settings</p>
                    <h3 className="mt-3 text-xl font-semibold">Rotate your access key with OTP confirmation</h3>
                    <p className="mt-3 text-sm leading-6 text-slate-400">
                      Submit your current password, define a replacement credential, and validate the change with a 5-minute email challenge.
                    </p>

                    <div className="mt-6 rounded-2xl border border-cyan-400/15 bg-cyan-400/10 px-4 py-4 text-sm text-cyan-100">
                      <p className="font-medium">Verification timer</p>
                      <p className="mt-1 text-cyan-100/80">
                        {profileExpiresAt ? `Code expires in ${formatCountdown(profileSecondsLeft)}.` : 'No active verification challenge.'}
                      </p>
                    </div>
                  </div>

                  <div className="space-y-5">
                    <form onSubmit={handlePasswordChallenge} className="rounded-[1.5rem] border border-white/10 bg-white/5 p-5">
                      <p className="text-xs uppercase tracking-[0.3em] text-slate-500">Credential Request</p>
                      <div className="mt-4 space-y-4">
                        <label className="block text-sm text-slate-300">
                          <span className="mb-2 block">Current Password</span>
                          <input
                            type="password"
                            value={currentPassword}
                            onChange={(event) => setCurrentPassword(event.target.value)}
                            className="w-full rounded-2xl border border-white/10 bg-slate-950/80 px-4 py-3 text-white outline-none transition focus:border-cyan-400/50 focus:ring-2 focus:ring-cyan-400/15"
                          />
                        </label>
                        <label className="block text-sm text-slate-300">
                          <span className="mb-2 block">New Password</span>
                          <input
                            type="password"
                            value={newPassword}
                            onChange={(event) => setNewPassword(event.target.value)}
                            className="w-full rounded-2xl border border-white/10 bg-slate-950/80 px-4 py-3 text-white outline-none transition focus:border-cyan-400/50 focus:ring-2 focus:ring-cyan-400/15"
                          />
                        </label>
                        <label className="block text-sm text-slate-300">
                          <span className="mb-2 block">Confirm New Password</span>
                          <input
                            type="password"
                            value={confirmPassword}
                            onChange={(event) => setConfirmPassword(event.target.value)}
                            className="w-full rounded-2xl border border-white/10 bg-slate-950/80 px-4 py-3 text-white outline-none transition focus:border-cyan-400/50 focus:ring-2 focus:ring-cyan-400/15"
                          />
                        </label>
                      </div>

                      <button
                        type="submit"
                        disabled={profileBusy}
                        className="mt-5 w-full rounded-full bg-gradient-to-r from-cyan-400 to-emerald-400 px-5 py-4 text-sm font-semibold text-slate-950 transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-50"
                      >
                        {profileBusy ? 'Requesting...' : 'Request Password Rotation OTP'}
                      </button>
                    </form>

                    {profileExpiresAt && (
                      <form onSubmit={handlePasswordConfirmation} className="rounded-[1.5rem] border border-white/10 bg-white/5 p-5">
                        <div className="flex items-center justify-between gap-3">
                          <div>
                            <p className="text-xs uppercase tracking-[0.3em] text-slate-500">Verification Window</p>
                            <h3 className="mt-2 text-lg font-semibold">Confirm password rotation</h3>
                          </div>
                          <span className="rounded-full border border-cyan-400/20 bg-cyan-400/10 px-3 py-1 text-sm text-cyan-100">
                            {formatCountdown(profileSecondsLeft)}
                          </span>
                        </div>

                        <label className="mt-4 block text-sm text-slate-300">
                          <span className="mb-2 block">6-digit OTP</span>
                          <input
                            inputMode="numeric"
                            maxLength={6}
                            value={profileOtpCode}
                            onChange={(event) => setProfileOtpCode(event.target.value.replace(/\D/g, '').slice(0, 6))}
                            className="w-full rounded-2xl border border-white/10 bg-slate-950/80 px-4 py-3 text-white outline-none transition focus:border-cyan-400/50 focus:ring-2 focus:ring-cyan-400/15"
                          />
                        </label>

                        <button
                          type="submit"
                          disabled={profileBusy || profileSecondsLeft === 0}
                          className="mt-5 w-full rounded-full border border-cyan-400/30 bg-cyan-400/10 px-5 py-4 text-sm font-semibold text-cyan-100 transition hover:bg-cyan-400/15 disabled:cursor-not-allowed disabled:opacity-50"
                        >
                          {profileBusy ? 'Confirming...' : 'Confirm Password Rotation'}
                        </button>
                      </form>
                    )}

                    {profileError && (
                      <div className="rounded-2xl border border-rose-400/20 bg-rose-400/10 px-4 py-3 text-sm text-rose-100">
                        {profileError}
                      </div>
                    )}

                    {profileNotice && (
                      <div className="rounded-2xl border border-emerald-400/20 bg-emerald-400/10 px-4 py-3 text-sm text-emerald-100">
                        {profileNotice}
                      </div>
                    )}
                  </div>
                </div>
              )}

              {activeTab === 'governance' && isSuperAdmin && (
                <div className="space-y-6">
                  <div className="rounded-[1.5rem] border border-white/10 bg-white/5 p-5">
                    <div className="flex flex-wrap items-center justify-between gap-4">
                      <div>
                        <p className="text-xs uppercase tracking-[0.3em] text-slate-500">Governance Control Center</p>
                        <h3 className="mt-3 text-xl font-semibold">Pending administrative profiles</h3>
                        <p className="mt-2 text-sm text-slate-400">
                          Approval is executed directly by super-admin action, and activation credentials are immediately dispatched to the target user.
                        </p>
                      </div>

                      <button
                        type="button"
                        onClick={() => void fetchPendingAdmins()}
                        className="rounded-full border border-white/10 px-4 py-2 text-sm text-slate-300 transition hover:border-emerald-400/40 hover:text-white"
                      >
                        Refresh Queue
                      </button>
                    </div>
                  </div>

                  <div className="overflow-hidden rounded-[1.5rem] border border-white/10 bg-white/5">
                    <div className="grid grid-cols-[1.4fr_0.9fr_0.8fr] gap-4 border-b border-white/10 px-5 py-4 text-xs uppercase tracking-[0.3em] text-slate-500">
                      <span>Pending Admin</span>
                      <span>Requested</span>
                      <span>Action</span>
                    </div>

                    {pendingBusy ? (
                      <div className="px-5 py-6 text-sm text-slate-400">Loading pending approvals...</div>
                    ) : pendingAdmins.length === 0 ? (
                      <div className="px-5 py-6">
                        <div className="rounded-2xl border border-emerald-400/15 bg-emerald-400/10 px-4 py-4 text-sm text-emerald-100">
                          <p className="font-medium">No pending administrative applications found in the ledger.</p>
                          <p className="mt-1 text-emerald-100/80">All registration requests have been reviewed and actioned.</p>
                        </div>
                      </div>
                    ) : (
                      pendingAdmins.map((pendingAdmin) => (
                        <div key={pendingAdmin.id} className="grid grid-cols-[1.4fr_0.9fr_0.8fr] gap-4 border-b border-white/5 px-5 py-4 text-sm text-slate-200 last:border-b-0">
                          <div>
                            <p className="font-medium text-white">{pendingAdmin.email}</p>
                            <p className="mt-1 text-xs uppercase tracking-[0.2em] text-slate-500">{pendingAdmin.role}</p>
                          </div>
                          <div className="text-slate-400">{new Date(pendingAdmin.created_at).toLocaleString()}</div>
                          <div>
                            <button
                              type="button"
                              onClick={(e) => void handleApprovePendingUser(e, pendingAdmin.email)}
                              disabled={governanceBusy}
                              className="rounded-full border border-emerald-400/20 bg-emerald-400/10 px-4 py-2 text-xs font-semibold text-emerald-100 transition hover:bg-emerald-400/15 disabled:cursor-not-allowed disabled:opacity-50"
                            >
                              Approve Profile Access
                            </button>
                          </div>
                        </div>
                      ))
                    )}
                  </div>

                  {governanceError && (
                    <div className="rounded-2xl border border-rose-400/20 bg-rose-400/10 px-4 py-3 text-sm text-rose-100">
                      {governanceError}
                    </div>
                  )}

                  {governanceNotice && (
                    <div className="rounded-2xl border border-emerald-400/20 bg-emerald-400/10 px-4 py-3 text-sm text-emerald-100">
                      {governanceNotice}
                    </div>
                  )}
                </div>
              )}
            </div>
          </div>
        </div>,
        document.body
      )
    : null;

  return (
    <>
      <button
        type="button"
        onClick={() => setIsOpen(true)}
        className={
          inline
            ? "inline-flex items-center gap-3 rounded-full border border-white/15 bg-white/10 px-3 py-2 text-sm font-semibold text-white shadow-[0_8px_30px_rgba(0,0,0,0.35)] backdrop-blur-2xl transition hover:border-cyan-400/40 hover:bg-cyan-400/10"
            : "absolute right-4 top-4 z-20 inline-flex items-center gap-3 rounded-full border border-white/15 bg-white/10 px-4 py-3 text-sm font-semibold text-white shadow-[0_18px_80px_rgba(0,0,0,0.45)] backdrop-blur-2xl transition hover:border-cyan-400/40 hover:bg-cyan-400/10 sm:right-6 sm:top-6"
        }
      >
        <span className="h-2.5 w-2.5 rounded-full bg-gradient-to-r from-cyan-400 to-emerald-400" />
        Profile Control
      </button>

      {modal}
    </>
  );
}
