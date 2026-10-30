import type { ReactNode } from "react";
import { Card, Chip, Table } from "@heroui/react";
import { useRef } from "react";
import { useSearch, useNavigate } from "@tanstack/react-router";
import { useWorkspace } from "@/features/workspace/public";
import { errorMessage } from "@/api/errors";
import { PageHeader, Loading, Notice, EmptyState } from "@/shared/ui/feedback";
import { Pagination } from "@/shared/ui/Pagination";
import { nextPage, previousPage } from "@/shared/lib/search";
import { useMemberships } from "../queries";
import { useChangeRole } from "../mutations";
import { MemberRow } from "../components/MemberRow";

export function MembersPage() {
  const { scope, can, access } = useWorkspace();
  const search = useSearch({ from: "/members" });
  const navigate = useNavigate({ from: "/members" });
  const query = useMemberships(scope, search.cursor, can("memberships:manage"));
  const mutation = useChangeRole(scope);
  const intents = useRef(new Map<string, string>());

  if (!can("memberships:manage")) {
    return (
      <EmptyState
        title="A little outside your orbit"
        description="Organization membership management requires an administrator. Your other workspace features remain available."
      />
    );
  }

  let content: ReactNode;

  if (query.isPending) {
    content = <Loading label="Loading organization members…" />;
  } else if (query.error) {
    content = <Notice tone="error">{errorMessage(query.error)}</Notice>;
  } else {
    content = (
      <Table className="members-table">
        <Table.ScrollContainer
          className="table-scroll"
          tabIndex={0}
          role="region"
          aria-label="Organization members; scroll horizontally for all columns"
        >
          <Table.Content aria-label="Organization members">
            <Table.Header>
              <Table.Column isRowHeader>MEMBER</Table.Column>
              <Table.Column>ACCOUNT</Table.Column>
              <Table.Column>ROLE</Table.Column>
              <Table.Column className="numeric">ACTION</Table.Column>
            </Table.Header>
            <Table.Body>
              {query.data?.data.map((member) => (
                <MemberRow
                  key={`${member.actor_id}:${member.role}`}
                  member={member}
                  roles={access?.roles ?? []}
                  self={member.actor_id === scope.actor}
                  pending={mutation.isPending}
                  onChangeRole={async (role) => {
                    const signature = `${member.actor_id}:${role}`;

                    if (!intents.current.has(signature)) {
                      intents.current.set(signature, crypto.randomUUID());
                    }

                    try {
                      await mutation.mutateAsync({
                        actor: member.actor_id,
                        role,
                        key: intents.current.get(signature)!,
                      });
                    } catch {

                      /* Display server policy failure above. */
                    }
                  }}
                />
              ))}
            </Table.Body>
          </Table.Content>
        </Table.ScrollContainer>
      </Table>
    );
  }

  return (
    <>
      <PageHeader
        eyebrow="WORKSPACE / PEOPLE"
        title="Good work is a team effort."
        description="Clear roles. Shared responsibility. The right access for each person."
      />
      {mutation.error && (
        <Notice tone="error">{errorMessage(mutation.error)}</Notice>
      )}
      <Card className="panel">
        <div className="panel-heading">
          <div>
            <h2>Organization members</h2>
            <p>Role choices and effective capabilities come from Pluto.</p>
          </div>
          <Chip variant="soft" size="sm" className="tag">
            ADMINISTRATOR ACCESS
          </Chip>
        </div>
        {content}
        <Pagination
          count={query.data?.data.length ?? 0}
          previous={!!search.trail?.length}
          more={query.data?.page.has_more ?? false}
          onPrevious={() => {
            void navigate({ search: (old) => previousPage(old) });
          }}
          onNext={() => {
            if (query.data) {
              void navigate({
                search: (old) => nextPage(old, query.data!.page.next_cursor),
              });
            }
          }}
        />
      </Card>
      <p className="page-footnote">
        Changing a role is a server-authorized, audited action. You cannot
        change your own role.
      </p>
    </>
  );
}
