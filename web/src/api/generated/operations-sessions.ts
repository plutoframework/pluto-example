/** Generated from Pluto OpenAPI. Run npm run api:generate; do not edit. */
import type { components } from "./components";
export interface sessionsOperations {
    "sessions.get": {
        parameters: {
            query?: never;
            header?: never;
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
                    "application/json": components["schemas"]["SessionResponse"];
                };
            };
            default: components["responses"]["Error"];
        };
    };
    "sessions.delete": {
        parameters: {
            query?: never;
            header?: never;
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
                    "application/json": components["schemas"]["ChangedResponse"];
                };
            };
            default: components["responses"]["Error"];
        };
    };
}
