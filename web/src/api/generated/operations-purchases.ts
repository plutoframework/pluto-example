/** Generated from Pluto OpenAPI. Run npm run api:generate; do not edit. */
import type { components } from "./components";
export interface purchasesOperations {
    "purchases.list": {
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
                    "application/json": components["schemas"]["PurchasesResponse"];
                };
            };
            default: components["responses"]["Error"];
        };
    };
    "purchases.create": {
        parameters: {
            query?: never;
            header: {
                /** @description Active organization, verified against token membership. */
                "X-Organization-ID": string;
                /** @description Reuse for retries of this action. 409 means conflict or pending reconciliation. Completed retention 24h. */
                "Idempotency-Key": string;
            };
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["CreatePurchaseRequest"];
            };
        };
        responses: {
            /** @description Success */
            201: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["PurchaseResponse"];
                };
            };
            default: components["responses"]["Error"];
        };
    };
    "purchases.export": {
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
                    "application/json": components["schemas"]["PurchaseExportResponse"];
                };
            };
            default: components["responses"]["Error"];
        };
    };
    "purchases.get": {
        parameters: {
            query?: never;
            header: {
                /** @description Active organization, verified against token membership. */
                "X-Organization-ID": string;
            };
            path: {
                purchaseID: string;
            };
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
                    "application/json": components["schemas"]["PurchaseResponse"];
                };
            };
            default: components["responses"]["Error"];
        };
    };
    "purchases.decide": {
        parameters: {
            query?: never;
            header: {
                /** @description Active organization, verified against token membership. */
                "X-Organization-ID": string;
                /** @description Reuse for retries of this action. 409 means conflict or pending reconciliation. Completed retention 24h. */
                "Idempotency-Key": string;
            };
            path: {
                purchaseID: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["DecisionRequest"];
            };
        };
        responses: {
            /** @description Success */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["PurchaseResponse"];
                };
            };
            default: components["responses"]["Error"];
        };
    };
    "purchases.review": {
        parameters: {
            query?: never;
            header: {
                /** @description Active organization, verified against token membership. */
                "X-Organization-ID": string;
            };
            path: {
                purchaseID: string;
            };
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
                    "application/json": components["schemas"]["PurchaseReviewResponse"];
                };
            };
            default: components["responses"]["Error"];
        };
    };
    "purchases.submit": {
        parameters: {
            query?: never;
            header: {
                /** @description Active organization, verified against token membership. */
                "X-Organization-ID": string;
                /** @description Reuse for retries of this action. 409 means conflict or pending reconciliation. Completed retention 24h. */
                "Idempotency-Key": string;
            };
            path: {
                purchaseID: string;
            };
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
                    "application/json": components["schemas"]["PurchaseResponse"];
                };
            };
            default: components["responses"]["Error"];
        };
    };
}
