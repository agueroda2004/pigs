import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { firstValueFrom } from 'rxjs';

import { UsersService } from './users.service';

describe('UsersService', () => {
  let service: UsersService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(UsersService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('creates a user with the given payload', async () => {
    const payload = { name: 'Ana', username: 'ana', password: 'Password1!', role: 'User' as const };
    const promise = firstValueFrom(service.createUser(payload));

    const call = http.expectOne(
      (request) => request.url.endsWith('/users') && request.method === 'POST',
    );
    expect(call.request.body).toEqual(payload);
    call.flush({
      id: '1',
      name: 'Ana',
      username: 'ana',
      role: 'User',
      active: true,
      created_at: '',
      updated_at: '',
    });

    await expect(promise).resolves.toMatchObject({ username: 'ana' });
  });

  it('lists users from the admin endpoint', async () => {
    const promise = firstValueFrom(service.listUsers());

    const call = http.expectOne(
      (request) => request.url.endsWith('/admin/users') && request.method === 'GET',
    );
    call.flush([
      {
        id: '1',
        name: 'Ana',
        username: 'ana',
        role: 'User',
        active: true,
        created_at: '',
        updated_at: '',
      },
    ]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('updates a user through the admin endpoint', async () => {
    const payload = { name: 'Ana Updated', username: 'ana', role: 'Admin' as const, active: false };
    const promise = firstValueFrom(service.updateUser('42', payload));

    const call = http.expectOne(
      (request) => request.url.endsWith('/admin/users/42') && request.method === 'PATCH',
    );
    expect(call.request.body).toEqual(payload);
    call.flush({
      id: '42',
      name: 'Ana Updated',
      username: 'ana',
      role: 'Admin',
      active: false,
      created_at: '',
      updated_at: '',
    });

    await expect(promise).resolves.toMatchObject({
      name: 'Ana Updated',
      role: 'Admin',
      active: false,
    });
  });
});
