import { WritableSignal, signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { AuthService } from '../../../core/auth/auth.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { PigletFostering } from '../../../core/piglet-fosterings/piglet-fostering.models';
import { PigletFosteringsService } from '../../../core/piglet-fosterings/piglet-fosterings.service';
import { SowsService } from '../../../core/sows/sows.service';
import { PigletFosteringsPage } from './piglet-fosterings-page';

function buildFostering(id: string): PigletFostering {
  return {
    id,
    donor_farrowing_id: 'farrowing-1',
    receiver_farrowing_id: 'farrowing-2',
    donor_sow_id: 'sow-1',
    receiver_sow_id: 'sow-2',
    movement_date: '2026-04-25',
    quantity: 2,
    note: null,
    created_at: '2026-04-25T12:00:00',
    updated_at: '2026-04-25T12:00:00',
    created_by: 'admin',
    updated_by: 'admin',
  };
}

class PigletFosteringsStub {
  listPigletFosterings = vi.fn(() => of([buildFostering('1'), buildFostering('2')]));
  createPigletFostering = vi.fn(() => of(buildFostering('3')));
}

class SowsStub {
  listSows = vi.fn(() =>
    of([
      { id: 'sow-1', code: 'C-001' },
      { id: 'sow-2', code: 'C-002' },
    ]),
  );
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('PigletFosteringsPage', () => {
  let stub: PigletFosteringsStub;
  let notifications: NotificationsStub;
  let isAdmin: WritableSignal<boolean>;

  beforeEach(async () => {
    stub = new PigletFosteringsStub();
    notifications = new NotificationsStub();
    isAdmin = signal(true);
    await TestBed.configureTestingModule({
      imports: [PigletFosteringsPage],
      providers: [
        { provide: PigletFosteringsService, useValue: stub },
        { provide: SowsService, useValue: new SowsStub() },
        { provide: NotificationService, useValue: notifications },
        { provide: AuthService, useValue: { isAdmin } },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => TestBed.createComponent(PigletFosteringsPage).componentInstance as any;

  it('loads fosterings on init and renders a card per fostering', async () => {
    const fixture = TestBed.createComponent(PigletFosteringsPage);

    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();

    expect(stub.listPigletFosterings).toHaveBeenCalled();
    expect(fixture.nativeElement.querySelectorAll('app-piglet-fostering-card')).toHaveLength(2);
  });

  it('resolves both sow codes from the lookup', async () => {
    const component = create();
    await component.loadLookups();

    expect(component.donorSowCode(buildFostering('1'))).toBe('C-001');
    expect(component.receiverSowCode(buildFostering('1'))).toBe('C-002');
  });

  it('shows the error state and retries', async () => {
    stub.listPigletFosterings = vi.fn(() => throwError(() => new Error('failed')));
    const component = create();

    await component.loadFosterings();
    expect(component.error()).toBe(true);

    stub.listPigletFosterings = vi.fn(() => of([buildFostering('1')]));
    await component.loadFosterings();

    expect(component.error()).toBe(false);
    expect(component.fosterings()).toHaveLength(1);
  });

  it('reloads the list after registering a fostering', async () => {
    const component = create();
    await component.loadFosterings();
    component.openModal();

    stub.listPigletFosterings = vi.fn(() =>
      of([buildFostering('1'), buildFostering('2'), buildFostering('3')]),
    );
    await component.onCreated({ donor: 'C-001', receiver: 'C-002' });

    expect(notifications.success).toHaveBeenCalledWith(
      'Traslado de "C-001" a "C-002" registrado correctamente',
    );
    expect(component.fosterings()).toHaveLength(3);
  });

  it('hides the register button for non-admins', () => {
    isAdmin.set(false);
    const fixture = TestBed.createComponent(PigletFosteringsPage);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).not.toContain('Registrar traslado');
  });

  it('applies the donor and receiver sow filters when searching', async () => {
    const component = create();
    component.filterForm.setValue({
      donor_sow_id: 'sow-1',
      receiver_sow_id: 'sow-2',
      from: '',
      to: '',
    });

    await component.search();

    expect(stub.listPigletFosterings).toHaveBeenLastCalledWith({
      donor_sow_id: 'sow-1',
      receiver_sow_id: 'sow-2',
    });
  });

  it('applies the date range filter when searching', async () => {
    const component = create();
    component.filterForm.setValue({
      donor_sow_id: '',
      receiver_sow_id: '',
      from: '2026-04-01',
      to: '2026-04-30',
    });

    await component.search();

    expect(stub.listPigletFosterings).toHaveBeenLastCalledWith({
      from: '2026-04-01',
      to: '2026-04-30',
    });
  });

  it('rejects a range whose start is after its end', async () => {
    const component = create();
    component.filterForm.setValue({
      donor_sow_id: '',
      receiver_sow_id: '',
      from: '2026-04-30',
      to: '2026-04-01',
    });

    await component.search();

    expect(component.rangeError()).toBeTruthy();
    expect(stub.listPigletFosterings).not.toHaveBeenCalled();
  });

  it('clears the filters and reloads without them', async () => {
    const component = create();
    component.filterForm.setValue({
      donor_sow_id: 'sow-1',
      receiver_sow_id: '',
      from: '',
      to: '',
    });
    await component.search();

    stub.listPigletFosterings = vi.fn(() => of([buildFostering('1')]));
    await component.clearFilters();

    expect(stub.listPigletFosterings).toHaveBeenCalledWith({});
    expect(component.filterForm.getRawValue()).toEqual({
      donor_sow_id: '',
      receiver_sow_id: '',
      from: '',
      to: '',
    });
  });
});
