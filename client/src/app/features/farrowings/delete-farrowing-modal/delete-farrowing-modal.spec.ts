import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { Farrowing } from '../../../core/farrowings/farrowing.models';
import { FarrowingsService } from '../../../core/farrowings/farrowings.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { DeleteFarrowingModal } from './delete-farrowing-modal';

function buildFarrowing(): Farrowing {
  return {
    id: '42',
    sow_id: 'sow-1',
    sow_code: 'C-001',
    service_id: 'service-1',
    farrow_date: '2026-04-20',
    start_time: null,
    end_time: null,
    location: null,
    live_born: 10,
    stillborn: 1,
    mummified: 0,
    current_piglets: 10,
    litter_weight: null,
    stillborn_weight: null,
    is_manipulated: false,
    is_nurse: false,
    nurse_start_date: null,
    note: null,
    operators: [],
    medications: [],
  };
}

class FarrowingsStub {
  deleteFarrowing = vi.fn(() => of(undefined));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('DeleteFarrowingModal', () => {
  let stub: FarrowingsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new FarrowingsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [DeleteFarrowingModal],
      providers: [
        { provide: FarrowingsService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  const create = () => {
    const fixture = TestBed.createComponent(DeleteFarrowingModal);
    fixture.componentRef.setInput('farrowing', buildFarrowing());
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return fixture.componentInstance as any;
  };

  it('does nothing when there is no farrowing', async () => {
    const fixture = TestBed.createComponent(DeleteFarrowingModal);
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    await (fixture.componentInstance as any).confirm();

    expect(stub.deleteFarrowing).not.toHaveBeenCalled();
  });

  it('deletes the farrowing and emits it', async () => {
    const component = create();
    const deleted = vi.fn();
    component.deleted.subscribe(deleted);

    await component.confirm();

    expect(stub.deleteFarrowing).toHaveBeenCalledWith('42');
    expect(deleted).toHaveBeenCalledWith(buildFarrowing());
  });

  it('shows the server message when the farrowing has linked records', async () => {
    stub.deleteFarrowing = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 409,
            error: { error: 'Este parto tiene registros enlazados' },
          }),
      ),
    );
    const component = create();
    const deleted = vi.fn();
    component.deleted.subscribe(deleted);

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith('Este parto tiene registros enlazados');
    expect(deleted).not.toHaveBeenCalled();
    expect(component.loading()).toBe(false);
  });

  it('shows a fallback when the server returns no message', async () => {
    stub.deleteFarrowing = vi.fn(() => throwError(() => new HttpErrorResponse({ status: 500 })));
    const component = create();
    const deleted = vi.fn();
    component.deleted.subscribe(deleted);

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith('No se pudo eliminar el parto');
    expect(deleted).not.toHaveBeenCalled();
  });

  it('shows a fallback when the farrowing is missing', async () => {
    stub.deleteFarrowing = vi.fn(() => throwError(() => new HttpErrorResponse({ status: 404 })));
    const component = create();

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith('El parto no existe');
  });
});
