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
  created_at: '',
  updated_at: '',
  created_by: 'admin',
  updated_by: 'admin',
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

  it('lists sows', async () => {
    const promise = firstValueFrom(service.listSows());

    const call = http.expectOne(
      (request) => request.url.endsWith('/sows') && request.method === 'GET',
    );
    expect(call.request.params.keys()).toHaveLength(0);
    call.flush([sowPayload]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('lists sows with filters', async () => {
    const promise = firstValueFrom(
      service.listSows({
        code: 'C-001',
        breed_id: 'breed-1',
        origin: 'Externo',
        active: false,
        state: 'Gestando',
      }),
    );

    const call = http.expectOne(
      (request) => request.url.endsWith('/sows') && request.method === 'GET',
    );
    expect(call.request.params.get('code')).toBe('C-001');
    expect(call.request.params.get('breed_id')).toBe('breed-1');
    expect(call.request.params.get('origin')).toBe('Externo');
    expect(call.request.params.get('active')).toBe('false');
    expect(call.request.params.get('state')).toBe('Gestando');
    call.flush([sowPayload]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('lists sow options without a filter', async () => {
    const promise = firstValueFrom(service.listSowOptions());

    const call = http.expectOne(
      (request) => request.url.endsWith('/sows/options') && request.method === 'GET',
    );
    expect(call.request.params.keys()).toHaveLength(0);
    call.flush([{ id: 'sow-1', code: 'C-001' }]);

    await expect(promise).resolves.toEqual([{ id: 'sow-1', code: 'C-001' }]);
  });

  it('lists sow options filtered by active', async () => {
    const promise = firstValueFrom(service.listSowOptions(true));

    const call = http.expectOne(
      (request) => request.url.endsWith('/sows/options') && request.method === 'GET',
    );
    expect(call.request.params.get('active')).toBe('true');
    call.flush([{ id: 'sow-1', code: 'C-001' }]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('lists sow options with active false', async () => {
    const promise = firstValueFrom(service.listSowOptions(false));

    const call = http.expectOne(
      (request) => request.url.endsWith('/sows/options') && request.method === 'GET',
    );
    expect(call.request.params.get('active')).toBe('false');
    call.flush([]);

    await expect(promise).resolves.toEqual([]);
  });

  it('creates a sow', async () => {
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
    call.flush(sowPayload);

    await expect(promise).resolves.toMatchObject({ code: 'C-001' });
  });

  it('updates a sow', async () => {
    const payload = { code: 'C-002', note: '' };
    const promise = firstValueFrom(service.updateSow('7', payload));

    const call = http.expectOne(
      (request) => request.url.endsWith('/sows/7') && request.method === 'PATCH',
    );
    expect(call.request.body).toEqual(payload);
    call.flush({ ...sowPayload, code: 'C-002' });

    await expect(promise).resolves.toMatchObject({ code: 'C-002' });
  });
});
