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
    <div
      className="table-scroll"
      tabIndex={0}
      role="region"
      aria-label="Purchase requests; scroll horizontally for all columns"
    >
      <table>
        <thead>
          <tr>
            <th>REQUEST</th>
            <th>REQUESTED BY</th>
            <th>STATUS</th>
            <th>SUBMITTED</th>
            <th className="numeric">AMOUNT</th>
            <th>
              <span className="sr-only">View request</span>
            </th>
          </tr>
        </thead>
        <tbody>
          {purchases.map((purchase) => (
            <tr key={purchase.id}>
              <td>
                <Link
                  to="/purchases/$purchaseId"
                  params={{ purchaseId: purchase.id }}
                  className="record-title"
                >
                  {purchase.title}
                </Link>
                <span className="record-id">{purchase.id.slice(0, 12)}</span>
              </td>
              <td>
                <div className="person">
                  <span className="avatar">
                    {initials(purchase.creator_id)}
                  </span>
                  <span>
                    {purchase.creator_id === actor
                      ? "You"
                      : "Organization member"}
                    <small>{purchase.creator_id.slice(0, 12)}</small>
                  </span>
                </div>
              </td>
              <td>
                <PurchaseStatus status={purchase.status} />
              </td>
              <td className="muted">{date(purchase.created_at)}</td>
              <td className="numeric amount">{money(purchase.amount_cents)}</td>
              <td>
                <Link
                  to="/purchases/$purchaseId"
                  params={{ purchaseId: purchase.id }}
                  className="row-link"
                  aria-label={`View ${purchase.title}`}
                >
                  ↗
                </Link>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
