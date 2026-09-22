import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { vi } from 'vitest';

import { AuthService } from './auth.service';

describe('AuthService', () => {
  let service: AuthService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(AuthService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('starts unauthenticated', () => {
    expect(service.isAuthenticated()).toBe(false);
    expect(service.user()).toBeNull();
  });

  it('loads the current user', async () => {
    const promise = service.loadCurrentUser();
    http
      .expectOne((request) => request.url.endsWith('/auth/me'))
      .flush({ id: '1', username: 'ana', role: 'Admin' });

    await expect(promise).resolves.toEqual({ id: '1', username: 'ana', role: 'Admin' });
    expect(service.isAuthenticated()).toBe(true);
    expect(service.isAdmin()).toBe(true);
  });

  it('clears the user when loading fails', async () => {
    const promise = service.loadCurrentUser();
    http
      .expectOne((request) => request.url.endsWith('/auth/me'))
      .flush(null, { status: 401, statusText: 'Unauthorized' });

    await expect(promise).resolves.toBeNull();
    expect(service.isAuthenticated()).toBe(false);
  });

  it('logs in and then loads the user', async () => {
    const promise = service.login({ username: 'ana', password: 'secret' });
    http.expectOne((request) => request.url.endsWith('/auth/login')).flush({});

    await vi.waitFor(() => {
      http
        .expectOne((request) => request.url.endsWith('/auth/me'))
        .flush({ id: '1', username: 'ana', role: 'User' });
    });

    await promise;
    expect(service.user()).toEqual({ id: '1', username: 'ana', role: 'User' });
    expect(service.isAdmin()).toBe(false);
  });

  it('clears the user on logout even if the request fails', async () => {
    const load = service.loadCurrentUser();
    http
      .expectOne((request) => request.url.endsWith('/auth/me'))
      .flush({ id: '1', username: 'ana', role: 'User' });
    await load;

    const promise = service.logout();
    http
      .expectOne((request) => request.url.endsWith('/auth/logout'))
      .flush(null, { status: 500, statusText: 'Server Error' });

    await expect(promise).rejects.toBeTruthy();
    expect(service.isAuthenticated()).toBe(false);
  });
});
