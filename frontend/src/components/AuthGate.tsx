import { useState, useEffect, useCallback } from "react";
import { getConfig, getSetupCheck, setApiKey, clearApiKey, ApiError } from "../api/client";
import SetupPage from "./SetupPage";

interface AuthGateProps {
  children: React.ReactNode;
}

export default function AuthGate({ children }: AuthGateProps) {
  const [authed, setAuthed] = useState<boolean | null>(null);
  const [needsSetup, setNeedsSetup] = useState(false);
  const [serverError, setServerError] = useState("");
  const [keyInput, setKeyInput] = useState("");
  const [loginError, setLoginError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const checkAuth = useCallback(async () => {
    try {
      await getConfig();
      setAuthed(true);
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.status === 401) {
          clearApiKey();
          setAuthed(false);
        } else if (err.status === 503) {
          // Could be setup-required or a real server error
          try {
            const check = await getSetupCheck();
            if (check.needs_setup) {
              setNeedsSetup(true);
            } else {
              // Key exists but 503 for another reason — show login
              setAuthed(false);
            }
          } catch {
            setServerError("Cannot reach DEVON server.");
          }
        } else {
          setServerError(err.message);
        }
      } else {
        setServerError("Cannot reach DEVON server.");
      }
    }
  }, []);

  useEffect(() => {
    checkAuth();
  }, [checkAuth]);

  function handleSetupComplete(key: string) {
    setApiKey(key);
    setNeedsSetup(false);
    setAuthed(true);
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const key = keyInput.trim();
    if (!key) return;

    setSubmitting(true);
    setLoginError("");
    setApiKey(key);

    try {
      await getConfig();
      setAuthed(true);
    } catch (err) {
      clearApiKey();
      if (err instanceof ApiError && err.status === 401) {
        setLoginError("Invalid API key.");
      } else {
        setLoginError("Connection failed. Check that the server is running.");
      }
    } finally {
      setSubmitting(false);
    }
  }

  // First-run setup flow
  if (needsSetup) {
    return <SetupPage onComplete={handleSetupComplete} />;
  }

  // Checking auth on mount
  if (authed === null && !serverError) {
    return (
      <div className="flex h-screen items-center justify-center bg-ctp-base">
        <p className="text-ctp-subtext0">Connecting...</p>
      </div>
    );
  }

  // Server-side misconfiguration
  if (serverError) {
    return (
      <div className="flex h-screen items-center justify-center bg-ctp-base px-4">
        <div className="w-full max-w-sm rounded-lg border border-ctp-red/30 bg-ctp-mantle p-6 text-center">
          <h2 className="text-lg font-bold text-ctp-text mb-2">Server Error</h2>
          <p className="text-sm text-ctp-subtext0">{serverError}</p>
        </div>
      </div>
    );
  }

  // Needs API key
  if (authed === false) {
    return (
      <div className="flex h-screen items-center justify-center bg-ctp-base px-4">
        <form
          onSubmit={handleSubmit}
          className="w-full max-w-sm rounded-lg border border-ctp-surface0 bg-ctp-mantle p-6"
        >
          <h1 className="text-xl font-bold text-ctp-blue mb-1">DEVON</h1>
          <p className="text-sm text-ctp-subtext0 mb-5">Enter your API key to continue.</p>

          {loginError && (
            <div className="mb-4 rounded-lg border border-ctp-red/30 bg-ctp-red/10 px-3 py-2 text-sm text-ctp-red">
              {loginError}
            </div>
          )}

          <label className="block mb-4">
            <span className="text-sm font-medium text-ctp-subtext1">API Key</span>
            <input
              type="password"
              value={keyInput}
              onChange={(e) => setKeyInput(e.target.value)}
              placeholder="Bearer token"
              autoFocus
              className="mt-1 block w-full rounded-md border border-ctp-surface1 bg-ctp-surface0 px-3 py-2 text-sm text-ctp-text placeholder-ctp-overlay0 focus:border-ctp-blue focus:outline-none focus:ring-1 focus:ring-ctp-blue"
            />
          </label>

          <button
            type="submit"
            disabled={submitting || !keyInput.trim()}
            className="w-full rounded-lg bg-ctp-blue px-4 py-2 text-sm font-medium text-ctp-crust hover:bg-ctp-blue/80 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {submitting ? "Verifying..." : "Sign In"}
          </button>
        </form>
      </div>
    );
  }

  return <>{children}</>;
}
