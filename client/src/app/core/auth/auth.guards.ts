import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';

import { landingUrl } from './auth.navigation';
import { AuthService } from './auth.service';

/** Allows authenticated users through and otherwise sends them to login with a return URL. */
export const authGuard: CanActivateFn = (_route, state) => {
  const auth = inject(AuthService);
  const router = inject(Router);

  if (auth.isAuthenticated()) {
    return true;
  }

  return router.createUrlTree(['/login'], { queryParams: { returnUrl: state.url } });
};

/** Allows administrators through and redirects any other user to their role-based landing. */
export const adminGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  const router = inject(Router);

  if (auth.isAdmin()) {
    return true;
  }

  return router.createUrlTree([auth.isAuthenticated() ? landingUrl(false) : '/login']);
};

/** Sends authenticated users to their role-based landing and lets guests reach the login page. */
export const guestGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  const router = inject(Router);

  if (!auth.isAuthenticated()) {
    return true;
  }

  return router.createUrlTree([landingUrl(auth.isAdmin())]);
};
