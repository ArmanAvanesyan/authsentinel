/**
 * Shared TypeScript helpers and re-exports for BFFs and servers.
 * Use @authsentinel/sdk-core for browser client and types;
 * use @authsentinel/sdk-node for Express/Fastify middleware.
 */
export {
  createClient,
  type AuthClient,
  type AuthClientConfig,
  type Principal,
  type RefreshResult,
  type SessionInfo,
  type SessionStatus,
  type SessionUser,
} from "@authsentinel/sdk-core";

