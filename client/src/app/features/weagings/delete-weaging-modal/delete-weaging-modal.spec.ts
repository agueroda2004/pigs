import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { Observable, Subject, of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { Weaging } from '../../../core/weagings/weaging.models';
import { WeagingsService } from '../../../core/weagings/weagings.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { DeleteWeagingModal } from './delete-weaging-modal';

function buildWeaging(): Weaging {
  return {
    id: '42',
    farrowing_id: 'farrowing-1',
    sow_id: 'sow-1',
    sow_code: 'C-001',
    weaging_date: '2026-04-25',
    quantity: 8,
    total_weight: 120.5,
    destination: 'Nave 2',
    note: null,
  };
}

class WeagingsStub {
  deleteWeaging: () => Observable<void> = vi.fn(() => of(void 0));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('DeleteWeagingModal', () => {
  let stub: WeagingsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new WeagingsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [DeleteWeagingModal],
      providers: [
        { provide: WeagingsService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  const create = () => {
    const fixture = TestBed.createComponent(DeleteWeagingModal);
    fixture.componentRef.setInput('weaging', buildWeaging());
    fixture.componentRef.setInput('sowCode', 'C-001');
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return fixture.componentInstance as any;
  };

  it('does nothing when there is no weaging', async () => {
    const fixture = TestBed.createComponent(DeleteWeagingModal);
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    await (fixture.componentInstance as any).confirm();

    expect(stub.deleteWeaging).not.toHaveBeenCalled();
  });

  it('deletes the weaging and emits it', async () => {
    const component = create();
    const deleted = vi.fn();
    component.deleted.subscribe(deleted);

    await component.confirm();

    expect(stub.deleteWeaging).toHaveBeenCalledWith('42');
    expect(deleted).toHaveBeenCalledWith(buildWeaging());
    expect(component.loading()).toBe(false);
  });

  it('keeps the loading state and ignores duplicate submissions', async () => {
    const pending = new Subject<void>();
    stub.deleteWeaging = vi.fn(() => pending.asObservable());
    const component = create();
    const deleted = vi.fn();
    component.deleted.subscribe(deleted);

    const first = component.confirm();
    expect(component.loading()).toBe(true);

    await component.confirm();
    expect(stub.deleteWeaging).toHaveBeenCalledTimes(1);

    pending.next();
    pending.complete();
    await first;

    expect(deleted).toHaveBeenCalledTimes(1);
    expect(component.loading()).toBe(false);
  });

  it('shows the server message when later events block the deletion', async () => {
    stub.deleteWeaging = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 409,
            error: {
              error: 'El destete no se puede eliminar porque la cerda tiene eventos posteriores',
            },
          }),
      ),
    );
    const component = create();
    const deleted = vi.fn();
    component.deleted.subscribe(deleted);

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith(
      'El destete no se puede eliminar porque la cerda tiene eventos posteriores',
    );
    expect(deleted).not.toHaveBeenCalled();
    expect(component.loading()).toBe(false);
  });

  it('shows a conflict fallback when the server returns no message', async () => {
    stub.deleteWeaging = vi.fn(() => throwError(() => new HttpErrorResponse({ status: 409 })));
    const component = create();

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith(
      'No se puede eliminar el destete porque la cerda tiene montas o bajas posteriores',
    );
  });

  it('shows a fallback when the weaging is missing', async () => {
    stub.deleteWeaging = vi.fn(() => throwError(() => new HttpErrorResponse({ status: 404 })));
    const component = create();

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith('El destete no existe');
  });

  it('shows a generic fallback for unexpected failures', async () => {
    stub.deleteWeaging = vi.fn(() => throwError(() => new HttpErrorResponse({ status: 500 })));
    const component = create();

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith('No se pudo eliminar el destete');
  });
});
