/** Generated from Pluto OpenAPI. Run npm run api:generate; do not edit. */
export interface components {
    schemas: {
        Access: {
            actor_id: string;
            organization_id: string;
            permissions: string[];
            role: string;
            role_name: string;
            roles: components["schemas"]["RoleOption"][];
        };
        AccessResponse: {
            data: components["schemas"]["Access"];
        };
        AuditEvent: {
            action: string;
            actor_id: string;
            id: string;
            /** Format: date-time */
            occurred_at: string;
            organization_id: string;
            outcome: string;
            resource_id: string;
        };
        AuditResponse: {
            data: components["schemas"]["AuditEvent"][];
            page: components["schemas"]["Metadata"];
        };
        AuthResult: {
            code: string;
        };
        ChallengeInput: {
            token: string;
        };
        ChangeRoleRequest: {
            role: string;
        };
        Changed: {
            updated: boolean;
        };
        ChangedResponse: {
            data: components["schemas"]["Changed"];
        };
        CreatePurchaseRequest: {
            /** Format: int64 */
            amount_cents: number;
            title: string;
        };
        DecisionRequest: {
            decision: string;
        };
        EmailInput: {
            email: string;
        };
        Error: {
            code: string;
            message: string;
            request_id: string;
        };
        ErrorResponse: {
            error: components["schemas"]["Error"];
        };
        FeaturesResponse: {
            data: {
                [key: string]: boolean;
            };
        };
        Membership: {
            actor_id: string;
            display_name: string;
            role: string;
        };
        MembershipsResponse: {
            data: components["schemas"]["Membership"][];
            page: components["schemas"]["Metadata"];
        };
        Metadata: {
            has_more: boolean;
            next_cursor: string;
        };
        Organization: {
            id: string;
            name: string;
        };
        OrganizationResponse: {
            data: components["schemas"]["Organization"];
        };
        PasswordInput: {
            email: string;
            password: string;
        };
        Purchase: {
            /** Format: int64 */
            amount_cents: number;
            /** Format: date-time */
            created_at: string;
            creator_id: string;
            id: string;
            organization_id: string;
            status: string;
            title: string;
            /** Format: date-time */
            updated_at: string;
            /** Format: int64 */
            version: number;
        };
        PurchaseExportResponse: {
            data: components["schemas"]["Purchase"][];
        };
        PurchaseResponse: {
            data: components["schemas"]["Purchase"];
        };
        PurchaseReview: {
            demo: boolean;
            next_steps: string[];
            summary: string;
        };
        PurchaseReviewResponse: {
            data: components["schemas"]["PurchaseReview"];
        };
        PurchasesResponse: {
            data: components["schemas"]["Purchase"][];
            page: components["schemas"]["Metadata"];
        };
        ResetInput: {
            password: string;
            token: string;
        };
        RoleOption: {
            description: string;
            id: string;
            name: string;
        };
        Session: {
            actor_id: string;
            /** Format: date-time */
            expires_at: string;
            id: string;
        };
        SessionResponse: {
            data: components["schemas"]["Session"];
        };
    };
    responses: {
        /** @description Error envelope; HTTP status identifies invalid input, auth, policy, conflict or internal failure. */
        Error: {
            headers: {
                [name: string]: unknown;
            };
            content: {
                "application/json": components["schemas"]["ErrorResponse"];
            };
        };
    };
    parameters: never;
    requestBodies: never;
    headers: never;
    pathItems: never;
}
