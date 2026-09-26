import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { firstValueFrom } from 'rxjs';

import { AbortionsService } from './abortions.service';

const abortionPayload = {
  id: '1',
  sow_id: 'sow-1',
  service_id: 'service-1',
  abortion_date: '2026-01-15',
  cause: 'Infeccioso',
  note: null,
  created_at: '',
  updated_at: '',
  created_by: 'admin',
  updated_by: 'admin',
};

describe('AbortionsService', () => {
  let service: AbortionsService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(AbortionsService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('lists abortions', async () => {
    const promise = firstValueFrom(service.listAbortions());

    const call = http.expectOne(
      (request) => request.url.endsWith('/abortions') && request.method === 'GET',
    );
    expect(call.request.params.keys()).toHaveLength(0);
    call.flush([abortionPayload]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('lists abortions with the sow filter', async () => {
    const promise = firstValueFrom(service.listAbortions({ sow_id: 'sow-1' }));

    const call = http.expectOne(
      (request) => request.url.endsWith('/abortions') && request.method === 'GET',
    );
    expect(call.request.params.get('sow_id')).toBe('sow-1');
    call.flush([abortionPayload]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('creates an abortion', async () => {
    const payload = {
      sow_id: 'sow-1',
      abortion_date: '2026-01-15',
      cause: 'Infeccioso' as const,
    };
    const promise = firstValueFrom(service.createAbortion(payload));

    const call = http.expectOne(
      (request) => request.url.endsWith('/abortions') && request.method === 'POST',
    );
    expect(call.request.body).toEqual(payload);
    call.flush(abortionPayload);

    await expect(promise).resolves.toMatchObject({ sow_id: 'sow-1' });
  });
});
