/** Generated from Pluto OpenAPI. Run npm run api:generate; do not edit. */
import type { accessOperations } from "./operations-access";
import type { auditOperations } from "./operations-audit";
import type { featuresOperations } from "./operations-features";
import type { membershipsOperations } from "./operations-memberships";
import type { organizationOperations } from "./operations-organization";
import type { purchasesOperations } from "./operations-purchases";
import type { credentialsOperations } from "./operations-credentials";
import type { oidcOperations } from "./operations-oidc";
import type { sessionsOperations } from "./operations-sessions";
export interface operations extends accessOperations, auditOperations, featuresOperations, membershipsOperations, organizationOperations, purchasesOperations, credentialsOperations, oidcOperations, sessionsOperations {}
