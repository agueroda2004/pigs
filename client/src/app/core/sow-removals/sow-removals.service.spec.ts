import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { firstValueFrom } from 'rxjs';

import { SowRemovalsService } from './sow-removals.service';

const removalPayload = {
  id: '1',
  sow_id: 'sow-1',
  removal_date: '2026-01-20',
  type: 'Muerte',
  reason: 'Enfermedad',
  note: null,
  last_state: 'Gestando',
  created_at: '',
  updated_at: '',
  created_by: 'admin',
  updated_by: 'admin',
};

describe('SowRemovalsService', () => {
  let service: SowRemovalsService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(SowRemovalsService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('lists removals', async () => {
    const promise = firstValueFrom(service.listSowRemovals());

    const call = http.expectOne(
      (request) => request.url.endsWith('/sow-removals') && request.method === 'GET',
    );
    expect(call.request.params.keys()).toHaveLength(0);
    call.flush([removalPayload]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('lists removals with the sow filter', async () => {
    const promise = firstValueFrom(service.listSowRemovals({ sow_id: 'sow-1' }));

    const call = http.expectOne(
      (request) => request.url.endsWith('/sow-removals') && request.method === 'GET',
    );
    expect(call.request.params.get('sow_id')).toBe('sow-1');
    call.flush([removalPayload]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('creates a removal', async () => {
    const payload = {
      sow_id: 'sow-1',
      removal_date: '2026-01-20',
      type: 'Muerte' as const,
      reason: 'Enfermedad' as const,
    };
    const promise = firstValueFrom(service.createSowRemoval(payload));

    const call = http.expectOne(
      (request) => request.url.endsWith('/sow-removals') && request.method === 'POST',
    );
    expect(call.request.body).toEqual(payload);
    call.flush(removalPayload);

    await expect(promise).resolves.toMatchObject({ sow_id: 'sow-1' });
  });
});
