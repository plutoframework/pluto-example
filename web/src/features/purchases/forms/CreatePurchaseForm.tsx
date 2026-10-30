import { Button, Description, Input, Label, TextField } from "@heroui/react";
import { useState } from "react";
import type { CreatePurchase } from "../types";
import { Notice } from "@/shared/ui/feedback";

export function parseAmount(value: string): number | null {
  if (!/^\d{1,7}(?:\.\d{1,2})?$/.test(value)) {
    return null;
  }

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

        if (pending) {
          return;
        }

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
        await onSubmit({
          title: title.trim(),
          amount_cents: cents,
        });
      }}
    >
      {error && <Notice tone="error">{error}</Notice>}
      <TextField
        name="title"
        value={title}
        onChange={setTitle}
        isRequired
        isDisabled={pending}
        maxLength={160}
      >
        <Label>What does your team need?</Label>
        <Input
          variant="secondary"
          placeholder="e.g. Field research equipment"
          autoFocus
        />
      </TextField>
      <TextField
        name="amount"
        value={amount}
        onChange={setAmount}
        isRequired
        isDisabled={pending}
        inputMode="decimal"
      >
        <Label>Estimated cost (USD)</Label>
        <Input variant="secondary" placeholder="0.00" />
        <Description>
          A draft first. Submit it when you are ready for review.
        </Description>
      </TextField>
      <div className="button-row">
        <Button variant="primary" isDisabled={pending} type="submit">
          {pending ? "Saving…" : "Create draft"}
        </Button>
        <Button variant="outline" type="button" onClick={onCancel}>
          Cancel
        </Button>
      </div>
    </form>
  );
}
