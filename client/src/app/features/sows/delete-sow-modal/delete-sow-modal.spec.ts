import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { NotificationService } from '../../../core/notifications/notification.service';
import { Sow } from '../../../core/sows/sow.models';
import { SowsService } from '../../../core/sows/sows.service';
import { DeleteSowModal } from './delete-sow-modal';

function buildSow(): Sow {
  return {
    id: '42',
    code: 'C-001',
    location: 'Corral A',
    active: true,
    entry_date: '2026-01-10',
    birth_date: null,
    note: null,
    state: 'Viva',
    origin: 'Propio',
    parity: 2,
    breed_id: 'breed-1',
  };
}

class SowsStub {
  deleteSow = vi.fn(() => of(undefined));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('DeleteSowModal', () => {
  let stub: SowsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new SowsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [DeleteSowModal],
      providers: [
        { provide: SowsService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  const create = () => {
    const fixture = TestBed.createComponent(DeleteSowModal);
    fixture.componentRef.setInput('sow', buildSow());
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return fixture.componentInstance as any;
  };

  it('does nothing when there is no sow', async () => {
    const fixture = TestBed.createComponent(DeleteSowModal);
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    await (fixture.componentInstance as any).confirm();

    expect(stub.deleteSow).not.toHaveBeenCalled();
  });

  it('deletes the sow and emits it', async () => {
    const component = create();
    const deleted = vi.fn();
    component.deleted.subscribe(deleted);

    await component.confirm();

    expect(stub.deleteSow).toHaveBeenCalledWith('42');
    expect(deleted).toHaveBeenCalledWith(buildSow());
  });

  it('shows the server message when the sow has linked records', async () => {
    stub.deleteSow = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 409,
            error: {
              error:
                'No se puede eliminar esta cerda porque tiene registros enlazados. Desactívala en su lugar.',
            },
          }),
      ),
    );
    const component = create();

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith(
      'No se puede eliminar esta cerda porque tiene registros enlazados. Desactívala en su lugar.',
    );
  });

  it('shows a fallback when the sow is missing', async () => {
    stub.deleteSow = vi.fn(() => throwError(() => new HttpErrorResponse({ status: 404 })));
    const component = create();

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith('La cerda no existe');
  });
});
