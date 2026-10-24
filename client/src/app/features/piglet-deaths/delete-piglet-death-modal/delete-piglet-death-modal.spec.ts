import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { PigletDeath } from '../../../core/piglet-deaths/piglet-death.models';
import { PigletDeathsService } from '../../../core/piglet-deaths/piglet-deaths.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { DeletePigletDeathModal } from './delete-piglet-death-modal';

function buildDeath(): PigletDeath {
  return {
    id: '42',
    farrowing_id: 'farrowing-1',
    sow_id: 'sow-1',
    sow_code: 'C-001',
    operator_id: 'operator-1',
    operator_name: 'Operador',
    death_date: '2026-04-25',
    quantity: 2,
    weight: 2.5,
    cause: 'Aplastado',
    turn: 'Mañana',
    note: null,
  };
}

class PigletDeathsStub {
  deletePigletDeath = vi.fn(() => of(void 0));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('DeletePigletDeathModal', () => {
  let stub: PigletDeathsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new PigletDeathsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [DeletePigletDeathModal],
      providers: [
        { provide: PigletDeathsService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  const create = () => {
    const fixture = TestBed.createComponent(DeletePigletDeathModal);
    fixture.componentRef.setInput('death', buildDeath());
    fixture.componentRef.setInput('sowCode', 'C-001');
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return fixture.componentInstance as any;
  };

  it('does nothing when there is no death', async () => {
    const fixture = TestBed.createComponent(DeletePigletDeathModal);
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    await (fixture.componentInstance as any).confirm();

    expect(stub.deletePigletDeath).not.toHaveBeenCalled();
  });

  it('deletes the death and emits it', async () => {
    const component = create();
    const deleted = vi.fn();
    component.deleted.subscribe(deleted);

    await component.confirm();

    expect(stub.deletePigletDeath).toHaveBeenCalledWith('42');
    expect(deleted).toHaveBeenCalledWith(buildDeath());
  });

  it('shows the server message when a later weaging blocks the deletion', async () => {
    stub.deletePigletDeath = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 409,
            error: { error: 'No se puede eliminar la muerte porque existe un destete posterior' },
          }),
      ),
    );
    const component = create();
    const deleted = vi.fn();
    component.deleted.subscribe(deleted);

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith(
      'No se puede eliminar la muerte porque existe un destete posterior',
    );
    expect(deleted).not.toHaveBeenCalled();
    expect(component.loading()).toBe(false);
  });

  it('shows a fallback when the server returns no message', async () => {
    stub.deletePigletDeath = vi.fn(() => throwError(() => new HttpErrorResponse({ status: 500 })));
    const component = create();
    const deleted = vi.fn();
    component.deleted.subscribe(deleted);

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith('No se pudo eliminar la muerte');
    expect(deleted).not.toHaveBeenCalled();
  });

  it('shows a fallback when the death is missing', async () => {
    stub.deletePigletDeath = vi.fn(() => throwError(() => new HttpErrorResponse({ status: 404 })));
    const component = create();

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith('La muerte no existe');
  });
});
