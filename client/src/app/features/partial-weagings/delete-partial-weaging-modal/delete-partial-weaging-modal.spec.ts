import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { PartialWeaging } from '../../../core/partial-weagings/partial-weaging.models';
import { PartialWeagingsService } from '../../../core/partial-weagings/partial-weagings.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { DeletePartialWeagingModal } from './delete-partial-weaging-modal';

function buildWeaging(): PartialWeaging {
  return {
    id: '42',
    farrowing_id: 'farrowing-1',
    sow_id: 'sow-1',
    sow_code: 'C-001',
    weaging_date: '2026-04-25',
    quantity: 2,
    total_weight: 4.5,
    type: 'Normal',
    note: null,
  };
}

class PartialWeagingsStub {
  deletePartialWeaging = vi.fn(() => of(void 0));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('DeletePartialWeagingModal', () => {
  let stub: PartialWeagingsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new PartialWeagingsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [DeletePartialWeagingModal],
      providers: [
        { provide: PartialWeagingsService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  const create = () => {
    const fixture = TestBed.createComponent(DeletePartialWeagingModal);
    fixture.componentRef.setInput('weaging', buildWeaging());
    fixture.componentRef.setInput('sowCode', 'C-001');
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return fixture.componentInstance as any;
  };

  it('does nothing when there is no weaging', async () => {
    const fixture = TestBed.createComponent(DeletePartialWeagingModal);
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    await (fixture.componentInstance as any).confirm();

    expect(stub.deletePartialWeaging).not.toHaveBeenCalled();
  });

  it('deletes the weaging and emits it', async () => {
    const component = create();
    const deleted = vi.fn();
    component.deleted.subscribe(deleted);

    await component.confirm();

    expect(stub.deletePartialWeaging).toHaveBeenCalledWith('42');
    expect(deleted).toHaveBeenCalledWith(buildWeaging());
    expect(component.loading()).toBe(false);
  });

  it('shows the server message when later events block the deletion', async () => {
    stub.deletePartialWeaging = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 409,
            error: {
              error:
                'No se puede eliminar el destete parcial porque existen montas o bajas posteriores',
            },
          }),
      ),
    );
    const component = create();
    const deleted = vi.fn();
    component.deleted.subscribe(deleted);

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith(
      'No se puede eliminar el destete parcial porque existen montas o bajas posteriores',
    );
    expect(deleted).not.toHaveBeenCalled();
    expect(component.loading()).toBe(false);
  });

  it('shows a fallback when the server returns no message', async () => {
    stub.deletePartialWeaging = vi.fn(() =>
      throwError(() => new HttpErrorResponse({ status: 500 })),
    );
    const component = create();
    const deleted = vi.fn();
    component.deleted.subscribe(deleted);

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith('No se pudo eliminar el destete parcial');
    expect(deleted).not.toHaveBeenCalled();
  });

  it('shows a fallback when the weaging is missing', async () => {
    stub.deletePartialWeaging = vi.fn(() =>
      throwError(() => new HttpErrorResponse({ status: 404 })),
    );
    const component = create();

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith('El destete parcial no existe');
  });
});
