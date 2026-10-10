import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { NotificationService } from '../../../core/notifications/notification.service';
import { Service } from '../../../core/services/service.models';
import { ServicesService } from '../../../core/services/services.service';
import { DeleteServiceModal } from './delete-service-modal';

function buildService(): Service {
  return {
    id: '42',
    sow_id: 'sow-1',
    sow_code: 'C-001',
    expected_farrowing_date: '2026-05-04',
    note: null,
    state: 'Confirmado',
    location: 'Nave 1',
    mounts: [],
  };
}

class ServicesStub {
  deleteService = vi.fn(() => of(undefined));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('DeleteServiceModal', () => {
  let stub: ServicesStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new ServicesStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [DeleteServiceModal],
      providers: [
        { provide: ServicesService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  const create = () => {
    const fixture = TestBed.createComponent(DeleteServiceModal);
    fixture.componentRef.setInput('service', buildService());
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return fixture.componentInstance as any;
  };

  it('does nothing when there is no service', async () => {
    const fixture = TestBed.createComponent(DeleteServiceModal);
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    await (fixture.componentInstance as any).confirm();

    expect(stub.deleteService).not.toHaveBeenCalled();
  });

  it('deletes the service and emits it', async () => {
    const component = create();
    const deleted = vi.fn();
    component.deleted.subscribe(deleted);

    await component.confirm();

    expect(stub.deleteService).toHaveBeenCalledWith('42');
    expect(deleted).toHaveBeenCalledWith(buildService());
  });

  it('shows the server message when the service cannot be deleted', async () => {
    stub.deleteService = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 409,
            error: { error: 'Solo se pueden eliminar servicios en estado Confirmado' },
          }),
      ),
    );
    const component = create();

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith(
      'Solo se pueden eliminar servicios en estado Confirmado',
    );
  });

  it('shows a fallback when the service is missing', async () => {
    stub.deleteService = vi.fn(() => throwError(() => new HttpErrorResponse({ status: 404 })));
    const component = create();

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith('El servicio no existe');
  });
});
