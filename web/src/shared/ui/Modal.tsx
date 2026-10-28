import type { ReactNode } from "react";
import { useDialog } from "../lib/useDialog";
export function Modal({
  title,
  children,
  onClose,
}: {
  title: string;
  children: ReactNode;
  onClose: () => void;
}) {
  const reference = useDialog();
  return (
    <dialog
      aria-labelledby="modal-title"
      className="modal"
      ref={reference}
      onCancel={(event) => {
        event.preventDefault();
        onClose();
      }}
    >
      <div className="modal-heading">
        <span className="eyebrow">MAKE ROOM FOR WHAT'S NEXT</span>
        <h2 id="modal-title">{title}</h2>
      </div>
      {children}
      <button
        className="modal-close"
        aria-label="Close dialog"
        onClick={onClose}
      >
        ×
      </button>
    </dialog>
  );
}
