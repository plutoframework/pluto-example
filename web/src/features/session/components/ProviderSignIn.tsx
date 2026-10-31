import { Button } from "@heroui/react";

type Props = {
  available: boolean;
  pending: boolean;
  failed: boolean;
  redirecting: boolean;
  requestFailed: boolean;
  onContinue: () => void;
};

export function ProviderSignIn({
  available,
  pending,
  failed,
  redirecting,
  requestFailed,
  onContinue,
}: Props) {
  if (pending) {
    return <p role="status">Checking sign-in options…</p>;
  }

  if (failed) {
    return <p role="status">Sign-in options are temporarily unavailable.</p>;
  }

  if (!available) {
    return null;
  }

  return (
    <div>
      {requestFailed && <p role="alert">Unable to start sign-in. Try again.</p>}
      <Button type="button" isDisabled={redirecting} onClick={onContinue}>
        {redirecting ? "Redirecting…" : "Continue with your identity provider"}
      </Button>
    </div>
  );
}
