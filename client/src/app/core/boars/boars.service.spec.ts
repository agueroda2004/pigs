import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { firstValueFrom } from 'rxjs';

import { BoarsService } from './boars.service';

const boarPayload = {
  id: '1',
  code: 'B-001',
  location: null,
  active: true,
  entry_date: '2026-01-10',
  birth_date: null,
  note: null,
  state: 'Vivo',
  origin: 'Propio',
  breed_id: 'breed-1',
};

describe('BoarsService', () => {
  let service: BoarsService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(BoarsService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('lists boars', async () => {
    const promise = firstValueFrom(service.listBoars());

    const call = http.expectOne(
      (request) => request.url.endsWith('/boars') && request.method === 'GET',
    );
    expect(call.request.params.keys()).toHaveLength(0);
    call.flush([boarPayload]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('lists boars with filters', async () => {
    const promise = firstValueFrom(
      service.listBoars({
        code: 'B-001',
        breed_id: 'breed-1',
        origin: 'Externo',
        active: false,
      }),
    );

    const call = http.expectOne(
      (request) => request.url.endsWith('/boars') && request.method === 'GET',
    );
    expect(call.request.params.get('code')).toBe('B-001');
    expect(call.request.params.get('breed_id')).toBe('breed-1');
    expect(call.request.params.get('origin')).toBe('Externo');
    expect(call.request.params.get('active')).toBe('false');
    call.flush([boarPayload]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('creates a boar', async () => {
    const payload = {
      code: 'B-001',
      entry_date: '2026-01-10',
      origin: 'Propio' as const,
      breed_id: 'breed-1',
    };
    const promise = firstValueFrom(service.createBoar(payload));

    const call = http.expectOne(
      (request) => request.url.endsWith('/boars') && request.method === 'POST',
    );
    expect(call.request.body).toEqual(payload);
    call.flush(null);

    await expect(promise).resolves.toBeNull();
  });

  it('updates a boar', async () => {
    const payload = { code: 'B-002', note: '' };
    const promise = firstValueFrom(service.updateBoar('7', payload));

    const call = http.expectOne(
      (request) => request.url.endsWith('/boars/7') && request.method === 'PATCH',
    );
    expect(call.request.body).toEqual(payload);
    call.flush(null);

    await expect(promise).resolves.toBeNull();
  });

  it('lists the boar dropdown with the active filter', async () => {
    const promise = firstValueFrom(service.listBoarDropdown(true));

    const call = http.expectOne(
      (request) => request.url.endsWith('/boars/dropdown') && request.method === 'GET',
    );
    expect(call.request.params.get('active')).toBe('true');
    expect(call.request.params.has('state')).toBe(false);
    call.flush([{ id: '1', code: 'B-001', active: true }]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('lists the boar dropdown with the state filter', async () => {
    const promise = firstValueFrom(service.listBoarDropdown(false, 'Vivo'));

    const call = http.expectOne(
      (request) => request.url.endsWith('/boars/dropdown') && request.method === 'GET',
    );
    expect(call.request.params.get('active')).toBe('false');
    expect(call.request.params.get('state')).toBe('Vivo');
    call.flush([]);

    await expect(promise).resolves.toHaveLength(0);
  });

  it('deletes a boar', async () => {
    const promise = firstValueFrom(service.deleteBoar('7'));

    const call = http.expectOne(
      (request) => request.url.endsWith('/boars/7') && request.method === 'DELETE',
    );
    call.flush(null);

    await expect(promise).resolves.toBeNull();
  });
});
