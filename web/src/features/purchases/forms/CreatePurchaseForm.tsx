import { useState } from "react";
import type { CreatePurchase } from "../types";
import { Notice } from "@/shared/ui/feedback";

export function parseAmount(value: string): number | null {
  if (!/^\d{1,7}(?:\.\d{1,2})?$/.test(value)) return null;
  const [whole = "0", fraction = ""] = value.split(".");
  const cents = Number(whole) * 100 + Number(fraction.padEnd(2, "0"));
  return Number.isSafeInteger(cents) && cents >= 1 && cents <= 100_000_000
    ? cents
    : null;
}
export function CreatePurchaseForm({
  pending,
  onSubmit,
  onCancel,
}: {
  pending: boolean;
  onSubmit: (input: CreatePurchase) => Promise<void>;
  onCancel: () => void;
}) {
  const [title, setTitle] = useState("");
  const [amount, setAmount] = useState("");
  const [error, setError] = useState("");
  return (
    <form
      className="stack"
      onSubmit={async (event) => {
        event.preventDefault();
        if (pending) return;
        const cents = parseAmount(amount);
        if (
          cents === null ||
          !title.trim() ||
          new TextEncoder().encode(title).length > 160
        ) {
          setError(
            "Enter a title up to 160 UTF-8 bytes and an amount from $0.01 to $1,000,000.00.",
          );
          return;
        }
        setError("");
        await onSubmit({ title: title.trim(), amount_cents: cents });
      }}
    >
      {error && <Notice tone="error">{error}</Notice>}
      <label>
        What does your team need?
        <input
          name="title"
          value={title}
          onChange={(event) => setTitle(event.target.value)}
          placeholder="e.g. Field research equipment"
          required
          maxLength={160}
        />
      </label>
      <label>
        Estimated cost (USD)
        <div className="money-input">
          <span>$</span>
          <input
            name="amount"
            inputMode="decimal"
            value={amount}
            onChange={(event) => setAmount(event.target.value)}
            placeholder="0.00"
            required
          />
        </div>
        <span className="field-help">
          A draft first. Submit it when you are ready for review.
        </span>
      </label>
      <div className="button-row">
        <button className="button primary" disabled={pending} type="submit">
          {pending ? "Saving…" : "Create draft"}
        </button>
        <button className="button ghost" type="button" onClick={onCancel}>
          Cancel
        </button>
      </div>
    </form>
  );
}
