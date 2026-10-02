import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { firstValueFrom } from 'rxjs';

import { MedicationsService } from './medications.service';

const medicationPayload = {
  id: '1',
  name: 'Ivermectina',
  active: true,
  created_at: '',
  updated_at: '',
  created_by: 'admin',
  updated_by: 'admin',
};

describe('MedicationsService', () => {
  let service: MedicationsService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(MedicationsService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('lists medications without filters', async () => {
    const promise = firstValueFrom(service.listMedications());

    const call = http.expectOne(
      (request) => request.url.endsWith('/medications') && request.method === 'GET',
    );
    expect(call.request.params.has('name')).toBe(false);
    expect(call.request.params.has('active')).toBe(false);
    call.flush([medicationPayload]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('lists medications with name and active filters', async () => {
    const promise = firstValueFrom(service.listMedications({ name: 'Iver', active: false }));

    const call = http.expectOne(
      (request) => request.url.endsWith('/medications') && request.method === 'GET',
    );
    expect(call.request.params.get('name')).toBe('Iver');
    expect(call.request.params.get('active')).toBe('false');
    call.flush([medicationPayload]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('lists medication options excluding inactive by default', async () => {
    const promise = firstValueFrom(service.listMedicationOptions());

    const call = http.expectOne(
      (request) => request.url.endsWith('/medications/options') && request.method === 'GET',
    );
    expect(call.request.params.get('include_inactive')).toBe('false');
    call.flush([{ id: '1', name: 'Ivermectina' }]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('lists medication options including inactive', async () => {
    const promise = firstValueFrom(service.listMedicationOptions(true));

    const call = http.expectOne(
      (request) => request.url.endsWith('/medications/options') && request.method === 'GET',
    );
    expect(call.request.params.get('include_inactive')).toBe('true');
    call.flush([]);

    await expect(promise).resolves.toHaveLength(0);
  });

  it('creates a medication', async () => {
    const promise = firstValueFrom(service.createMedication({ name: 'Ivermectina' }));

    const call = http.expectOne(
      (request) => request.url.endsWith('/medications') && request.method === 'POST',
    );
    expect(call.request.body).toEqual({ name: 'Ivermectina' });
    call.flush(medicationPayload);

    await expect(promise).resolves.toMatchObject({ name: 'Ivermectina' });
  });

  it('updates a medication', async () => {
    const payload = { name: 'Penicilina', active: false };
    const promise = firstValueFrom(service.updateMedication('7', payload));

    const call = http.expectOne(
      (request) => request.url.endsWith('/medications/7') && request.method === 'PATCH',
    );
    expect(call.request.body).toEqual(payload);
    call.flush({ ...medicationPayload, name: 'Penicilina', active: false });

    await expect(promise).resolves.toMatchObject({ name: 'Penicilina', active: false });
  });
});
