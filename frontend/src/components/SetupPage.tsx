import { useState } from "react";
import { runSetup, ApiError } from "../api/client";

interface SetupPageProps {
  onComplete: (key: string) => void;
}

type Phase = "welcome" | "display" | "error";

export default function SetupPage({ onComplete }: SetupPageProps) {
  const [phase, setPhase] = useState<Phase>("welcome");
  const [apiKey, setApiKey] = useState("");
  const [revealed, setRevealed] = useState(false);
  const [copied, setCopied] = useState(false);
  const [error, setError] = useState("");
  const [generating, setGenerating] = useState(false);

  async function handleGenerate() {
    setGenerating(true);
    setError("");
    try {
      const res = await runSetup();
      setApiKey(res.api_key);
      setPhase("display");
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        setError("An API key has already been configured. Refresh and sign in.");
      } else {
        setError("Failed to connect to DEVON server.");
      }
      setPhase("error");
    } finally {
      setGenerating(false);
    }
  }

  async function handleCopy() {
    try {
      await navigator.clipboard.writeText(apiKey);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      /* clipboard API unavailable — user can select manually */
    }
  }

  // Welcome — explain and generate
  if (phase === "welcome") {
    return (
      <div className="flex h-screen items-center justify-center bg-ctp-base px-4">
        <div className="w-full max-w-md rounded-lg border border-ctp-surface0 bg-ctp-mantle p-8 text-center">
          <h1 className="text-2xl font-bold text-ctp-blue mb-2">DEVON</h1>
          <p className="text-sm text-ctp-subtext0 mb-6">
            Welcome! This instance needs an API key before you can use it.
            Click below to generate a secure key. You will only see it once.
          </p>
          <button
            onClick={handleGenerate}
            disabled={generating}
            className="rounded-lg bg-ctp-blue px-6 py-2.5 text-sm font-medium text-ctp-crust hover:bg-ctp-blue/80 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {generating ? "Generating..." : "Generate API Key"}
          </button>
        </div>
      </div>
    );
  }

  // Display — show key once
  if (phase === "display") {
    const masked = "\u2022".repeat(apiKey.length);
    return (
      <div className="flex h-screen items-center justify-center bg-ctp-base px-4">
        <div className="w-full max-w-md rounded-lg border border-ctp-surface0 bg-ctp-mantle p-8">
          <h1 className="text-2xl font-bold text-ctp-blue mb-2 text-center">API Key Generated</h1>

          <div className="mt-4 rounded-md border border-ctp-surface1 bg-ctp-surface0 p-3 font-mono text-sm text-ctp-text break-all select-all">
            {revealed ? apiKey : masked}
          </div>

          <div className="mt-3 flex gap-2">
            <button
              onClick={() => setRevealed((r) => !r)}
              className="flex-1 rounded-lg border border-ctp-surface1 bg-ctp-surface0 px-3 py-2 text-sm text-ctp-text hover:bg-ctp-surface1 transition-colors"
            >
              {revealed ? "Hide" : "Reveal"}
            </button>
            <button
              onClick={handleCopy}
              className="flex-1 rounded-lg border border-ctp-surface1 bg-ctp-surface0 px-3 py-2 text-sm text-ctp-text hover:bg-ctp-surface1 transition-colors"
            >
              {copied ? "Copied!" : "Copy"}
            </button>
          </div>

          <div className="mt-4 rounded-lg border border-ctp-yellow/30 bg-ctp-yellow/10 px-3 py-2 text-sm text-ctp-yellow">
            Save this key now — it will not be shown again.
          </div>

          <button
            onClick={() => onComplete(apiKey)}
            className="mt-5 w-full rounded-lg bg-ctp-blue px-4 py-2.5 text-sm font-medium text-ctp-crust hover:bg-ctp-blue/80 transition-colors"
          >
            I've saved my key — continue
          </button>
        </div>
      </div>
    );
  }

  // Error
  return (
    <div className="flex h-screen items-center justify-center bg-ctp-base px-4">
      <div className="w-full max-w-md rounded-lg border border-ctp-red/30 bg-ctp-mantle p-8 text-center">
        <h2 className="text-lg font-bold text-ctp-text mb-2">Setup Error</h2>
        <p className="text-sm text-ctp-subtext0">{error}</p>
        <button
          onClick={() => window.location.reload()}
          className="mt-4 rounded-lg border border-ctp-surface1 bg-ctp-surface0 px-4 py-2 text-sm text-ctp-text hover:bg-ctp-surface1 transition-colors"
        >
          Retry
        </button>
      </div>
    </div>
  );
}
