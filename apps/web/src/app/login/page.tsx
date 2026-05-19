'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useSession } from '../components/session-provider';

type LoginResponse = {
  message?: string;
  admin_id: number;
  email: string;
  role: string;
  session_token: string;
};

type RegisterResponse = {
  message?: string;
};

type FormMode = 'login' | 'request-access';

export default function LoginPage() {
  const router = useRouter();
  const { setAdminSession } = useSession();
  const [mode, setMode] = useState<FormMode>('login');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [notice, setNotice] = useState('');
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const submitLabel = mode === 'login' ? 'Sign In' : 'Request Access';

  const baseApiUrl = (process.env.NEXT_PUBLIC_API_URL || '').trim();
  // Standardize formatting to avoid double-slashes or missing routes
  const sanitizedApiUrl = baseApiUrl.endsWith('/api/v1') ? baseApiUrl : `${baseApiUrl}/api/v1`;
  // Construct the absolute address pointing directly to the Go backend
  const directLoginUrl = `${sanitizedApiUrl}/auth/login`;
  const registerEndpoint = `${sanitizedApiUrl}/auth/register`;

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError('');
    setNotice('');
    setIsSubmitting(true);

    try {
      if (mode === 'login') {
        const response = await fetch(directLoginUrl, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ email, password }),
        });

        const payload = (await response.json().catch(() => ({}))) as LoginResponse & { error?: string };

        if (!response.ok) {
          setError(String(payload?.error ?? 'Unable to authenticate admin'));
          return;
        }

        setAdminSession({
          adminId: payload.admin_id,
          email: payload.email,
          role: payload.role,
          token: payload.session_token,
        });
        router.replace('/map');
        return;
      }

      const response = await fetch(registerEndpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email }),
      });

      const payload = (await response.json().catch(() => ({}))) as RegisterResponse & { error?: string };

      if (!response.ok) {
        setError(String(payload?.error ?? 'Unable to request administrative access'));
        return;
      }

      setNotice('Awaiting admin sign-off');
      setPassword('');
      setMode('login');
    } catch (submissionError) {
      setError(String(submissionError));
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <main className="relative min-h-screen overflow-hidden bg-[#050816] text-white">
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_top_left,rgba(34,211,238,0.14),transparent_30%),radial-gradient(circle_at_bottom_right,rgba(16,185,129,0.16),transparent_30%),linear-gradient(180deg,#050816_0%,#040711_100%)]" />
      <div className="absolute left-[-10%] top-[-8%] h-72 w-72 rounded-full bg-cyan-400/20 blur-3xl" />
      <div className="absolute bottom-[-10%] right-[-6%] h-80 w-80 rounded-full bg-emerald-400/12 blur-3xl" />

      <section className="relative mx-auto flex min-h-screen w-full max-w-6xl items-center px-6 py-10 lg:px-10">
        <div className="grid w-full gap-6 lg:grid-cols-[0.9fr_1.1fr] lg:gap-8">
          <div className="rounded-[2rem] border border-white/10 bg-white/5 p-8 shadow-[0_30px_120px_rgba(0,0,0,0.45)] backdrop-blur-2xl lg:p-10">
            <p className="text-xs uppercase tracking-[0.45em] text-cyan-300/80">Administrative Gateway</p>
            <h1 className="mt-5 text-3xl font-semibold leading-tight sm:text-4xl">
              Secure the governance console.
            </h1>
            <p className="mt-4 text-sm leading-6 text-slate-300 sm:text-base">
              Admin credentials persist into localStorage-backed session state and unlock the sealing action in the map sidebar.
            </p>

            <div className="mt-8 rounded-3xl border border-white/10 bg-slate-950/40 p-5 text-sm text-slate-300">
              <p className="font-medium text-white">Session behavior</p>
              <div className="mt-3 space-y-2">
                <p>Login response stores admin ID, role, and token.</p>
                <p>Public sessions stay read-only and never show the seal control.</p>
                <p>Request access swaps the form into email-only mode.</p>
              </div>
            </div>
          </div>

          <div className="rounded-[2rem] border border-white/10 bg-white/5 p-6 shadow-[0_30px_120px_rgba(0,0,0,0.45)] backdrop-blur-2xl sm:p-8">
            <div className="flex items-center justify-between gap-4">
              <div>
                <p className="text-xs uppercase tracking-[0.35em] text-slate-500">Portal</p>
                <h2 className="mt-2 text-2xl font-semibold">{mode === 'login' ? 'Administrator Sign In' : 'Request Administrative Access'}</h2>
              </div>

              <button
                type="button"
                onClick={() => {
                  setMode(mode === 'login' ? 'request-access' : 'login');
                  setError('');
                  setNotice('');
                }}
                className="text-right text-sm text-cyan-300 transition hover:text-cyan-200"
              >
                {mode === 'login' ? 'New Administrative User? Request Access' : 'Back to Sign In'}
              </button>
            </div>

            <form onSubmit={handleSubmit} className="mt-8 space-y-4">
              <div>
                <label className="mb-2 block text-sm font-medium text-slate-200" htmlFor="admin-email">
                  {mode === 'login' ? 'Email or username' : 'Email'}
                </label>
                <input
                  id="admin-email"
                  type={mode === 'login' ? 'text' : 'email'}
                  value={email}
                  onChange={(event) => setEmail(event.target.value)}
                  placeholder={mode === 'login' ? 'super or admin@urbanmemory.gov' : 'admin@urbanmemory.gov'}
                  autoComplete={mode === 'login' ? 'username' : 'email'}
                  className="w-full rounded-2xl border border-white/10 bg-slate-950/80 px-4 py-3 text-white outline-none transition placeholder:text-slate-500 focus:border-cyan-400/50 focus:ring-2 focus:ring-cyan-400/15"
                />
              </div>

              {mode === 'login' ? (
                <div>
                  <label className="mb-2 block text-sm font-medium text-slate-200" htmlFor="admin-password">
                    Password
                  </label>
                  <input
                    id="admin-password"
                    type="password"
                    value={password}
                    onChange={(event) => setPassword(event.target.value)}
                    placeholder="Enter your admin password"
                    className="w-full rounded-2xl border border-white/10 bg-slate-950/80 px-4 py-3 text-white outline-none transition placeholder:text-slate-500 focus:border-cyan-400/50 focus:ring-2 focus:ring-cyan-400/15"
                  />
                </div>
              ) : (
                <div className="rounded-2xl border border-cyan-400/15 bg-cyan-400/10 px-4 py-4 text-sm text-cyan-100">
                  <p className="font-medium">Email-only registration</p>
                  <p className="mt-1 text-cyan-100/80">Submitting this form creates a locked pending request for super-admin review.</p>
                </div>
              )}

              <button
                type="submit"
                disabled={isSubmitting}
                className="w-full rounded-full bg-gradient-to-r from-cyan-400 to-emerald-400 px-5 py-4 text-sm font-semibold text-slate-950 transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-50"
              >
                {isSubmitting ? 'Working...' : submitLabel}
              </button>
            </form>

            {error && (
              <div className="mt-5 rounded-2xl border border-rose-400/20 bg-rose-400/10 px-4 py-3 text-sm text-rose-100">
                {error}
              </div>
            )}

            {notice && (
              <div className="mt-5 rounded-2xl border border-emerald-400/20 bg-emerald-400/10 px-4 py-3 text-sm text-emerald-100">
                <p className="font-medium">{notice}</p>
              </div>
            )}
          </div>
        </div>
      </section>
    </main>
  );
}
