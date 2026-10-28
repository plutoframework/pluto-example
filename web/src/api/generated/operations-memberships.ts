/** Generated from Pluto OpenAPI. Run npm run api:generate; do not edit. */
import type { components } from "./components";
export interface membershipsOperations {
    "memberships.list": {
        parameters: {
            query?: {
                limit?: number;
                cursor?: string;
            };
            header: {
                /** @description Active organization, verified against token membership. */
                "X-Organization-ID": string;
            };
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Success */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["MembershipsResponse"];
                };
            };
            default: components["responses"]["Error"];
        };
    };
    "memberships.change_role": {
        parameters: {
            query?: never;
            header: {
                /** @description Active organization, verified against token membership. */
                "X-Organization-ID": string;
                /** @description Reuse for retries of this action. 409 means conflict or pending reconciliation. Completed retention 24h. */
                "Idempotency-Key": string;
            };
            path: {
                actorID: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["ChangeRoleRequest"];
            };
        };
        responses: {
            /** @description Success */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ChangedResponse"];
                };
            };
            default: components["responses"]["Error"];
        };
    };
}
