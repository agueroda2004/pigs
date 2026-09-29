import { HttpErrorResponse } from '@angular/common/http';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { NotificationService } from '../../../core/notifications/notification.service';
import { SowRemoval } from '../../../core/sow-removals/sow-removal.models';
import { SowRemovalsService } from '../../../core/sow-removals/sow-removals.service';
import { EditSowRemovalModal } from './edit-sow-removal-modal';

function buildRemoval(overrides: Partial<SowRemoval> = {}): SowRemoval {
  return {
    id: '1',
    sow_id: 'sow-1',
    removal_date: '2026-01-20',
    type: 'Muerte',
    reason: 'Enfermedad',
    note: 'Nota original',
    last_state: 'Gestando',
    created_at: '2026-01-20T12:00:00',
    updated_at: '2026-01-20T12:00:00',
    created_by: 'admin',
    updated_by: 'admin',
    ...overrides,
  };
}

class SowRemovalsStub {
  updateSowRemoval = vi.fn((_id: string, _request: Record<string, unknown>) => of(buildRemoval()));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('EditSowRemovalModal', () => {
  let stub: SowRemovalsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new SowRemovalsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [EditSowRemovalModal],
      providers: [
        { provide: SowRemovalsService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  async function openWith(removal: SowRemoval): Promise<ComponentFixture<EditSowRemovalModal>> {
    const fixture = TestBed.createComponent(EditSowRemovalModal);
    fixture.componentRef.setInput('removal', removal);
    fixture.componentRef.setInput('sowCode', 'C-001');
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();
    return fixture;
  }

  it('prefills the form from the removal', async () => {
    const fixture = await openWith(buildRemoval());
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;

    expect(component.form.getRawValue()).toEqual({
      removal_date: '2026-01-20',
      type: 'Muerte',
      reason: 'Enfermedad',
      note: 'Nota original',
    });
  });

  it('sends only the changed fields', async () => {
    const fixture = await openWith(buildRemoval());
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;
    component.form.patchValue({ reason: 'Otro' });

    await component.submit();

    expect(stub.updateSowRemoval).toHaveBeenCalledWith('1', { reason: 'Otro' });
  });

  it('clears the note with an empty string', async () => {
    const fixture = await openWith(buildRemoval());
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;
    component.form.patchValue({ note: '' });

    await component.submit();

    expect(stub.updateSowRemoval).toHaveBeenCalledWith('1', { note: '' });
  });

  it('never sends the sow identifier', async () => {
    const fixture = await openWith(buildRemoval());
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;
    component.form.patchValue({ reason: 'Otro' });

    await component.submit();

    const request = stub.updateSowRemoval.mock.calls[0][1] as Record<string, unknown>;
    expect(request).not.toHaveProperty('sow_id');
  });

  it('emits the current removal when nothing changed', async () => {
    const removal = buildRemoval();
    const fixture = await openWith(removal);
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;
    const updated = vi.fn();
    component.updated.subscribe(updated);

    await component.submit();

    expect(stub.updateSowRemoval).not.toHaveBeenCalled();
    expect(updated).toHaveBeenCalledWith(removal);
  });

  it('shows the server error message when the update fails', async () => {
    stub.updateSowRemoval = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 400,
            error: {
              error: 'La fecha de la baja no puede ser anterior a la fecha de ingreso de la cerda',
            },
          }),
      ),
    );
    const fixture = await openWith(buildRemoval());
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;
    component.form.patchValue({ reason: 'Otro' });

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith(
      'La fecha de la baja no puede ser anterior a la fecha de ingreso de la cerda',
    );
  });
});
