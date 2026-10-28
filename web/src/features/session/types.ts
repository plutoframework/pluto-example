import type { components } from "@/api/generated/components";
export type Session = components["schemas"]["Session"];
export type AuthMode =
  "signin" | "enroll" | "verify" | "resend" | "recover" | "reset";
export type AuthInput = { email: string; password: string; token: string };
