import { WritableSignal, signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { Abortion } from '../../../core/abortions/abortion.models';
import { AbortionsService } from '../../../core/abortions/abortions.service';
import { AuthService } from '../../../core/auth/auth.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { SowsService } from '../../../core/sows/sows.service';
import { AbortionsPage } from './abortions-page';

function buildAbortion(id: string): Abortion {
  return {
    id,
    sow_id: 'sow-1',
    service_id: 'service-1',
    abortion_date: '2026-01-15',
    cause: 'Infeccioso',
    note: null,
    created_at: '2026-01-15T12:00:00',
    updated_at: '2026-01-15T12:00:00',
    created_by: 'admin',
    updated_by: 'admin',
  };
}

class AbortionsStub {
  listAbortions = vi.fn(() => of([buildAbortion('1'), buildAbortion('2')]));
  createAbortion = vi.fn(() => of(buildAbortion('3')));
}

class SowsStub {
  listSows = vi.fn(() => of([{ id: 'sow-1', code: 'C-001' }]));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('AbortionsPage', () => {
  let stub: AbortionsStub;
  let notifications: NotificationsStub;
  let isAdmin: WritableSignal<boolean>;

  beforeEach(async () => {
    stub = new AbortionsStub();
    notifications = new NotificationsStub();
    isAdmin = signal(true);
    await TestBed.configureTestingModule({
      imports: [AbortionsPage],
      providers: [
        { provide: AbortionsService, useValue: stub },
        { provide: SowsService, useValue: new SowsStub() },
        { provide: NotificationService, useValue: notifications },
        { provide: AuthService, useValue: { isAdmin } },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => TestBed.createComponent(AbortionsPage).componentInstance as any;

  it('loads abortions on init and renders a card per abortion', async () => {
    const fixture = TestBed.createComponent(AbortionsPage);

    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();

    expect(stub.listAbortions).toHaveBeenCalled();
    expect(fixture.nativeElement.querySelectorAll('app-abortion-card')).toHaveLength(2);
  });

  it('resolves the sow code from the lookup', async () => {
    const component = create();
    await component.loadLookups();

    expect(component.sowCode(buildAbortion('1'))).toBe('C-001');
  });

  it('shows the error state and retries', async () => {
    stub.listAbortions = vi.fn(() => throwError(() => new Error('failed')));
    const component = create();

    await component.loadAbortions();
    expect(component.error()).toBe(true);

    stub.listAbortions = vi.fn(() => of([buildAbortion('1')]));
    await component.loadAbortions();

    expect(component.error()).toBe(false);
    expect(component.abortions()).toHaveLength(1);
  });

  it('reloads the list and keeps the modal open after registering an abortion', async () => {
    const component = create();
    await component.loadAbortions();
    component.openModal();

    stub.listAbortions = vi.fn(() =>
      of([buildAbortion('1'), buildAbortion('2'), buildAbortion('3')]),
    );
    await component.onCreated('C-001');

    expect(notifications.success).toHaveBeenCalledWith(
      'Aborto de la cerda "C-001" registrado correctamente',
    );
    expect(component.modalOpen()).toBe(true);
    expect(component.abortions()).toHaveLength(3);
  });

  it('hides the register button for non-admins', () => {
    isAdmin.set(false);
    const fixture = TestBed.createComponent(AbortionsPage);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).not.toContain('Registrar aborto');
  });

  it('applies the sow filter when searching', async () => {
    const component = create();
    component.filterForm.setValue({ sow_id: 'sow-1' });

    await component.search();

    expect(stub.listAbortions).toHaveBeenLastCalledWith({ sow_id: 'sow-1' });
  });

  it('clears the filters and reloads without them', async () => {
    const component = create();
    component.filterForm.setValue({ sow_id: 'sow-1' });
    await component.search();

    stub.listAbortions = vi.fn(() => of([buildAbortion('1')]));
    await component.clearFilters();

    expect(stub.listAbortions).toHaveBeenCalledWith({});
    expect(component.filterForm.getRawValue()).toEqual({ sow_id: '' });
  });
});
