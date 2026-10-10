import { WritableSignal, signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { AuthService } from '../../../core/auth/auth.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { OperatorsService } from '../../../core/operators/operators.service';
import { PigletDeath } from '../../../core/piglet-deaths/piglet-death.models';
import { PigletDeathsService } from '../../../core/piglet-deaths/piglet-deaths.service';
import { SowsService } from '../../../core/sows/sows.service';
import { PigletDeathsPage } from './piglet-deaths-page';

function buildDeath(id: string): PigletDeath {
  return {
    id,
    farrowing_id: 'farrowing-1',
    sow_id: 'sow-1',
    operator_id: 'operator-1',
    operator_name: 'Operador',
    death_date: '2026-04-25',
    quantity: 2,
    weight: 2.5,
    cause: 'Aplastado',
    turn: 'Mañana',
    note: null,
    created_at: '2026-04-25T12:00:00',
    updated_at: '2026-04-25T12:00:00',
    created_by: 'admin',
    updated_by: 'admin',
  };
}

class PigletDeathsStub {
  listPigletDeaths = vi.fn(() => of([buildDeath('1'), buildDeath('2')]));
  createPigletDeath = vi.fn(() => of(buildDeath('3')));
}

class SowsStub {
  listSowDropdown = vi.fn(() => of([{ id: 'sow-1', code: 'C-001' }]));
}

class OperatorsStub {
  listOperators = vi.fn(() => of([]));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('PigletDeathsPage', () => {
  let stub: PigletDeathsStub;
  let notifications: NotificationsStub;
  let isAdmin: WritableSignal<boolean>;

  beforeEach(async () => {
    stub = new PigletDeathsStub();
    notifications = new NotificationsStub();
    isAdmin = signal(true);
    await TestBed.configureTestingModule({
      imports: [PigletDeathsPage],
      providers: [
        { provide: PigletDeathsService, useValue: stub },
        { provide: SowsService, useValue: new SowsStub() },
        { provide: OperatorsService, useValue: new OperatorsStub() },
        { provide: NotificationService, useValue: notifications },
        { provide: AuthService, useValue: { isAdmin } },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => TestBed.createComponent(PigletDeathsPage).componentInstance as any;

  it('loads deaths on init and renders a card per death', async () => {
    const fixture = TestBed.createComponent(PigletDeathsPage);

    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();

    expect(stub.listPigletDeaths).toHaveBeenCalled();
    expect(fixture.nativeElement.querySelectorAll('app-piglet-death-card')).toHaveLength(2);
  });

  it('resolves the sow code from the lookup', async () => {
    const component = create();
    await component.loadLookups();

    expect(component.sowCode(buildDeath('1'))).toBe('C-001');
  });

  it('shows the error state and retries', async () => {
    stub.listPigletDeaths = vi.fn(() => throwError(() => new Error('failed')));
    const component = create();

    await component.loadDeaths();
    expect(component.error()).toBe(true);

    stub.listPigletDeaths = vi.fn(() => of([buildDeath('1')]));
    await component.loadDeaths();

    expect(component.error()).toBe(false);
    expect(component.deaths()).toHaveLength(1);
  });

  it('reloads the list after registering a death', async () => {
    const component = create();
    await component.loadDeaths();
    component.openModal();

    stub.listPigletDeaths = vi.fn(() =>
      of([buildDeath('1'), buildDeath('2'), buildDeath('3')]),
    );
    await component.onCreated('C-001');

    expect(notifications.success).toHaveBeenCalledWith(
      'Muerte de la cerda "C-001" registrada correctamente',
    );
    expect(component.deaths()).toHaveLength(3);
  });

  it('hides the register button for non-admins', () => {
    isAdmin.set(false);
    const fixture = TestBed.createComponent(PigletDeathsPage);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).not.toContain('Registrar muerte');
  });

  it('applies the sow filter when searching', async () => {
    const component = create();
    component.filterForm.setValue({ sow_id: 'sow-1', from: '', to: '' });

    await component.search();

    expect(stub.listPigletDeaths).toHaveBeenLastCalledWith({ sow_id: 'sow-1' });
  });

  it('applies the date range filter when searching', async () => {
    const component = create();
    component.filterForm.setValue({ sow_id: '', from: '2026-04-01', to: '2026-04-30' });

    await component.search();

    expect(stub.listPigletDeaths).toHaveBeenLastCalledWith({
      from: '2026-04-01',
      to: '2026-04-30',
    });
  });

  it('rejects a range whose start is after its end', async () => {
    const component = create();
    component.filterForm.setValue({ sow_id: '', from: '2026-04-30', to: '2026-04-01' });

    await component.search();

    expect(component.rangeError()).toBeTruthy();
    expect(stub.listPigletDeaths).not.toHaveBeenCalled();
  });

  it('clears the filters and reloads without them', async () => {
    const component = create();
    component.filterForm.setValue({ sow_id: 'sow-1', from: '', to: '' });
    await component.search();

    stub.listPigletDeaths = vi.fn(() => of([buildDeath('1')]));
    await component.clearFilters();

    expect(stub.listPigletDeaths).toHaveBeenCalledWith({});
    expect(component.filterForm.getRawValue()).toEqual({ sow_id: '', from: '', to: '' });
  });
});
