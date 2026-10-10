import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { Abortion } from '../../../core/abortions/abortion.models';
import { AbortionsService } from '../../../core/abortions/abortions.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { DeleteAbortionModal } from './delete-abortion-modal';

function buildAbortion(): Abortion {
  return {
    id: '42',
    sow_id: 'sow-1',
    sow_code: 'C-001',
    service_id: 'service-1',
    abortion_date: '2026-01-15',
    cause: 'Infeccioso',
    note: null,
  };
}

class AbortionsStub {
  deleteAbortion = vi.fn(() => of(undefined));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('DeleteAbortionModal', () => {
  let stub: AbortionsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new AbortionsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [DeleteAbortionModal],
      providers: [
        { provide: AbortionsService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  const create = () => {
    const fixture = TestBed.createComponent(DeleteAbortionModal);
    fixture.componentRef.setInput('abortion', buildAbortion());
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return fixture.componentInstance as any;
  };

  it('does nothing when there is no abortion', async () => {
    const fixture = TestBed.createComponent(DeleteAbortionModal);
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    await (fixture.componentInstance as any).confirm();

    expect(stub.deleteAbortion).not.toHaveBeenCalled();
  });

  it('deletes the abortion and emits it', async () => {
    const component = create();
    const deleted = vi.fn();
    component.deleted.subscribe(deleted);

    await component.confirm();

    expect(stub.deleteAbortion).toHaveBeenCalledWith('42');
    expect(deleted).toHaveBeenCalledWith(buildAbortion());
  });

  it('shows the server message when the abortion cannot be deleted', async () => {
    stub.deleteAbortion = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 409,
            error: { error: 'El aborto tiene registros posteriores y no se puede eliminar' },
          }),
      ),
    );
    const component = create();

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith(
      'El aborto tiene registros posteriores y no se puede eliminar',
    );
  });

  it('shows a fallback when the abortion is missing', async () => {
    stub.deleteAbortion = vi.fn(() => throwError(() => new HttpErrorResponse({ status: 404 })));
    const component = create();

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith('El aborto no existe');
  });
});
