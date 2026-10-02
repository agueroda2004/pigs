import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { MedicationsService } from '../../../core/medications/medications.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { CreateMedicationModal } from './create-medication-modal';

class MedicationsStub {
  createMedication = vi.fn(() => of({}));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('CreateMedicationModal', () => {
  let stub: MedicationsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new MedicationsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [CreateMedicationModal],
      providers: [
        { provide: MedicationsService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => TestBed.createComponent(CreateMedicationModal).componentInstance as any;

  it('does not submit when the form is empty', async () => {
    const component = create();
    await component.submit();
    expect(stub.createMedication).not.toHaveBeenCalled();
  });

  it('creates a medication and emits the name', async () => {
    const component = create();
    const created = vi.fn();
    component.created.subscribe(created);
    component.form.setValue({ name: 'Ivermectina' });

    await component.submit();

    expect(stub.createMedication).toHaveBeenCalledWith({ name: 'Ivermectina' });
    expect(created).toHaveBeenCalledWith('Ivermectina');
  });

  it('shows an error toast when the name is taken', async () => {
    stub.createMedication = vi.fn(() => throwError(() => new HttpErrorResponse({ status: 409 })));
    const component = create();
    component.form.setValue({ name: 'Ivermectina' });

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith('El nombre del medicamento ya existe');
  });
});
