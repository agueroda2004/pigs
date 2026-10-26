import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { PigletFostering } from '../../../core/piglet-fosterings/piglet-fostering.models';
import { PigletFosteringsService } from '../../../core/piglet-fosterings/piglet-fosterings.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { DeletePigletFosteringModal } from './delete-piglet-fostering-modal';

function buildFostering(): PigletFostering {
  return {
    id: '42',
    donor_farrowing_id: 'farrowing-1',
    receiver_farrowing_id: 'farrowing-2',
    donor_sow_id: 'sow-1',
    receiver_sow_id: 'sow-2',
    movement_date: '2026-04-25',
    quantity: 2,
    note: null,
    created_at: '',
    updated_at: '',
    created_by: 'admin',
    updated_by: 'admin',
  };
}

class PigletFosteringsStub {
  deletePigletFostering = vi.fn(() => of(void 0));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('DeletePigletFosteringModal', () => {
  let stub: PigletFosteringsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new PigletFosteringsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [DeletePigletFosteringModal],
      providers: [
        { provide: PigletFosteringsService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  const create = () => {
    const fixture = TestBed.createComponent(DeletePigletFosteringModal);
    fixture.componentRef.setInput('fostering', buildFostering());
    fixture.componentRef.setInput('donorSowCode', 'C-001');
    fixture.componentRef.setInput('receiverSowCode', 'C-002');
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return fixture.componentInstance as any;
  };

  it('does nothing when there is no fostering', async () => {
    const fixture = TestBed.createComponent(DeletePigletFosteringModal);
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    await (fixture.componentInstance as any).confirm();

    expect(stub.deletePigletFostering).not.toHaveBeenCalled();
  });

  it('deletes the fostering and emits it', async () => {
    const component = create();
    const deleted = vi.fn();
    component.deleted.subscribe(deleted);

    await component.confirm();

    expect(stub.deletePigletFostering).toHaveBeenCalledWith('42');
    expect(deleted).toHaveBeenCalledWith(buildFostering());
    expect(component.loading()).toBe(false);
  });

  it('shows the server message when a later weaging blocks the deletion', async () => {
    stub.deletePigletFostering = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 409,
            error: {
              error: 'No se puede eliminar el traslado porque existe un destete parcial posterior',
            },
          }),
      ),
    );
    const component = create();
    const deleted = vi.fn();
    component.deleted.subscribe(deleted);

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith(
      'No se puede eliminar el traslado porque existe un destete parcial posterior',
    );
    expect(deleted).not.toHaveBeenCalled();
    expect(component.loading()).toBe(false);
  });

  it('shows a fallback when the server returns no message', async () => {
    stub.deletePigletFostering = vi.fn(() =>
      throwError(() => new HttpErrorResponse({ status: 500 })),
    );
    const component = create();
    const deleted = vi.fn();
    component.deleted.subscribe(deleted);

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith('No se pudo eliminar el traslado');
    expect(deleted).not.toHaveBeenCalled();
  });

  it('shows a fallback when the fostering is missing', async () => {
    stub.deletePigletFostering = vi.fn(() =>
      throwError(() => new HttpErrorResponse({ status: 404 })),
    );
    const component = create();

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith('El traslado no existe');
  });
});
