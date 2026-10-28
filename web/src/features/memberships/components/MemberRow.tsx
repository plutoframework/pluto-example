import { Avatar, Button, Label, ListBox, Select, Table } from "@heroui/react";
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
    <Table.Row id={member.actor_id} data-testid={`member-${member.actor_id}`}>
      <Table.Cell>
        <div className="person">
          <Avatar size="sm" aria-hidden="true">
            <Avatar.Fallback>{initials(member.display_name)}</Avatar.Fallback>
          </Avatar>
          <span className="member-name">
            {member.display_name}
            {self && <small>You</small>}
          </span>
        </div>
      </Table.Cell>
      <Table.Cell className="muted mono">
        {member.actor_id.length > 18
          ? `${member.actor_id.slice(0, 18)}…`
          : member.actor_id}
      </Table.Cell>
      <Table.Cell>
        <Select
          variant="secondary"
          aria-label={`Role for ${member.display_name}`}
          value={role}
          isDisabled={self || pending}
          onChange={(value) => {
            if (typeof value === "string") setRole(value);
          }}
          className="role-select"
        >
          <Select.Trigger aria-label={`Role for ${member.display_name}`}>
            <Select.Value />
            <Select.Indicator />
          </Select.Trigger>
          <Select.Popover>
            <ListBox aria-label="Available roles">
              {roles.map((option) => (
                <ListBox.Item
                  id={option.id}
                  key={option.id}
                  textValue={option.name}
                >
                  <Label>{option.name}</Label>
                  <ListBox.ItemIndicator />
                </ListBox.Item>
              ))}
            </ListBox>
          </Select.Popover>
        </Select>
      </Table.Cell>
      <Table.Cell className="numeric">
        {self ? (
          <span className="muted">Managed by another admin</span>
        ) : (
          <Button
            variant="outline"
            size="sm"
            isDisabled={pending || role === member.role}
            onClick={() => {
              void onChangeRole(role);
            }}
          >
            Update role
          </Button>
        )}
      </Table.Cell>
    </Table.Row>
  );
}
