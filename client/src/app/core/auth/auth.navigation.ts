/** Landing route used by administrators after authentication. */
export const ADMIN_HOME = '/';

/** Landing route used by regular users after authentication. */
export const USER_HOME = '/home';

/**
 * Resolves the landing route that corresponds to the authenticated user's role.
 * Administrators keep the dashboard while regular users go to their home page.
 */
export function landingUrl(isAdmin: boolean): string {
  return isAdmin ? ADMIN_HOME : USER_HOME;
}
