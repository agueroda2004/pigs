import { HttpErrorResponse } from '@angular/common/http';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { NotificationService } from '../../../core/notifications/notification.service';
import { SowRemoval } from '../../../core/sow-removals/sow-removal.models';
import { SowRemovalsService } from '../../../core/sow-removals/sow-removals.service';
import { DeleteSowRemovalModal } from './delete-sow-removal-modal';

function buildRemoval(overrides: Partial<SowRemoval> = {}): SowRemoval {
  return {
    id: '1',
    sow_id: 'sow-1',
    removal_date: '2026-01-20',
    type: 'Muerte',
    reason: 'Enfermedad',
    note: null,
    last_state: 'Gestando',
    created_at: '2026-01-20T12:00:00',
    updated_at: '2026-01-20T12:00:00',
    created_by: 'admin',
    updated_by: 'admin',
    ...overrides,
  };
}

class SowRemovalsStub {
  deleteSowRemoval = vi.fn((_id: string) => of(void 0));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('DeleteSowRemovalModal', () => {
  let stub: SowRemovalsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new SowRemovalsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [DeleteSowRemovalModal],
      providers: [
        { provide: SowRemovalsService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  async function openWith(
    removal: SowRemoval | null,
  ): Promise<ComponentFixture<DeleteSowRemovalModal>> {
    const fixture = TestBed.createComponent(DeleteSowRemovalModal);
    fixture.componentRef.setInput('removal', removal);
    fixture.componentRef.setInput('sowCode', removal ? 'C-001' : null);
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

    expect(stub.deleteSowRemoval).not.toHaveBeenCalled();
  });

  it('deletes the removal and emits deleted', async () => {
    const fixture = await openWith(buildRemoval());
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;
    const deleted = vi.fn();
    component.deleted.subscribe(deleted);

    await component.confirm();

    expect(stub.deleteSowRemoval).toHaveBeenCalledWith('1');
    expect(deleted).toHaveBeenCalled();
  });

  it('shows the server error message when deletion fails', async () => {
    stub.deleteSowRemoval = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 409,
            error: { error: 'La cerda no está en el estado esperado para eliminar la baja' },
          }),
      ),
    );
    const fixture = await openWith(buildRemoval());
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith(
      'La cerda no está en el estado esperado para eliminar la baja',
    );
  });

  it('falls back to a generic conflict message', async () => {
    stub.deleteSowRemoval = vi.fn(() => throwError(() => new HttpErrorResponse({ status: 409 })));
    const fixture = await openWith(buildRemoval());
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith('La cerda ya no está en el estado de la baja');
  });
});
