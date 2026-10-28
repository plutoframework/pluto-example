/** Generated from Pluto OpenAPI. Run npm run api:generate; do not edit. */
import type { components } from "./components";
export interface auditOperations {
    "audit.list": {
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
                    "application/json": components["schemas"]["AuditResponse"];
                };
            };
            default: components["responses"]["Error"];
        };
    };
}
