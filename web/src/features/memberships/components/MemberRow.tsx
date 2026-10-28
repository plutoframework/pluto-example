import { useState } from "react";
import { initials } from "@/shared/lib/format";
import type { Membership, RoleOption } from "../types";
export function MemberRow({
  member,
  roles,
  self,
  pending,
  onChangeRole,
}: {
  member: Membership;
  roles: RoleOption[];
  self: boolean;
  pending: boolean;
  onChangeRole: (role: string) => Promise<void>;
}) {
  const [role, setRole] = useState(member.role);
  return (
    <tr data-testid={`member-${member.actor_id}`}>
      <td>
        <div className="person">
          <span className="avatar">{initials(member.display_name)}</span>
          <span className="member-name">
            {member.display_name}
            {self && <small>You</small>}
          </span>
        </div>
      </td>
      <td className="muted mono">{member.actor_id.slice(0, 18)}…</td>
      <td>
        <label>
          <span className="sr-only">Role for {member.display_name}</span>
          <select
            value={role}
            disabled={self || pending}
            onChange={(event) => setRole(event.target.value)}
          >
            {roles.map((option) => (
              <option key={option.id} value={option.id}>
                {option.name}
              </option>
            ))}
          </select>
        </label>
      </td>
      <td className="numeric">
        {self ? (
          <span className="muted">Managed by another admin</span>
        ) : (
          <button
            className="button ghost small"
            disabled={pending || role === member.role}
            onClick={() => {
              void onChangeRole(role);
            }}
          >
            Update role
          </button>
        )}
      </td>
    </tr>
  );
}
