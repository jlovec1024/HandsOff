/**
 * Application route constants
 * Centralized route management to avoid hard-coded paths
 */
export const ROUTES = {
  HOME: "/",
  LOGIN: "/login",
} as const;

export type RouteKey = keyof typeof ROUTES;
