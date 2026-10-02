import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { Medication } from '../../../core/medications/medication.models';
import { MedicationsService } from '../../../core/medications/medications.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { EditMedicationModal } from './edit-medication-modal';

function buildMedication(): Medication {
  return {
    id: '42',
    name: 'Ivermectina',
    active: true,
    created_at: '2026-01-02T12:00:00',
    updated_at: '2026-01-02T12:00:00',
    created_by: 'admin',
    updated_by: 'admin',
  };
}

class MedicationsStub {
  updateMedication = vi.fn(() => of(buildMedication()));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('EditMedicationModal', () => {
  let stub: MedicationsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new MedicationsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [EditMedicationModal],
      providers: [
        { provide: MedicationsService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => {
    const fixture = TestBed.createComponent(EditMedicationModal);
    fixture.componentRef.setInput('medication', buildMedication());
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();
    return fixture.componentInstance as any;
  };

  it('prefills the form from the selected medication', () => {
    const component = create();

    expect(component.form.getRawValue()).toEqual({ name: 'Ivermectina', active: true });
  });

  it('updates the medication with changed fields and emits the result', async () => {
    const component = create();
    const updated = vi.fn();
    component.updated.subscribe(updated);
    component.form.controls.name.setValue('Penicilina');
    component.form.controls.active.setValue(false);

    await component.submit();

    expect(stub.updateMedication).toHaveBeenCalledWith('42', {
      name: 'Penicilina',
      active: false,
    });
    expect(updated).toHaveBeenCalled();
  });

  it('does not call the API when nothing changed', async () => {
    const component = create();
    const updated = vi.fn();
    component.updated.subscribe(updated);

    await component.submit();

    expect(stub.updateMedication).not.toHaveBeenCalled();
    expect(updated).toHaveBeenCalled();
  });

  it('shows an error toast when the name is taken', async () => {
    stub.updateMedication = vi.fn(() => throwError(() => new HttpErrorResponse({ status: 409 })));
    const component = create();
    component.form.controls.name.setValue('Penicilina');

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith('El nombre del medicamento ya existe');
  });
});
