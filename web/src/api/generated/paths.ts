/** Generated from Pluto OpenAPI. Run npm run api:generate; do not edit. */
import type { operations } from "./operations";
export interface paths {
    "/api/access": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Read active organization capabilities
         * @description Describe this request's verified role and effective capabilities for UI visibility. Business operations independently enforce authorization.
         */
        get: operations["access.get"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/audit": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Read audit history
         * @description Members see their own actions; approvers and administrators see organization history.
         */
        get: operations["audit.list"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/features": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Read visible feature flags
         * @description Return allowlisted effective flags for the authenticated actor and verified organization; flags do not grant permissions.
         */
        get: operations["features.get"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/memberships": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * List organization members
         * @description Administrators list memberships in the verified organization.
         */
        get: operations["memberships.list"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/memberships/{actorID}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        /**
         * Change a member role
         * @description Administrators change another existing member's role. Self changes are forbidden.
         */
        patch: operations["memberships.change_role"];
        trace?: never;
    };
    "/api/organization": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Read organization profile
         * @description Read cached display metadata after live session, membership and permission checks.
         */
        get: operations["organization.get"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/purchases": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * List visible purchase requests
         * @description Members see their own requests; approvers and administrators see all requests in their organization.
         */
        get: operations["purchases.list"];
        put?: never;
        /**
         * Create purchase request
         * @description Create a draft request in the verified active organization.
         */
        post: operations["purchases.create"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/purchases/export": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Export purchase requests
         * @description Return a JSON export for an authorized approver when purchase_export is enabled.
         */
        get: operations["purchases.export"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/purchases/{purchaseID}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Read purchase request
         * @description Read an owned request or an organization request visible to an approver.
         */
        get: operations["purchases.get"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/purchases/{purchaseID}/decision": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Decide purchase request
         * @description An approver approves or rejects a submitted request created by someone else.
         */
        post: operations["purchases.decide"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/purchases/{purchaseID}/review": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Review purchase
         * @description Read-only advisory review. Repeated requests start new runs.
         */
        post: operations["purchases.review"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/purchases/{purchaseID}/submit": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Submit purchase request
         * @description The creator submits a draft for approval. Other transitions return conflict.
         */
        post: operations["purchases.submit"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/auth/enroll": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Enroll account
         * @description POST only, exact trusted Origin required; opaque code response. No automatic business privileges. JSON is bounded to 8 KiB.
         */
        post: operations["credentials.enroll"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/auth/logout": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Sign out
         * @description POST only, exact trusted Origin required; opaque code response. No automatic business privileges. JSON is bounded to 8 KiB.
         */
        post: operations["credentials.logout"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/auth/oidc/begin": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Begin OIDC sign-in
         * @description Trusted-Origin POST; redirects to configured provider with S256 PKCE.
         */
        post: operations["oidc.begin"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/auth/oidc/callback": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Complete OIDC sign-in
         * @description One-use browser/state/nonce callback; no automatic business grants.
         */
        get: operations["oidc.callback"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/auth/providers": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Configured sign-in providers
         * @description Public provider IDs only; no secrets or endpoint configuration.
         */
        get: operations["oidc.providers"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/auth/recovery": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Request recovery
         * @description POST only, exact trusted Origin required; opaque code response. No automatic business privileges. JSON is bounded to 8 KiB.
         */
        post: operations["credentials.recovery"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/auth/reset": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Reset password
         * @description POST only, exact trusted Origin required; opaque code response. No automatic business privileges. JSON is bounded to 8 KiB.
         */
        post: operations["credentials.reset"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/auth/session": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Read current session
         * @description Read verified session metadata without choosing an organization.
         */
        get: operations["sessions.get"];
        put?: never;
        post?: never;
        /**
         * Sign out
         * @description Revoke the current session before clearing its cookie. Does not require an active organization membership.
         */
        delete: operations["sessions.delete"];
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/auth/signin": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Sign in
         * @description POST only, exact trusted Origin required; opaque code response. No automatic business privileges. JSON is bounded to 8 KiB.
         */
        post: operations["credentials.signin"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/auth/verification": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Resend verification
         * @description POST only, exact trusted Origin required; opaque code response. No automatic business privileges. JSON is bounded to 8 KiB.
         */
        post: operations["credentials.verification"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/auth/verify": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Verify email
         * @description POST only, exact trusted Origin required; opaque code response. No automatic business privileges. JSON is bounded to 8 KiB.
         */
        post: operations["credentials.verify"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
}
