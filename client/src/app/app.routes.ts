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
        path: 'boar-removals',
        loadComponent: () =>
          import('./features/boar-removals/boar-removals-page/boar-removals-page').then(
            (m) => m.BoarRemovalsPage,
          ),
      },
      {
        path: 'sows',
        loadComponent: () => import('./features/sows/sows-page/sows-page').then((m) => m.SowsPage),
      },
      {
        path: 'services',
        loadComponent: () =>
          import('./features/services/services-page/services-page').then((m) => m.ServicesPage),
      },
      {
        path: 'abortions',
        loadComponent: () =>
          import('./features/abortions/abortions-page/abortions-page').then((m) => m.AbortionsPage),
      },
      {
        path: 'farrowings',
        loadComponent: () =>
          import('./features/farrowings/farrowings-page/farrowings-page').then(
            (m) => m.FarrowingsPage,
          ),
      },
      {
        path: 'sow-removals',
        loadComponent: () =>
          import('./features/sow-removals/sow-removals-page/sow-removals-page').then(
            (m) => m.SowRemovalsPage,
          ),
      },
      {
        path: 'medications',
        loadComponent: () =>
          import('./features/medications/medications-page/medications-page').then(
            (m) => m.MedicationsPage,
          ),
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
