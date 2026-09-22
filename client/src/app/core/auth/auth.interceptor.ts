import { HttpErrorResponse, HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { catchError, from, switchMap, throwError } from 'rxjs';

import { AuthService } from './auth.service';

function isAuthRequest(url: string): boolean {
  return url.includes('/auth/');
}

export const authInterceptor: HttpInterceptorFn = (request, next) => {
  const auth = inject(AuthService);
  const router = inject(Router);

  return next(request).pipe(
    catchError((error: HttpErrorResponse) => {
      if (error.status !== 401 || isAuthRequest(request.url)) {
        return throwError(() => error);
      }

      return from(auth.refresh()).pipe(
        switchMap(() => next(request)),
        catchError(() => {
          auth.clear();
          void router.navigate(['/login']);
          return throwError(() => error);
        }),
      );
    }),
  );
};
