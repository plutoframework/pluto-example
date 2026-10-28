/** Generated from Pluto OpenAPI. Run npm run api:generate; do not edit. */
import type { components } from "./components";
export interface accessOperations {
    "access.get": {
        parameters: {
            query?: never;
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
                    "application/json": components["schemas"]["AccessResponse"];
                };
            };
            default: components["responses"]["Error"];
        };
    };
}
