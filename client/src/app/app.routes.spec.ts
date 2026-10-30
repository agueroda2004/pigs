import { signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { Router, provideRouter } from '@angular/router';
import { RouterTestingHarness } from '@angular/router/testing';

import { routes } from './app.routes';
import { AuthUser } from './core/auth/auth.models';
import { AuthService } from './core/auth/auth.service';

function authStub(user: AuthUser) {
  const current = signal<AuthUser | null>(user);

  return {
    user: current.asReadonly(),
    isAuthenticated: () => current() !== null,
    isAdmin: () => current()?.role === 'Admin',
  };
}

describe('app routes', () => {
  const setup = async (user: AuthUser): Promise<RouterTestingHarness> => {
    await TestBed.configureTestingModule({
      providers: [provideRouter(routes), { provide: AuthService, useValue: authStub(user) }],
    }).compileComponents();

    return RouterTestingHarness.create();
  };

  it('shows the dashboard to administrators at the root', async () => {
    const harness = await setup({ id: '1', username: 'ana', role: 'Admin' });

    await harness.navigateByUrl('/');

    expect(TestBed.inject(Router).url).toBe('/');
    expect(harness.routeNativeElement?.textContent).toContain('Sesión iniciada como Admin');
  });

  it('sends regular users from the root to their home page', async () => {
    const harness = await setup({ id: '2', username: 'luis', role: 'User' });

    await harness.navigateByUrl('/');

    expect(TestBed.inject(Router).url).toBe('/home');
    expect(harness.routeNativeElement?.textContent).toContain('Hola, luis');
  });

  it('blocks admin-only routes for regular users', async () => {
    const harness = await setup({ id: '2', username: 'luis', role: 'User' });

    await harness.navigateByUrl('/users');

    expect(TestBed.inject(Router).url).toBe('/home');
  });

  it('renders the reports placeholder for regular users', async () => {
    const harness = await setup({ id: '2', username: 'luis', role: 'User' });

    await harness.navigateByUrl('/reports');

    expect(TestBed.inject(Router).url).toBe('/reports');
    expect(harness.routeNativeElement?.textContent).toContain('Reportes');
  });

  it('renders the sow history placeholder for regular users', async () => {
    const harness = await setup({ id: '2', username: 'luis', role: 'User' });

    await harness.navigateByUrl('/sow-history');

    expect(TestBed.inject(Router).url).toBe('/sow-history');
    expect(harness.routeNativeElement?.textContent).toContain('Historial de cerdas');
  });
});
