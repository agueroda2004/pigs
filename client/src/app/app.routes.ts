import { Routes } from '@angular/router';

import { adminGuard, authGuard, guestGuard } from './core/auth/auth.guards';

export const routes: Routes = [
  {
    path: 'login',
    canActivate: [guestGuard],
    loadComponent: () => import('./features/auth/login/login').then((m) => m.Login),
  },
  {
    path: '',
    canActivate: [authGuard],
    loadComponent: () => import('./features/layout/shell/shell').then((m) => m.Shell),
    children: [
      {
        path: '',
        loadComponent: () => import('./features/dashboard/dashboard').then((m) => m.Dashboard),
      },
      {
        path: 'breeds',
        loadComponent: () =>
          import('./features/breeds/breeds-page/breeds-page').then((m) => m.BreedsPage),
      },
      {
        path: 'boars',
        loadComponent: () =>
          import('./features/boars/boars-page/boars-page').then((m) => m.BoarsPage),
      },
      {
        path: 'sows',
        loadComponent: () => import('./features/sows/sows-page/sows-page').then((m) => m.SowsPage),
      },
      {
        path: 'operators',
        canActivate: [adminGuard],
        loadComponent: () =>
          import('./features/operators/operators-page/operators-page').then((m) => m.OperatorsPage),
      },
      {
        path: 'users',
        canActivate: [adminGuard],
        loadComponent: () =>
          import('./features/users/users-page/users-page').then((m) => m.UsersPage),
      },
    ],
  },
  {
    path: '**',
    loadComponent: () => import('./features/not-found/not-found').then((m) => m.NotFound),
  },
];
