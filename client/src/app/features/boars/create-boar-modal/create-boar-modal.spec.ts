import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { BoarsService } from '../../../core/boars/boars.service';
import { BreedsService } from '../../../core/breeds/breeds.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { CreateBoarModal } from './create-boar-modal';

class BoarsStub {
  createBoar = vi.fn((_request: Record<string, unknown>) => of({}));
}

class BreedsStub {
  listBreeds = vi.fn(() => of([]));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('CreateBoarModal', () => {
  let stub: BoarsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new BoarsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [CreateBoarModal],
      providers: [
        { provide: BoarsService, useValue: stub },
        { provide: BreedsService, useValue: new BreedsStub() },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => TestBed.createComponent(CreateBoarModal).componentInstance as any;

  it('does not submit when the form is empty', async () => {
    const component = create();

    await component.submit();

    expect(stub.createBoar).not.toHaveBeenCalled();
  });

  it('creates a boar and emits the code', async () => {
    const component = create();
    const created = vi.fn();
    component.created.subscribe(created);
    component.form.setValue({
      code: 'B-001',
      breed_id: 'breed-1',
      origin: 'Propio',
      location: '',
      entry_date: '2026-01-10',
      birth_date: '',
      note: '',
    });

    await component.submit();

    expect(stub.createBoar).toHaveBeenCalledWith({
      code: 'B-001',
      entry_date: '2026-01-10',
      origin: 'Propio',
      breed_id: 'breed-1',
    });
    expect(created).toHaveBeenCalledWith('B-001');
  });

  it('omits birth_date when it is not set', async () => {
    const component = create();
    component.form.setValue({
      code: 'B-001',
      breed_id: 'breed-1',
      origin: 'Propio',
      location: '',
      entry_date: '2026-01-10',
      birth_date: '',
      note: '',
    });

    await component.submit();

    const request = stub.createBoar.mock.calls[0][0];
    expect(request).not.toHaveProperty('birth_date');
  });

  it('includes optional fields when provided', async () => {
    const component = create();
    component.form.setValue({
      code: 'B-002',
      breed_id: 'breed-2',
      origin: 'Externo',
      location: 'Corral B',
      entry_date: '2026-01-10',
      birth_date: '2025-12-01',
      note: 'Nota',
    });

    await component.submit();

    expect(stub.createBoar).toHaveBeenCalledWith({
      code: 'B-002',
      entry_date: '2026-01-10',
      origin: 'Externo',
      breed_id: 'breed-2',
      location: 'Corral B',
      birth_date: '2025-12-01',
      note: 'Nota',
    });
  });

  it('shows an error toast when the code is taken', async () => {
    stub.createBoar = vi.fn(() => throwError(() => new HttpErrorResponse({ status: 409 })));
    const component = create();
    component.form.setValue({
      code: 'B-001',
      breed_id: 'breed-1',
      origin: 'Propio',
      location: '',
      entry_date: '2026-01-10',
      birth_date: '',
      note: '',
    });

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith('El código del verraco ya existe');
  });
});
