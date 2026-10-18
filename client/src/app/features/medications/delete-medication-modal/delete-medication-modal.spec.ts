import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { Medication } from '../../../core/medications/medication.models';
import { MedicationsService } from '../../../core/medications/medications.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { DeleteMedicationModal } from './delete-medication-modal';

function buildMedication(): Medication {
  return {
    id: '42',
    name: 'Ivermectina',
    active: true,
  };
}

class MedicationsStub {
  deleteMedication = vi.fn(() => of(undefined));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('DeleteMedicationModal', () => {
  let stub: MedicationsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new MedicationsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [DeleteMedicationModal],
      providers: [
        { provide: MedicationsService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  const create = () => {
    const fixture = TestBed.createComponent(DeleteMedicationModal);
    fixture.componentRef.setInput('medication', buildMedication());
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return fixture.componentInstance as any;
  };

  it('does nothing when there is no medication', async () => {
    const fixture = TestBed.createComponent(DeleteMedicationModal);
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    await (fixture.componentInstance as any).confirm();

    expect(stub.deleteMedication).not.toHaveBeenCalled();
  });

  it('deletes the medication and emits it', async () => {
    const component = create();
    const deleted = vi.fn();
    component.deleted.subscribe(deleted);

    await component.confirm();

    expect(stub.deleteMedication).toHaveBeenCalledWith('42');
    expect(deleted).toHaveBeenCalledWith(buildMedication());
  });

  it('shows the server message when the medication has linked records', async () => {
    stub.deleteMedication = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 409,
            error: {
              error: 'No se puede eliminar este medicamento porque tiene registros enlazados.',
            },
          }),
      ),
    );
    const component = create();

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith(
      'No se puede eliminar este medicamento porque tiene registros enlazados.',
    );
  });

  it('shows a fallback when the medication is missing', async () => {
    stub.deleteMedication = vi.fn(() => throwError(() => new HttpErrorResponse({ status: 404 })));
    const component = create();

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith('El medicamento no existe');
  });
});
