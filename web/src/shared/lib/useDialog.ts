import { useLayoutEffect, useRef } from "react";

// Reviewed DOM integration: native focus trapping and explicit trigger restoration.
// Data fetching never belongs in an effect; this hook only manages the dialog.
export function useDialog() {
  const reference = useRef<HTMLDialogElement>(null);
  useLayoutEffect(() => {
    const trigger = document.activeElement;
    const dialog = reference.current;
    dialog?.showModal();
    return () => {
      dialog?.close();
      if (trigger instanceof HTMLElement && trigger.isConnected)
        trigger.focus();
    };
  }, []);
  return reference;
}
