import { HttpErrorResponse } from '@angular/common/http';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { BoarRemoval } from '../../../core/boar-removals/boar-removal.models';
import { BoarRemovalsService } from '../../../core/boar-removals/boar-removals.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { EditBoarRemovalModal } from './edit-boar-removal-modal';

function buildRemoval(overrides: Partial<BoarRemoval> = {}): BoarRemoval {
  return {
    id: '1',
    boar_id: 'boar-1',
    removal_date: '2026-01-20',
    type: 'Muerte',
    reason: 'Enfermedad',
    note: 'Nota original',
    last_state: 'Vivo',
    ...overrides,
  };
}

class BoarRemovalsStub {
  updateBoarRemoval = vi.fn((_id: string, _request: Record<string, unknown>) => of(void 0));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('EditBoarRemovalModal', () => {
  let stub: BoarRemovalsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new BoarRemovalsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [EditBoarRemovalModal],
      providers: [
        { provide: BoarRemovalsService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  async function openWith(removal: BoarRemoval): Promise<ComponentFixture<EditBoarRemovalModal>> {
    const fixture = TestBed.createComponent(EditBoarRemovalModal);
    fixture.componentRef.setInput('removal', removal);
    fixture.componentRef.setInput('boarCode', 'B-001');
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

    expect(stub.updateBoarRemoval).toHaveBeenCalledWith('1', { reason: 'Otro' });
  });

  it('emits the local removal after a successful update', async () => {
    const removal = buildRemoval();
    const fixture = await openWith(removal);
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;
    const updated = vi.fn();
    component.updated.subscribe(updated);
    component.form.patchValue({ reason: 'Otro' });

    await component.submit();

    expect(stub.updateBoarRemoval).toHaveBeenCalledWith('1', { reason: 'Otro' });
    expect(updated).toHaveBeenCalledWith(removal);
  });

  it('clears the note with an empty string', async () => {
    const fixture = await openWith(buildRemoval());
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;
    component.form.patchValue({ note: '' });

    await component.submit();

    expect(stub.updateBoarRemoval).toHaveBeenCalledWith('1', { note: '' });
  });

  it('never sends the boar identifier', async () => {
    const fixture = await openWith(buildRemoval());
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;
    component.form.patchValue({ reason: 'Otro' });

    await component.submit();

    const request = stub.updateBoarRemoval.mock.calls[0][1] as Record<string, unknown>;
    expect(request).not.toHaveProperty('boar_id');
  });

  it('emits the current removal when nothing changed', async () => {
    const removal = buildRemoval();
    const fixture = await openWith(removal);
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;
    const updated = vi.fn();
    component.updated.subscribe(updated);

    await component.submit();

    expect(stub.updateBoarRemoval).not.toHaveBeenCalled();
    expect(updated).toHaveBeenCalledWith(removal);
  });

  it('shows the server error message when the update fails', async () => {
    stub.updateBoarRemoval = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 400,
            error: {
              error: 'La fecha de la baja no puede ser anterior a la última monta del verraco',
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
      'La fecha de la baja no puede ser anterior a la última monta del verraco',
    );
  });
});
