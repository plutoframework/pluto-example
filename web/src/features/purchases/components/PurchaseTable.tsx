import { Avatar, Table } from "@heroui/react";
import { Link } from "@tanstack/react-router";
import { date, money } from "@/shared/lib/format";
import { initials } from "@/shared/lib/format";
import type { Purchase } from "../types";
import { PurchaseStatus } from "./PurchaseStatus";

export function PurchaseTable({
  purchases,
  actor,
}: {
  purchases: Purchase[];
  actor: string;
}) {
  return (
    <Table className="purchase-table">
      <Table.ScrollContainer
        className="table-scroll"
        tabIndex={0}
        role="region"
        aria-label="Purchase requests; scroll horizontally for all columns"
      >
        <Table.Content aria-label="Purchase requests">
          <Table.Header>
            <Table.Column isRowHeader>REQUEST</Table.Column>
            <Table.Column>REQUESTED BY</Table.Column>
            <Table.Column>STATUS</Table.Column>
            <Table.Column>SUBMITTED</Table.Column>
            <Table.Column className="numeric">AMOUNT</Table.Column>
            <Table.Column>
              <span className="sr-only">View request</span>
            </Table.Column>
          </Table.Header>
          <Table.Body>
            {purchases.map((purchase) => (
              <Table.Row id={purchase.id} key={purchase.id}>
                <Table.Cell>
                  <Link
                    to="/purchases/$purchaseId"
                    params={{ purchaseId: purchase.id }}
                    className="record-title"
                  >
                    {purchase.title}
                  </Link>
                  <span className="record-id">{purchase.id.slice(0, 12)}</span>
                </Table.Cell>
                <Table.Cell>
                  <div className="person">
                    <Avatar size="sm" aria-hidden="true">
                      <Avatar.Fallback>
                        {initials(purchase.creator_id)}
                      </Avatar.Fallback>
                    </Avatar>
                    <span>
                      {purchase.creator_id === actor
                        ? "You"
                        : "Organization member"}
                      <small>{purchase.creator_id.slice(0, 12)}</small>
                    </span>
                  </div>
                </Table.Cell>
                <Table.Cell>
                  <PurchaseStatus status={purchase.status} />
                </Table.Cell>
                <Table.Cell className="muted">
                  {date(purchase.created_at)}
                </Table.Cell>
                <Table.Cell className="numeric amount">
                  {money(purchase.amount_cents)}
                </Table.Cell>
                <Table.Cell>
                  <Link
                    to="/purchases/$purchaseId"
                    params={{ purchaseId: purchase.id }}
                    className="row-link"
                    aria-label={`View ${purchase.title}`}
                  >
                    ↗
                  </Link>
                </Table.Cell>
              </Table.Row>
            ))}
          </Table.Body>
        </Table.Content>
      </Table.ScrollContainer>
    </Table>
  );
}
