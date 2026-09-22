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
      created_at: '',
      updated_at: '',
    });

    await expect(promise).resolves.toMatchObject({ username: 'ana' });
  });
});
