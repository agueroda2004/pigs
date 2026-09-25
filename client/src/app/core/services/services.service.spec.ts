import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { firstValueFrom } from 'rxjs';

import { ServicesService } from './services.service';

const servicePayload = {
  id: '1',
  sow_id: 'sow-1',
  expected_farrowing_date: '2026-05-04',
  note: null,
  state: 'Confirmado',
  location: null,
  mounts: [
    {
      id: 'm1',
      service_id: '1',
      boar_id: 'boar-1',
      operator_id: 'operator-1',
      mount_number: 1,
      mount_date: '2026-01-10',
      type: 'Artificial',
      note: null,
      created_at: '',
      updated_at: '',
      created_by: 'admin',
      updated_by: 'admin',
    },
  ],
  created_at: '',
  updated_at: '',
  created_by: 'admin',
  updated_by: 'admin',
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

  it('lists services', async () => {
    const promise = firstValueFrom(service.listServices());

    const call = http.expectOne(
      (request) => request.url.endsWith('/services') && request.method === 'GET',
    );
    expect(call.request.params.keys()).toHaveLength(0);
    call.flush([servicePayload]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('lists services with filters', async () => {
    const promise = firstValueFrom(service.listServices({ sow_id: 'sow-1', state: 'Fallido' }));

    const call = http.expectOne(
      (request) => request.url.endsWith('/services') && request.method === 'GET',
    );
    expect(call.request.params.get('sow_id')).toBe('sow-1');
    expect(call.request.params.get('state')).toBe('Fallido');
    call.flush([servicePayload]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('creates a service with its mounts', async () => {
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
    call.flush(servicePayload);

    await expect(promise).resolves.toMatchObject({ sow_id: 'sow-1' });
  });
});
