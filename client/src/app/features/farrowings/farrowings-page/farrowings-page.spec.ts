import { WritableSignal, signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { AuthService } from '../../../core/auth/auth.service';
import { Farrowing } from '../../../core/farrowings/farrowing.models';
import { FarrowingsService } from '../../../core/farrowings/farrowings.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { SowsService } from '../../../core/sows/sows.service';
import { FarrowingsPage } from './farrowings-page';

function buildFarrowing(id: string): Farrowing {
  return {
    id,
    sow_id: 'sow-1',
    service_id: 'service-1',
    farrow_date: '2026-04-20',
    start_time: null,
    end_time: null,
    location: null,
    live_born: 10,
    stillborn: 1,
    mummified: 0,
    litter_weight: null,
    stillborn_weight: null,
    is_manipulated: false,
    note: null,
    operators: [],
    medications: [],
    created_at: '2026-04-20T12:00:00',
    updated_at: '2026-04-20T12:00:00',
    created_by: 'admin',
    updated_by: 'admin',
  };
}

class FarrowingsStub {
  listFarrowings = vi.fn(() => of([buildFarrowing('1'), buildFarrowing('2')]));
  createFarrowing = vi.fn(() => of(buildFarrowing('3')));
}

class SowsStub {
  listSows = vi.fn(() => of([{ id: 'sow-1', code: 'C-001' }]));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('FarrowingsPage', () => {
  let stub: FarrowingsStub;
  let notifications: NotificationsStub;
  let isAdmin: WritableSignal<boolean>;

  beforeEach(async () => {
    stub = new FarrowingsStub();
    notifications = new NotificationsStub();
    isAdmin = signal(true);
    await TestBed.configureTestingModule({
      imports: [FarrowingsPage],
      providers: [
        { provide: FarrowingsService, useValue: stub },
        { provide: SowsService, useValue: new SowsStub() },
        { provide: NotificationService, useValue: notifications },
        { provide: AuthService, useValue: { isAdmin } },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => TestBed.createComponent(FarrowingsPage).componentInstance as any;

  it('loads farrowings on init and renders a card per farrowing', async () => {
    const fixture = TestBed.createComponent(FarrowingsPage);

    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();

    expect(stub.listFarrowings).toHaveBeenCalled();
    expect(fixture.nativeElement.querySelectorAll('app-farrowing-card')).toHaveLength(2);
  });

  it('resolves the sow code from the lookup', async () => {
    const component = create();
    await component.loadLookups();

    expect(component.sowCode(buildFarrowing('1'))).toBe('C-001');
  });

  it('shows the error state and retries', async () => {
    stub.listFarrowings = vi.fn(() => throwError(() => new Error('failed')));
    const component = create();

    await component.loadFarrowings();
    expect(component.error()).toBe(true);

    stub.listFarrowings = vi.fn(() => of([buildFarrowing('1')]));
    await component.loadFarrowings();

    expect(component.error()).toBe(false);
    expect(component.farrowings()).toHaveLength(1);
  });

  it('reloads the list after registering a farrowing', async () => {
    const component = create();
    await component.loadFarrowings();
    component.openModal();

    stub.listFarrowings = vi.fn(() =>
      of([buildFarrowing('1'), buildFarrowing('2'), buildFarrowing('3')]),
    );
    await component.onCreated('C-001');

    expect(notifications.success).toHaveBeenCalledWith(
      'Parto de la cerda "C-001" registrado correctamente',
    );
    expect(component.farrowings()).toHaveLength(3);
  });

  it('hides the register button for non-admins', () => {
    isAdmin.set(false);
    const fixture = TestBed.createComponent(FarrowingsPage);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).not.toContain('Registrar parto');
  });

  it('applies the sow filter when searching', async () => {
    const component = create();
    component.filterForm.setValue({ sow_id: 'sow-1', from: '', to: '' });

    await component.search();

    expect(stub.listFarrowings).toHaveBeenLastCalledWith({ sow_id: 'sow-1' });
  });

  it('applies the date range filter when searching', async () => {
    const component = create();
    component.filterForm.setValue({ sow_id: '', from: '2026-04-01', to: '2026-04-30' });

    await component.search();

    expect(stub.listFarrowings).toHaveBeenLastCalledWith({
      from: '2026-04-01',
      to: '2026-04-30',
    });
  });

  it('rejects a range whose start is after its end', async () => {
    const component = create();
    component.filterForm.setValue({ sow_id: '', from: '2026-04-30', to: '2026-04-01' });

    await component.search();

    expect(component.rangeError()).toBeTruthy();
    expect(stub.listFarrowings).not.toHaveBeenCalled();
  });

  it('clears the filters and reloads without them', async () => {
    const component = create();
    component.filterForm.setValue({ sow_id: 'sow-1', from: '', to: '' });
    await component.search();

    stub.listFarrowings = vi.fn(() => of([buildFarrowing('1')]));
    await component.clearFilters();

    expect(stub.listFarrowings).toHaveBeenCalledWith({});
    expect(component.filterForm.getRawValue()).toEqual({ sow_id: '', from: '', to: '' });
  });
});
