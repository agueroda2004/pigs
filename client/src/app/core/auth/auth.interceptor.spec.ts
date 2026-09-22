import { HttpClient, provideHttpClient, withInterceptors } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { firstValueFrom } from 'rxjs';
import { vi } from 'vitest';

import { authInterceptor } from './auth.interceptor';
import { credentialsInterceptor } from './credentials.interceptor';

describe('auth interceptors', () => {
  let http: HttpClient;
  let controller: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [
        provideRouter([]),
        provideHttpClient(withInterceptors([credentialsInterceptor, authInterceptor])),
        provideHttpClientTesting(),
      ],
    });
    http = TestBed.inject(HttpClient);
    controller = TestBed.inject(HttpTestingController);
  });

  afterEach(() => controller.verify());

  it('sends credentials with every request', () => {
    http.get('/api/v1/users').subscribe();
    const request = controller.expectOne('/api/v1/users');

    expect(request.request.withCredentials).toBe(true);
    request.flush({});
  });

  it('refreshes the session and retries on 401', async () => {
    const result = firstValueFrom(http.get<{ ok: boolean }>('/api/v1/users'));

    controller.expectOne('/api/v1/users').flush(null, { status: 401, statusText: 'Unauthorized' });

    controller.expectOne((request) => request.url.endsWith('/auth/refresh')).flush({});

    await vi.waitFor(() => {
      controller.expectOne('/api/v1/users').flush({ ok: true });
    });

    await expect(result).resolves.toEqual({ ok: true });
  });

  it('does not refresh when an auth endpoint fails', async () => {
    const result = firstValueFrom(http.post('/api/v1/auth/login', {}));

    controller
      .expectOne('/api/v1/auth/login')
      .flush(null, { status: 401, statusText: 'Unauthorized' });

    await expect(result).rejects.toBeTruthy();
    controller.verify();
  });
});
