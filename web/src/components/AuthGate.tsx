import { ReactNode, useEffect, useState } from "react";
import { ApiError, clearToken, getToken, runSetup, setToken } from "../api";

// AuthGate blocks the app shell until a Bearer token is in session
// storage and verified to work. If the backend reports 503 on a /ping
// call the user is offered a "Provision admin token" flow that runs
// /api/v1/setup.
type Phase = "checking" | "setup" | "login" | "ready";

export default function AuthGate({ children }: { children: ReactNode }) {
  const [phase, setPhase] = useState<Phase>("checking");
  const [error, setError] = useState<string | null>(null);
  const [tokenInput, setTokenInput] = useState("");

  useEffect(() => {
    void verify();
  }, []);

  async function verify() {
    const token = getToken();
    if (!token) {
      // No token yet — determine whether setup is needed.
      try {
        const res = await fetch("/api/v1/ping");
        if (res.status === 503) {
          setPhase("setup");
          return;
        }
        setPhase("login");
      } catch (e) {
        setError(String(e));
        setPhase("login");
      }
      return;
    }
    try {
      const res = await fetch("/api/v1/ping", {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (res.status === 503) {
        setPhase("setup");
        return;
      }
      if (!res.ok) {
        clearToken();
        setPhase("login");
        return;
      }
      setPhase("ready");
    } catch (e) {
      setError(String(e));
      setPhase("login");
    }
  }

  async function provision() {
    setError(null);
    try {
      const res = await runSetup();
      if (res.admin_token) {
        setToken(res.admin_token);
        setPhase("ready");
      }
    } catch (e) {
      setError(e instanceof ApiError ? e.message : String(e));
    }
  }

  function submitToken() {
    if (!tokenInput.trim()) return;
    setToken(tokenInput.trim());
    void verify();
  }

  if (phase === "ready") return <>{children}</>;

  return (
    <div className="flex min-h-screen items-center justify-center bg-mocha-base p-6">
      <div className="w-full max-w-md rounded-lg border border-mocha-surface1 bg-mocha-mantle p-6 shadow-lg">
        <h1 className="mb-4 text-2xl font-semibold text-mocha-lavender">DEVON</h1>

        {phase === "checking" && <p className="text-mocha-subtext0">Verifying admin token…</p>}

        {phase === "setup" && (
          <>
            <p className="mb-4 text-sm text-mocha-subtext1">
              This DEVON deployment has not been provisioned. Generate an admin token now — copy
              it somewhere safe, you won't see it again.
            </p>
            <button
              type="button"
              onClick={provision}
              className="rounded bg-mocha-blue px-4 py-2 font-medium text-mocha-crust hover:bg-mocha-sky"
            >
              Provision admin token
            </button>
          </>
        )}

        {phase === "login" && (
          <>
            <p className="mb-3 text-sm text-mocha-subtext1">
              Paste your DEVON admin token to continue.
            </p>
            <input
              type="password"
              value={tokenInput}
              onChange={(e) => setTokenInput(e.target.value)}
              placeholder="DEVON_ADMIN_TOKEN"
              className="mb-3 w-full rounded border border-mocha-surface1 bg-mocha-base p-2 text-mocha-text"
            />
            <button
              type="button"
              onClick={submitToken}
              className="rounded bg-mocha-blue px-4 py-2 font-medium text-mocha-crust hover:bg-mocha-sky"
            >
              Sign in
            </button>
          </>
        )}

        {error && <p className="mt-4 text-sm text-mocha-red">{error}</p>}
      </div>
    </div>
  );
}
