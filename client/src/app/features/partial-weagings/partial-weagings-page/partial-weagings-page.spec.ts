import { WritableSignal, signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { AuthService } from '../../../core/auth/auth.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { PartialWeaging } from '../../../core/partial-weagings/partial-weaging.models';
import { PartialWeagingsService } from '../../../core/partial-weagings/partial-weagings.service';
import { SowsService } from '../../../core/sows/sows.service';
import { PartialWeagingsPage } from './partial-weagings-page';

function buildWeaging(id: string): PartialWeaging {
  return {
    id,
    farrowing_id: 'farrowing-1',
    sow_id: 'sow-1',
    weaging_date: '2026-04-25',
    quantity: 2,
    total_weight: 42.5,
    type: 'Normal',
    note: null,
    created_at: '2026-04-25T12:00:00',
    updated_at: '2026-04-25T12:00:00',
    created_by: 'admin',
    updated_by: 'admin',
  };
}

class PartialWeagingsStub {
  listPartialWeagings = vi.fn(() => of([buildWeaging('1'), buildWeaging('2')]));
  createPartialWeaging = vi.fn(() => of(buildWeaging('3')));
}

class SowsStub {
  listSows = vi.fn(() => of([{ id: 'sow-1', code: 'C-001' }]));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('PartialWeagingsPage', () => {
  let stub: PartialWeagingsStub;
  let notifications: NotificationsStub;
  let isAdmin: WritableSignal<boolean>;

  beforeEach(async () => {
    stub = new PartialWeagingsStub();
    notifications = new NotificationsStub();
    isAdmin = signal(true);
    await TestBed.configureTestingModule({
      imports: [PartialWeagingsPage],
      providers: [
        { provide: PartialWeagingsService, useValue: stub },
        { provide: SowsService, useValue: new SowsStub() },
        { provide: NotificationService, useValue: notifications },
        { provide: AuthService, useValue: { isAdmin } },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => TestBed.createComponent(PartialWeagingsPage).componentInstance as any;

  it('loads partial weagings on init and renders a card per weaging', async () => {
    const fixture = TestBed.createComponent(PartialWeagingsPage);

    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();

    expect(stub.listPartialWeagings).toHaveBeenCalled();
    expect(fixture.nativeElement.querySelectorAll('app-partial-weaging-card')).toHaveLength(2);
  });

  it('resolves the sow code from the lookup', async () => {
    const component = create();
    await component.loadLookups();

    expect(component.sowCode(buildWeaging('1'))).toBe('C-001');
  });

  it('shows the error state and retries', async () => {
    stub.listPartialWeagings = vi.fn(() => throwError(() => new Error('failed')));
    const component = create();

    await component.loadWeagings();
    expect(component.error()).toBe(true);

    stub.listPartialWeagings = vi.fn(() => of([buildWeaging('1')]));
    await component.loadWeagings();

    expect(component.error()).toBe(false);
    expect(component.weagings()).toHaveLength(1);
  });

  it('reloads the list after registering a weaging', async () => {
    const component = create();
    await component.loadWeagings();

    stub.listPartialWeagings = vi.fn(() =>
      of([buildWeaging('1'), buildWeaging('2'), buildWeaging('3')]),
    );
    await component.onCreated('C-001');

    expect(notifications.success).toHaveBeenCalledWith(
      'Destete parcial de "C-001" registrado correctamente',
    );
    expect(component.weagings()).toHaveLength(3);
  });

  it('hides the register button for non-admins', () => {
    isAdmin.set(false);
    const fixture = TestBed.createComponent(PartialWeagingsPage);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).not.toContain('Registrar destete');
  });

  it('applies the sow filter when searching', async () => {
    const component = create();
    component.filterForm.setValue({ sow_id: 'sow-1', from: '', to: '' });

    await component.search();

    expect(stub.listPartialWeagings).toHaveBeenLastCalledWith({ sow_id: 'sow-1' });
  });

  it('applies the date range filter when searching', async () => {
    const component = create();
    component.filterForm.setValue({ sow_id: '', from: '2026-04-01', to: '2026-04-30' });

    await component.search();

    expect(stub.listPartialWeagings).toHaveBeenLastCalledWith({
      from: '2026-04-01',
      to: '2026-04-30',
    });
  });

  it('rejects a range whose start is after its end', async () => {
    const component = create();
    component.filterForm.setValue({ sow_id: '', from: '2026-04-30', to: '2026-04-01' });

    await component.search();

    expect(component.rangeError()).toBeTruthy();
    expect(stub.listPartialWeagings).not.toHaveBeenCalled();
  });

  it('clears the filters and reloads without them', async () => {
    const component = create();
    component.filterForm.setValue({ sow_id: 'sow-1', from: '', to: '' });
    await component.search();

    stub.listPartialWeagings = vi.fn(() => of([buildWeaging('1')]));
    await component.clearFilters();

    expect(stub.listPartialWeagings).toHaveBeenCalledWith({});
    expect(component.filterForm.getRawValue()).toEqual({ sow_id: '', from: '', to: '' });
  });
});
