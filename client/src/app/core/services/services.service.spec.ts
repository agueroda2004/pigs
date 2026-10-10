import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { firstValueFrom } from 'rxjs';

import { ServicesService } from './services.service';

const servicePayload = {
  id: '1',
  sow_id: 'sow-1',
  sow_code: 'C-001',
  expected_farrowing_date: '2026-05-04',
  note: null,
  state: 'Confirmado',
  location: null,
  mounts: [
    {
      id: 'm1',
      service_id: '1',
      boar_id: 'boar-1',
      boar_code: 'V-001',
      operator_id: 'operator-1',
      operator_name: 'Ana',
      mount_number: 1,
      mount_date: '2026-01-10',
      type: 'Artificial',
      note: null,
    },
  ],
};

const pagePayload = {
  items: [servicePayload],
  total: 1,
  page: 1,
  page_size: 10,
  total_pages: 1,
};

describe('ServicesService', () => {
  let service: ServicesService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(ServicesService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('lists the first page of services', async () => {
    const promise = firstValueFrom(service.listServices());

    const call = http.expectOne(
      (request) => request.url.endsWith('/services') && request.method === 'GET',
    );
    expect(call.request.params.get('page')).toBe('1');
    call.flush(pagePayload);

    await expect(promise).resolves.toEqual(pagePayload);
  });

  it('lists services with filters and page', async () => {
    const promise = firstValueFrom(
      service.listServices({ sow_code: 'C-001', state: 'Fallido' }, 3),
    );

    const call = http.expectOne(
      (request) => request.url.endsWith('/services') && request.method === 'GET',
    );
    expect(call.request.params.get('sow_code')).toBe('C-001');
    expect(call.request.params.get('state')).toBe('Fallido');
    expect(call.request.params.get('page')).toBe('3');
    call.flush({ ...pagePayload, page: 3, total: 25, total_pages: 3 });

    await expect(promise).resolves.toMatchObject({ page: 3, total: 25, total_pages: 3 });
  });

  it('creates a service with its mounts without expecting a body', async () => {
    const payload = {
      sow_id: 'sow-1',
      mounts: [
        {
          boar_id: 'boar-1',
          operator_id: 'operator-1',
          mount_date: '2026-01-10',
          type: 'Artificial' as const,
        },
      ],
    };
    const promise = firstValueFrom(service.createService(payload));

    const call = http.expectOne(
      (request) => request.url.endsWith('/services') && request.method === 'POST',
    );
    expect(call.request.body).toEqual(payload);
    call.flush(null);

    await expect(promise).resolves.toBeNull();
  });

  it('updates a service without expecting a body', async () => {
    const payload = {
      location: 'Nave 2',
      note: 'ok',
      mounts: {
        create: [
          {
            boar_id: 'boar-1',
            operator_id: 'operator-1',
            mount_date: '2026-01-10',
            type: 'Natural' as const,
          },
        ],
        update: [
          {
            id: 'm1',
            boar_id: 'boar-1',
            operator_id: 'operator-1',
            mount_date: '2026-01-11',
            type: 'Artificial' as const,
          },
        ],
        delete: ['m2'],
      },
    };
    const promise = firstValueFrom(service.updateService('7', payload));

    const call = http.expectOne(
      (request) => request.url.endsWith('/services/7') && request.method === 'PATCH',
    );
    expect(call.request.body).toEqual(payload);
    call.flush(null);

    await expect(promise).resolves.toBeNull();
  });

  it('deletes a service', async () => {
    const promise = firstValueFrom(service.deleteService('7'));

    const call = http.expectOne(
      (request) => request.url.endsWith('/services/7') && request.method === 'DELETE',
    );
    call.flush(null);

    await expect(promise).resolves.toBeNull();
  });
});
