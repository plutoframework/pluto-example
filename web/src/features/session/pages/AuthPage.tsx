import { useState } from "react";
import { OrbitMark } from "@/shared/ui/icons";
import { Notice } from "@/shared/ui/feedback";
import { errorMessage } from "@/api/errors";
import { useAuthentication } from "../mutations";
import { AuthForm, modeLabels } from "../forms/AuthForm";
import type { AuthMode } from "../types";

export function AuthPage() {
  const [mode, setMode] = useState<AuthMode>("signin");
  const [message, setMessage] = useState("");
  const mutation = useAuthentication();
  return (
    <main className="auth-layout">
      <section className="auth-story">
        <a className="brand" href="/">
          <OrbitMark />
          <span>
            pluto<span className="brand-caption">RESEARCH OPERATIONS</span>
          </span>
        </a>
        <div className="auth-copy">
          <span className="eyebrow">LESS FRICTION. MORE POSSIBILITY.</span>
          <h1>
            Good ideas
            <br />
            deserve
            <br />
            <em>clear skies.</em>
          </h1>
          <p>
            A considered workspace for the people, purchases and decisions that
            move your organization forward.
          </p>
        </div>
        <div className="orbital-art" aria-hidden="true">
          <span />
          <span />
          <span />
          <i />
        </div>
        <div className="auth-footer">
          BUILT WITH PLUTO<span>01 / PURCHASING</span>
        </div>
      </section>
      <section className="auth-panel">
        <div className="auth-card">
          <span className="eyebrow">YOUR NEXT CHAPTER STARTS HERE</span>
          <h2>{modeLabels[mode]}</h2>
          <p className="muted">A secure session. A shared purpose.</p>
          {mutation.error && (
            <Notice tone="error">{errorMessage(mutation.error)}</Notice>
          )}
          {message && <Notice tone="success">{message}</Notice>}
          <AuthForm
            key={mode}
            mode={mode}
            pending={mutation.isPending}
            onSubmit={async (input) => {
              setMessage("");
              try {
                await mutation.mutateAsync({ mode, input });
                if (mode !== "signin")
                  setMessage(
                    mode === "verify"
                      ? "Confirmation complete. You can now sign in."
                      : mode === "reset"
                        ? "Password reset complete. Sign in with your new password."
                        : "If this request is eligible, a message is waiting in local Mailpit.",
                  );
              } catch {
                /* The mutation's error state is shown above without raw server text. */
              }
            }}
          />
          <nav className="auth-links" aria-label="Account actions">
            {(Object.keys(modeLabels) as AuthMode[])
              .filter((value) => value !== mode)
              .map((value) => (
                <button
                  key={value}
                  disabled={mutation.isPending}
                  onClick={() => {
                    setMode(value);
                    setMessage("");
                    mutation.reset();
                  }}
                >
                  {modeLabels[value]}
                </button>
              ))}
          </nav>
          <div className="local-note">
            <span className="signal-dot" />
            LOCAL DEVELOPMENT
            <p>
              Use an <code>example.test</code> address. Verification and
              recovery messages stay in{" "}
              <a href="http://127.0.0.1:8025" target="_blank" rel="noreferrer">
                Mailpit ↗
              </a>
              . New accounts start without organization access.
            </p>
          </div>
        </div>
      </section>
    </main>
  );
}
