import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { BreedsService } from '../../../core/breeds/breeds.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { SowsService } from '../../../core/sows/sows.service';
import { CreateSowModal } from './create-sow-modal';

class SowsStub {
  createSow = vi.fn((_request: Record<string, unknown>) => of({}));
}

class BreedsStub {
  listBreeds = vi.fn(() => of([]));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('CreateSowModal', () => {
  let stub: SowsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new SowsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [CreateSowModal],
      providers: [
        { provide: SowsService, useValue: stub },
        { provide: BreedsService, useValue: new BreedsStub() },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => TestBed.createComponent(CreateSowModal).componentInstance as any;

  it('does not submit when the form is empty', async () => {
    const component = create();

    await component.submit();

    expect(stub.createSow).not.toHaveBeenCalled();
  });

  it('creates a sow with a default parity and emits the code', async () => {
    const component = create();
    const created = vi.fn();
    component.created.subscribe(created);
    component.form.setValue({
      code: 'C-001',
      breed_id: 'breed-1',
      origin: 'Propio',
      parity: 0,
      location: '',
      entry_date: '2026-01-10',
      birth_date: '',
      note: '',
    });

    await component.submit();

    expect(stub.createSow).toHaveBeenCalledWith({
      code: 'C-001',
      entry_date: '2026-01-10',
      origin: 'Propio',
      parity: 0,
      breed_id: 'breed-1',
    });
    expect(created).toHaveBeenCalledWith('C-001');
  });

  it('omits birth_date when it is not set', async () => {
    const component = create();
    component.form.setValue({
      code: 'C-001',
      breed_id: 'breed-1',
      origin: 'Propio',
      parity: 0,
      location: '',
      entry_date: '2026-01-10',
      birth_date: '',
      note: '',
    });

    await component.submit();

    const request = stub.createSow.mock.calls[0][0];
    expect(request).not.toHaveProperty('birth_date');
  });

  it('includes optional fields when provided', async () => {
    const component = create();
    component.form.setValue({
      code: 'C-002',
      breed_id: 'breed-2',
      origin: 'Externo',
      parity: 4,
      location: 'Corral B',
      entry_date: '2026-01-10',
      birth_date: '2025-12-01',
      note: 'Nota',
    });

    await component.submit();

    expect(stub.createSow).toHaveBeenCalledWith({
      code: 'C-002',
      entry_date: '2026-01-10',
      origin: 'Externo',
      parity: 4,
      breed_id: 'breed-2',
      location: 'Corral B',
      birth_date: '2025-12-01',
      note: 'Nota',
    });
  });

  it('shows an error toast when the code is taken', async () => {
    stub.createSow = vi.fn(() => throwError(() => new HttpErrorResponse({ status: 409 })));
    const component = create();
    component.form.setValue({
      code: 'C-001',
      breed_id: 'breed-1',
      origin: 'Propio',
      parity: 0,
      location: '',
      entry_date: '2026-01-10',
      birth_date: '',
      note: '',
    });

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith('El código de la cerda ya existe');
  });
});
