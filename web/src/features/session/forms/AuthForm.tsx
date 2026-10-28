import { useState } from "react";
import type { AuthInput, AuthMode } from "../types";

export const modeLabels: Record<AuthMode, string> = {
  signin: "Sign in",
  enroll: "Create an account",
  verify: "Verify your email",
  resend: "Send verification again",
  recover: "Recover access",
  reset: "Set a new password",
};
export function AuthForm({
  mode,
  pending,
  onSubmit,
}: {
  mode: AuthMode;
  pending: boolean;
  onSubmit: (input: AuthInput) => Promise<void>;
}) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [token, setToken] = useState("");
  const hasEmail = ["signin", "enroll", "resend", "recover"].includes(mode);
  const hasPassword = ["signin", "enroll", "reset"].includes(mode);
  return (
    <form
      className="stack"
      onSubmit={async (event) => {
        event.preventDefault();
        if (pending) return;
        await onSubmit({ email, password, token });
        setPassword("");
        setToken("");
      }}
    >
      {hasEmail && (
        <label>
          Email address
          <input
            type="email"
            autoComplete="username"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="you@example.test"
            required
            maxLength={254}
          />
        </label>
      )}
      {hasPassword && (
        <label>
          Password
          <input
            type="password"
            autoComplete={
              mode === "signin" ? "current-password" : "new-password"
            }
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            minLength={mode === "signin" ? 1 : 12}
            maxLength={1024}
          />
        </label>
      )}
      {!hasEmail && (
        <label>
          Confirmation token
          <input
            type="password"
            autoComplete="one-time-code"
            value={token}
            onChange={(e) => setToken(e.target.value)}
            required
            maxLength={128}
          />
          <span className="field-help">
            Paste the token from your local Mailpit message. It is sent only on
            submission.
          </span>
        </label>
      )}
      <button className="button primary" disabled={pending} type="submit">
        {pending ? "Working…" : modeLabels[mode]}
        <span aria-hidden="true">↗</span>
      </button>
    </form>
  );
}
