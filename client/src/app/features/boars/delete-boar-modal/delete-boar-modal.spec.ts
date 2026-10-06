import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { Boar } from '../../../core/boars/boar.models';
import { BoarsService } from '../../../core/boars/boars.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { DeleteBoarModal } from './delete-boar-modal';

function buildBoar(): Boar {
  return {
    id: '42',
    code: 'B-001',
    location: 'Corral A',
    active: true,
    entry_date: '2026-01-10',
    birth_date: null,
    note: null,
    state: 'Vivo',
    origin: 'Propio',
    breed_id: 'breed-1',
  };
}

class BoarsStub {
  deleteBoar = vi.fn(() => of(undefined));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('DeleteBoarModal', () => {
  let stub: BoarsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new BoarsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [DeleteBoarModal],
      providers: [
        { provide: BoarsService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  const create = () => {
    const fixture = TestBed.createComponent(DeleteBoarModal);
    fixture.componentRef.setInput('boar', buildBoar());
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return fixture.componentInstance as any;
  };

  it('does nothing when there is no boar', async () => {
    const fixture = TestBed.createComponent(DeleteBoarModal);
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    await (fixture.componentInstance as any).confirm();

    expect(stub.deleteBoar).not.toHaveBeenCalled();
  });

  it('deletes the boar and emits it', async () => {
    const component = create();
    const deleted = vi.fn();
    component.deleted.subscribe(deleted);

    await component.confirm();

    expect(stub.deleteBoar).toHaveBeenCalledWith('42');
    expect(deleted).toHaveBeenCalledWith(buildBoar());
  });

  it('shows the server message when the boar has linked records', async () => {
    stub.deleteBoar = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 409,
            error: {
              error:
                'No se puede eliminar este verraco porque tiene registros enlazados. Desactívalo en su lugar.',
            },
          }),
      ),
    );
    const component = create();

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith(
      'No se puede eliminar este verraco porque tiene registros enlazados. Desactívalo en su lugar.',
    );
  });

  it('shows a fallback when the boar is missing', async () => {
    stub.deleteBoar = vi.fn(() => throwError(() => new HttpErrorResponse({ status: 404 })));
    const component = create();

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith('El verraco no existe');
  });
});
