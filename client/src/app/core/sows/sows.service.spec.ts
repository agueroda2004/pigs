import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { firstValueFrom } from 'rxjs';

import { SowsService } from './sows.service';

const sowPayload = {
  id: '1',
  code: 'C-001',
  location: null,
  active: true,
  entry_date: '2026-01-10',
  birth_date: null,
  note: null,
  state: 'Viva',
  origin: 'Propio',
  parity: 3,
  breed_id: 'breed-1',
};

const pagePayload = {
  items: [sowPayload],
  total: 1,
  page: 1,
  page_size: 20,
  total_pages: 1,
};

describe('SowsService', () => {
  let service: SowsService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(SowsService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('lists the first page of sows', async () => {
    const promise = firstValueFrom(service.listSows());

    const call = http.expectOne(
      (request) => request.url.endsWith('/sows') && request.method === 'GET',
    );
    expect(call.request.params.get('page')).toBe('1');
    call.flush(pagePayload);

    await expect(promise).resolves.toEqual(pagePayload);
  });

  it('lists sows with filters and page', async () => {
    const promise = firstValueFrom(
      service.listSows(
        {
          code: 'C-001',
          breed_id: 'breed-1',
          origin: 'Externo',
          active: false,
          state: 'Gestando',
        },
        3,
      ),
    );

    const call = http.expectOne(
      (request) => request.url.endsWith('/sows') && request.method === 'GET',
    );
    expect(call.request.params.get('code')).toBe('C-001');
    expect(call.request.params.get('breed_id')).toBe('breed-1');
    expect(call.request.params.get('origin')).toBe('Externo');
    expect(call.request.params.get('active')).toBe('false');
    expect(call.request.params.get('state')).toBe('Gestando');
    expect(call.request.params.get('page')).toBe('3');
    call.flush({ ...pagePayload, page: 3, total: 40, total_pages: 2 });

    await expect(promise).resolves.toMatchObject({ page: 3, total: 40, total_pages: 2 });
  });

  it('lists the sow dropdown without filters', async () => {
    const promise = firstValueFrom(service.listSowDropdown());

    const call = http.expectOne(
      (request) => request.url.endsWith('/sows/dropdown') && request.method === 'GET',
    );
    expect(call.request.params.keys()).toHaveLength(0);
    call.flush([{ id: 'sow-1', code: 'C-001' }]);

    await expect(promise).resolves.toEqual([{ id: 'sow-1', code: 'C-001' }]);
  });

  it('lists the sow dropdown filtered by active and states', async () => {
    const promise = firstValueFrom(service.listSowDropdown(true, ['Viva', 'Gestando']));

    const call = http.expectOne(
      (request) => request.url.endsWith('/sows/dropdown') && request.method === 'GET',
    );
    expect(call.request.params.get('active')).toBe('true');
    expect(call.request.params.getAll('states')).toEqual(['Viva', 'Gestando']);
    call.flush([{ id: 'sow-1', code: 'C-001' }]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('creates a sow without expecting a body', async () => {
    const payload = {
      code: 'C-001',
      entry_date: '2026-01-10',
      origin: 'Propio' as const,
      parity: 0,
      breed_id: 'breed-1',
    };
    const promise = firstValueFrom(service.createSow(payload));

    const call = http.expectOne(
      (request) => request.url.endsWith('/sows') && request.method === 'POST',
    );
    expect(call.request.body).toEqual(payload);
    call.flush(null);

    await expect(promise).resolves.toBeNull();
  });

  it('updates a sow without expecting a body', async () => {
    const payload = { code: 'C-002', note: '' };
    const promise = firstValueFrom(service.updateSow('7', payload));

    const call = http.expectOne(
      (request) => request.url.endsWith('/sows/7') && request.method === 'PATCH',
    );
    expect(call.request.body).toEqual(payload);
    call.flush(null);

    await expect(promise).resolves.toBeNull();
  });

  it('deletes a sow', async () => {
    const promise = firstValueFrom(service.deleteSow('7'));

    const call = http.expectOne(
      (request) => request.url.endsWith('/sows/7') && request.method === 'DELETE',
    );
    call.flush(null);

    await expect(promise).resolves.toBeNull();
  });
});
