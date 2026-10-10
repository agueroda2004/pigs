import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { Abortion } from '../../../core/abortions/abortion.models';
import { AbortionsService } from '../../../core/abortions/abortions.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { EditAbortionModal } from './edit-abortion-modal';

function buildAbortion(overrides: Partial<Abortion> = {}): Abortion {
  return {
    id: '42',
    sow_id: 'sow-1',
    sow_code: 'C-001',
    service_id: 'service-1',
    abortion_date: '2026-01-15',
    cause: 'Infeccioso',
    note: null,
    ...overrides,
  };
}

class AbortionsStub {
  updateAbortion = vi.fn((_id: string, _request: unknown) => of(undefined));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('EditAbortionModal', () => {
  let stub: AbortionsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new AbortionsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [EditAbortionModal],
      providers: [
        { provide: AbortionsService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  const create = async (abortion: Abortion = buildAbortion()) => {
    const fixture = TestBed.createComponent(EditAbortionModal);
    fixture.componentRef.setInput('abortion', abortion);
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();
    await fixture.whenStable();
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return fixture.componentInstance as any;
  };

  it('prefills the form from the abortion', async () => {
    const component = await create(buildAbortion({ note: 'nota' }));

    expect(component.form.getRawValue()).toEqual({
      abortion_date: '2026-01-15',
      cause: 'Infeccioso',
      note: 'nota',
    });
  });

  it('updates the abortion and emits updated', async () => {
    const component = await create();
    const updated = vi.fn();
    component.updated.subscribe(updated);
    component.form.patchValue({ abortion_date: '2026-01-20', cause: 'Traumatismo', note: '' });

    await component.submit();

    expect(stub.updateAbortion).toHaveBeenCalledWith('42', {
      abortion_date: '2026-01-20',
      cause: 'Traumatismo',
      note: '',
    });
    expect(updated).toHaveBeenCalled();
  });

  it('does not submit when the form is invalid', async () => {
    const component = await create();
    component.form.patchValue({ cause: '' });

    await component.submit();

    expect(stub.updateAbortion).not.toHaveBeenCalled();
  });

  it('shows the server message when the update fails', async () => {
    stub.updateAbortion = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 400,
            error: { error: 'La fecha del aborto debe ser posterior a la última monta' },
          }),
      ),
    );
    const component = await create();

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith(
      'La fecha del aborto debe ser posterior a la última monta',
    );
  });

  it('shows a fallback when the abortion is missing', async () => {
    stub.updateAbortion = vi.fn(() => throwError(() => new HttpErrorResponse({ status: 404 })));
    const component = await create();

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith('El aborto no existe');
  });
});
