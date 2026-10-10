import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { firstValueFrom } from 'rxjs';

import { AbortionsService } from './abortions.service';

const abortionPayload = {
  id: '1',
  sow_id: 'sow-1',
  sow_code: 'C-001',
  service_id: 'service-1',
  abortion_date: '2026-01-15',
  cause: 'Infeccioso',
  note: null,
};

const pagePayload = {
  items: [abortionPayload],
  total: 1,
  page: 1,
  page_size: 10,
  total_pages: 1,
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

  it('lists the first page of abortions', async () => {
    const promise = firstValueFrom(service.listAbortions());

    const call = http.expectOne(
      (request) => request.url.endsWith('/abortions') && request.method === 'GET',
    );
    expect(call.request.params.get('page')).toBe('1');
    call.flush(pagePayload);

    await expect(promise).resolves.toEqual(pagePayload);
  });

  it('lists abortions with the sow code filter and page', async () => {
    const promise = firstValueFrom(service.listAbortions({ sow_code: 'C-001' }, 3));

    const call = http.expectOne(
      (request) => request.url.endsWith('/abortions') && request.method === 'GET',
    );
    expect(call.request.params.get('sow_code')).toBe('C-001');
    expect(call.request.params.get('page')).toBe('3');
    call.flush({ ...pagePayload, page: 3, total: 25, total_pages: 3 });

    await expect(promise).resolves.toMatchObject({ page: 3, total: 25, total_pages: 3 });
  });

  it('updates an abortion without expecting a body', async () => {
    const payload = { abortion_date: '2026-01-20', cause: 'Traumatismo' as const, note: '' };
    const promise = firstValueFrom(service.updateAbortion('7', payload));

    const call = http.expectOne(
      (request) => request.url.endsWith('/abortions/7') && request.method === 'PATCH',
    );
    expect(call.request.body).toEqual(payload);
    call.flush(null);

    await expect(promise).resolves.toBeNull();
  });

  it('deletes an abortion', async () => {
    const promise = firstValueFrom(service.deleteAbortion('7'));

    const call = http.expectOne(
      (request) => request.url.endsWith('/abortions/7') && request.method === 'DELETE',
    );
    call.flush(null);

    await expect(promise).resolves.toBeNull();
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
    call.flush(null);

    await expect(promise).resolves.toBeNull();
  });
});
