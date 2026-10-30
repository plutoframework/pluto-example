import type { components } from "@/api/generated/components";

export type Access = components["schemas"]["Access"];

export type Organization = components["schemas"]["Organization"];

export type Scope = {
  actor: string;
  session: string;
  organization: string;
};

export function scopeKey(scope: Scope): readonly string[] {
  return ["scope", scope.actor, scope.session, scope.organization];
}
