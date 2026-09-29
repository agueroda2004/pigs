import { HttpErrorResponse } from '@angular/common/http';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { BoarRemoval } from '../../../core/boar-removals/boar-removal.models';
import { BoarRemovalsService } from '../../../core/boar-removals/boar-removals.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { DeleteBoarRemovalModal } from './delete-boar-removal-modal';

function buildRemoval(overrides: Partial<BoarRemoval> = {}): BoarRemoval {
  return {
    id: '1',
    boar_id: 'boar-1',
    removal_date: '2026-01-20',
    type: 'Muerte',
    reason: 'Enfermedad',
    note: null,
    last_state: 'Vivo',
    created_at: '2026-01-20T12:00:00',
    updated_at: '2026-01-20T12:00:00',
    created_by: 'admin',
    updated_by: 'admin',
    ...overrides,
  };
}

class BoarRemovalsStub {
  deleteBoarRemoval = vi.fn((_id: string) => of(void 0));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('DeleteBoarRemovalModal', () => {
  let stub: BoarRemovalsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new BoarRemovalsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [DeleteBoarRemovalModal],
      providers: [
        { provide: BoarRemovalsService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  async function openWith(
    removal: BoarRemoval | null,
  ): Promise<ComponentFixture<DeleteBoarRemovalModal>> {
    const fixture = TestBed.createComponent(DeleteBoarRemovalModal);
    fixture.componentRef.setInput('removal', removal);
    fixture.componentRef.setInput('boarCode', removal ? 'B-001' : null);
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();
    return fixture;
  }

  it('does not delete when there is no removal', async () => {
    const fixture = await openWith(null);
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;

    await component.confirm();

    expect(stub.deleteBoarRemoval).not.toHaveBeenCalled();
  });

  it('deletes the removal and emits deleted', async () => {
    const fixture = await openWith(buildRemoval());
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;
    const deleted = vi.fn();
    component.deleted.subscribe(deleted);

    await component.confirm();

    expect(stub.deleteBoarRemoval).toHaveBeenCalledWith('1');
    expect(deleted).toHaveBeenCalled();
  });

  it('shows the server error message when deletion fails', async () => {
    stub.deleteBoarRemoval = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 409,
            error: { error: 'El verraco no está en el estado esperado para eliminar la baja' },
          }),
      ),
    );
    const fixture = await openWith(buildRemoval());
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith(
      'El verraco no está en el estado esperado para eliminar la baja',
    );
  });

  it('falls back to a generic conflict message', async () => {
    stub.deleteBoarRemoval = vi.fn(() => throwError(() => new HttpErrorResponse({ status: 409 })));
    const fixture = await openWith(buildRemoval());
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith(
      'El verraco ya no está en el estado de la baja',
    );
  });
});
