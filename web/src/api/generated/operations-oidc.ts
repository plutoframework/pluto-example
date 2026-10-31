/** Generated from Pluto OpenAPI. Run npm run api:generate; do not edit. */
import type { components } from "./components";
export interface oidcOperations {
    "oidc.begin": {
        parameters: {
            query?: never;
            header: {
                Origin: string;
            };
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description With Accept: application/json, returns the server-built authorization URL. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["OIDCAuthorizationResponse"];
                };
            };
            /** @description Browser redirect; no JSON success body. */
            303: {
                headers: {
                    Location: string;
                    [name: string]: unknown;
                };
                content?: never;
            };
            /** @description Opaque authentication error; no credential data. */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": {
                        error: {
                            code: string;
                        };
                    };
                };
            };
        };
    };
    "oidc.callback": {
        parameters: {
            query: {
                state: string;
                code: string;
            };
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Browser redirect; no JSON success body. */
            303: {
                headers: {
                    Location: string;
                    [name: string]: unknown;
                };
                content?: never;
            };
            /** @description Opaque authentication error; no credential data. */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": {
                        error: {
                            code: string;
                        };
                    };
                };
            };
        };
    };
    "oidc.providers": {
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
                    "application/json": components["schemas"]["OIDCProviderList"];
                };
            };
            /** @description Opaque authentication error; no credential data. */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": {
                        error: {
                            code: string;
                        };
                    };
                };
            };
        };
    };
}
