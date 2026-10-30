import { TestBed } from '@angular/core/testing';
import { UrlTree, provideRouter } from '@angular/router';

import { AuthService } from './auth.service';
import { adminGuard, authGuard, guestGuard } from './auth.guards';

class AuthStub {
  authenticated = false;
  admin = false;

  isAuthenticated(): boolean {
    return this.authenticated;
  }

  isAdmin(): boolean {
    return this.admin;
  }
}

describe('auth guards', () => {
  let auth: AuthStub;

  beforeEach(() => {
    auth = new AuthStub();
    TestBed.configureTestingModule({
      providers: [provideRouter([]), { provide: AuthService, useValue: auth }],
    });
  });

  const run = (guard: typeof authGuard, url = '/dashboard') =>
    TestBed.runInInjectionContext(() => guard({} as never, { url } as never));

  it('allows authenticated users through authGuard', () => {
    auth.authenticated = true;
    expect(run(authGuard)).toBe(true);
  });

  it('redirects anonymous users to login with returnUrl', () => {
    const result = run(authGuard, '/dashboard');
    expect(result).toBeInstanceOf(UrlTree);
    expect((result as UrlTree).toString()).toContain('/login');
    expect((result as UrlTree).queryParams['returnUrl']).toBe('/dashboard');
  });

  it('redirects regular users away from the guest route to their home page', () => {
    auth.authenticated = true;
    const result = run(guestGuard);
    expect(result).toBeInstanceOf(UrlTree);
    expect((result as UrlTree).toString()).toBe('/home');
  });

  it('redirects admins away from the guest route to the dashboard', () => {
    auth.authenticated = true;
    auth.admin = true;
    const result = run(guestGuard);
    expect(result).toBeInstanceOf(UrlTree);
    expect((result as UrlTree).toString()).toBe('/');
  });

  it('allows admins through adminGuard', () => {
    auth.authenticated = true;
    auth.admin = true;
    expect(run(adminGuard)).toBe(true);
  });

  it('redirects regular users from admin routes to their home page', () => {
    auth.authenticated = true;
    const result = run(adminGuard);
    expect(result).toBeInstanceOf(UrlTree);
    expect((result as UrlTree).toString()).toBe('/home');
  });

  it('redirects anonymous users from admin routes to login', () => {
    const result = run(adminGuard);
    expect(result).toBeInstanceOf(UrlTree);
    expect((result as UrlTree).toString()).toBe('/login');
  });
});
