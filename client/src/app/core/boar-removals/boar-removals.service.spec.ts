import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { firstValueFrom } from 'rxjs';

import { BoarRemovalsService } from './boar-removals.service';

const removalPayload = {
  id: '1',
  boar_id: 'boar-1',
  removal_date: '2026-01-20',
  type: 'Muerte',
  reason: 'Enfermedad',
  note: null,
  last_state: 'Vivo',
  created_at: '',
  updated_at: '',
  created_by: 'admin',
  updated_by: 'admin',
};

describe('BoarRemovalsService', () => {
  let service: BoarRemovalsService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(BoarRemovalsService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('lists removals', async () => {
    const promise = firstValueFrom(service.listBoarRemovals());

    const call = http.expectOne(
      (request) => request.url.endsWith('/boar-removals') && request.method === 'GET',
    );
    expect(call.request.params.keys()).toHaveLength(0);
    call.flush([removalPayload]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('lists removals with the boar filter', async () => {
    const promise = firstValueFrom(service.listBoarRemovals({ boar_id: 'boar-1' }));

    const call = http.expectOne(
      (request) => request.url.endsWith('/boar-removals') && request.method === 'GET',
    );
    expect(call.request.params.get('boar_id')).toBe('boar-1');
    call.flush([removalPayload]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('creates a removal', async () => {
    const payload = {
      boar_id: 'boar-1',
      removal_date: '2026-01-20',
      type: 'Muerte' as const,
      reason: 'Enfermedad' as const,
    };
    const promise = firstValueFrom(service.createBoarRemoval(payload));

    const call = http.expectOne(
      (request) => request.url.endsWith('/boar-removals') && request.method === 'POST',
    );
    expect(call.request.body).toEqual(payload);
    call.flush(removalPayload);

    await expect(promise).resolves.toMatchObject({ boar_id: 'boar-1' });
  });

  it('updates a removal', async () => {
    const payload = { reason: 'Otro' as const, note: '' };
    const promise = firstValueFrom(service.updateBoarRemoval('1', payload));

    const call = http.expectOne(
      (request) => request.url.endsWith('/boar-removals/1') && request.method === 'PATCH',
    );
    expect(call.request.body).toEqual(payload);
    call.flush(removalPayload);

    await expect(promise).resolves.toMatchObject({ id: '1' });
  });

  it('deletes a removal', async () => {
    const promise = firstValueFrom(service.deleteBoarRemoval('1'));

    const call = http.expectOne(
      (request) => request.url.endsWith('/boar-removals/1') && request.method === 'DELETE',
    );
    call.flush(null, { status: 204, statusText: 'No Content' });

    await expect(promise).resolves.toBeNull();
  });
});
