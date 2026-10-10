import { WritableSignal, signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { AuthService } from '../../../core/auth/auth.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { SowsService } from '../../../core/sows/sows.service';
import { Weaging } from '../../../core/weagings/weaging.models';
import { WeagingsService } from '../../../core/weagings/weagings.service';
import { WeagingsPage } from './weagings-page';

function buildWeaging(id: string): Weaging {
  return {
    id,
    farrowing_id: 'farrowing-1',
    sow_id: 'sow-1',
    weaging_date: '2026-04-25',
    quantity: 8,
    total_weight: 120.5,
    destination: 'Nave 2',
    note: null,
    created_at: '2026-04-25T12:00:00',
    updated_at: '2026-04-25T12:00:00',
    created_by: 'admin',
    updated_by: 'admin',
  };
}

class WeagingsStub {
  listWeagings = vi.fn(() => of([buildWeaging('1'), buildWeaging('2')]));
  createWeaging = vi.fn(() => of(buildWeaging('3')));
}

class SowsStub {
  listSowDropdown = vi.fn(() => of([{ id: 'sow-1', code: 'C-001' }]));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('WeagingsPage', () => {
  let stub: WeagingsStub;
  let notifications: NotificationsStub;
  let isAdmin: WritableSignal<boolean>;

  beforeEach(async () => {
    stub = new WeagingsStub();
    notifications = new NotificationsStub();
    isAdmin = signal(true);
    await TestBed.configureTestingModule({
      imports: [WeagingsPage],
      providers: [
        { provide: WeagingsService, useValue: stub },
        { provide: SowsService, useValue: new SowsStub() },
        { provide: NotificationService, useValue: notifications },
        { provide: AuthService, useValue: { isAdmin } },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => TestBed.createComponent(WeagingsPage).componentInstance as any;

  it('loads weagings on init and renders a card per weaging', async () => {
    const fixture = TestBed.createComponent(WeagingsPage);

    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();

    expect(stub.listWeagings).toHaveBeenCalled();
    expect(fixture.nativeElement.querySelectorAll('app-weaging-card')).toHaveLength(2);
  });

  it('resolves the sow code from the lookup', async () => {
    const component = create();
    await component.loadLookups();

    expect(component.sowCode(buildWeaging('1'))).toBe('C-001');
  });

  it('shows the error state and retries', async () => {
    stub.listWeagings = vi.fn(() => throwError(() => new Error('failed')));
    const component = create();

    await component.loadWeagings();
    expect(component.error()).toBe(true);

    stub.listWeagings = vi.fn(() => of([buildWeaging('1')]));
    await component.loadWeagings();

    expect(component.error()).toBe(false);
    expect(component.weagings()).toHaveLength(1);
  });

  it('reloads the list after registering a weaging', async () => {
    const component = create();
    await component.loadWeagings();

    stub.listWeagings = vi.fn(() => of([buildWeaging('1'), buildWeaging('2'), buildWeaging('3')]));
    await component.onCreated('C-001');

    expect(notifications.success).toHaveBeenCalledWith(
      'Destete de "C-001" registrado correctamente',
    );
    expect(component.weagings()).toHaveLength(3);
  });

  it('hides the register button for non-admins', () => {
    isAdmin.set(false);
    const fixture = TestBed.createComponent(WeagingsPage);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).not.toContain('Registrar destete');
  });

  it('applies the sow filter when searching', async () => {
    const component = create();
    component.filterForm.setValue({ sow_id: 'sow-1', from: '', to: '' });

    await component.search();

    expect(stub.listWeagings).toHaveBeenLastCalledWith({ sow_id: 'sow-1' });
  });

  it('applies the date range filter when searching', async () => {
    const component = create();
    component.filterForm.setValue({ sow_id: '', from: '2026-04-01', to: '2026-04-30' });

    await component.search();

    expect(stub.listWeagings).toHaveBeenLastCalledWith({
      from: '2026-04-01',
      to: '2026-04-30',
    });
  });

  it('rejects a range whose start is after its end', async () => {
    const component = create();
    component.filterForm.setValue({ sow_id: '', from: '2026-04-30', to: '2026-04-01' });

    await component.search();

    expect(component.rangeError()).toBeTruthy();
    expect(stub.listWeagings).not.toHaveBeenCalled();
  });

  it('clears the filters and reloads without them', async () => {
    const component = create();
    component.filterForm.setValue({ sow_id: 'sow-1', from: '', to: '' });
    await component.search();

    stub.listWeagings = vi.fn(() => of([buildWeaging('1')]));
    await component.clearFilters();

    expect(stub.listWeagings).toHaveBeenCalledWith({});
    expect(component.filterForm.getRawValue()).toEqual({ sow_id: '', from: '', to: '' });
  });
});
