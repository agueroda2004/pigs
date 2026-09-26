import { WritableSignal, signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { AuthService } from '../../../core/auth/auth.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { SowRemoval } from '../../../core/sow-removals/sow-removal.models';
import { SowRemovalsService } from '../../../core/sow-removals/sow-removals.service';
import { SowsService } from '../../../core/sows/sows.service';
import { SowRemovalsPage } from './sow-removals-page';

function buildRemoval(id: string): SowRemoval {
  return {
    id,
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
  };
}

class SowRemovalsStub {
  listSowRemovals = vi.fn(() => of([buildRemoval('1'), buildRemoval('2')]));
  createSowRemoval = vi.fn(() => of(buildRemoval('3')));
}

class SowsStub {
  listSows = vi.fn(() => of([{ id: 'sow-1', code: 'C-001' }]));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('SowRemovalsPage', () => {
  let stub: SowRemovalsStub;
  let notifications: NotificationsStub;
  let isAdmin: WritableSignal<boolean>;

  beforeEach(async () => {
    stub = new SowRemovalsStub();
    notifications = new NotificationsStub();
    isAdmin = signal(true);
    await TestBed.configureTestingModule({
      imports: [SowRemovalsPage],
      providers: [
        { provide: SowRemovalsService, useValue: stub },
        { provide: SowsService, useValue: new SowsStub() },
        { provide: NotificationService, useValue: notifications },
        { provide: AuthService, useValue: { isAdmin } },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => TestBed.createComponent(SowRemovalsPage).componentInstance as any;

  it('loads removals on init and renders a card per removal', async () => {
    const fixture = TestBed.createComponent(SowRemovalsPage);

    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();

    expect(stub.listSowRemovals).toHaveBeenCalled();
    expect(fixture.nativeElement.querySelectorAll('app-sow-removal-card')).toHaveLength(2);
  });

  it('resolves the sow code from the lookup', async () => {
    const component = create();
    await component.loadLookups();

    expect(component.sowCode(buildRemoval('1'))).toBe('C-001');
  });

  it('shows the error state and retries', async () => {
    stub.listSowRemovals = vi.fn(() => throwError(() => new Error('failed')));
    const component = create();

    await component.loadSowRemovals();
    expect(component.error()).toBe(true);

    stub.listSowRemovals = vi.fn(() => of([buildRemoval('1')]));
    await component.loadSowRemovals();

    expect(component.error()).toBe(false);
    expect(component.removals()).toHaveLength(1);
  });

  it('reloads the list and keeps the modal open after registering a removal', async () => {
    const component = create();
    await component.loadSowRemovals();
    component.openModal();

    stub.listSowRemovals = vi.fn(() =>
      of([buildRemoval('1'), buildRemoval('2'), buildRemoval('3')]),
    );
    await component.onCreated('C-001');

    expect(notifications.success).toHaveBeenCalledWith(
      'Cerda "C-001" removida correctamente',
    );
    expect(component.modalOpen()).toBe(true);
    expect(component.removals()).toHaveLength(3);
  });

  it('hides the register button for non-admins', () => {
    isAdmin.set(false);
    const fixture = TestBed.createComponent(SowRemovalsPage);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).not.toContain('Registrar baja');
  });

  it('applies the sow filter when searching', async () => {
    const component = create();
    component.filterForm.setValue({ sow_id: 'sow-1' });

    await component.search();

    expect(stub.listSowRemovals).toHaveBeenLastCalledWith({ sow_id: 'sow-1' });
  });

  it('clears the filters and reloads without them', async () => {
    const component = create();
    component.filterForm.setValue({ sow_id: 'sow-1' });
    await component.search();

    stub.listSowRemovals = vi.fn(() => of([buildRemoval('1')]));
    await component.clearFilters();

    expect(stub.listSowRemovals).toHaveBeenCalledWith({});
    expect(component.filterForm.getRawValue()).toEqual({ sow_id: '' });
  });
});
