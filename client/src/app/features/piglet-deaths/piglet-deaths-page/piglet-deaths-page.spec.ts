import { WritableSignal, signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { AuthService } from '../../../core/auth/auth.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { OperatorsService } from '../../../core/operators/operators.service';
import { PigletDeath, PigletDeathPage } from '../../../core/piglet-deaths/piglet-death.models';
import { PigletDeathsService } from '../../../core/piglet-deaths/piglet-deaths.service';
import { PigletDeathsPage } from './piglet-deaths-page';

function buildDeath(id: string): PigletDeath {
  return {
    id,
    farrowing_id: 'farrowing-1',
    sow_id: 'sow-1',
    sow_code: 'C-001',
    operator_id: 'operator-1',
    operator_name: 'Operador',
    death_date: '2026-04-25',
    quantity: 2,
    weight: 2.5,
    cause: 'Aplastado',
    turn: 'Mañana',
    note: null,
  };
}

function pageOf(items: PigletDeath[], total = items.length, page = 1): PigletDeathPage {
  return {
    items,
    total,
    page,
    page_size: 10,
    total_pages: Math.ceil(total / 10),
  };
}

class PigletDeathsStub {
  listPigletDeaths = vi.fn((_filters = {}, _page = 1) =>
    of(pageOf([buildDeath('1'), buildDeath('2')])),
  );
  createPigletDeath = vi.fn(() => of(void 0));
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

    expect(stub.listPigletDeaths).toHaveBeenCalledWith({}, 1);
    expect(fixture.nativeElement.querySelectorAll('app-piglet-death-card')).toHaveLength(2);
  });

  it('shows the error state and retries', async () => {
    stub.listPigletDeaths = vi.fn(() => throwError(() => new Error('failed')));
    const component = create();

    await component.loadDeaths();
    expect(component.error()).toBe(true);

    stub.listPigletDeaths = vi.fn(() => of(pageOf([buildDeath('1')])));
    await component.loadDeaths();

    expect(component.error()).toBe(false);
    expect(component.deaths()).toHaveLength(1);
  });

  it('reloads the first page after registering a death', async () => {
    const component = create();
    await component.loadDeaths();
    component.openModal();

    stub.listPigletDeaths = vi.fn(() =>
      of(pageOf([buildDeath('1'), buildDeath('2'), buildDeath('3')])),
    );
    await component.onCreated('C-001');

    expect(notifications.success).toHaveBeenCalledWith(
      'Muerte de la cerda "C-001" registrada correctamente',
    );
    expect(component.page()).toBe(1);
    expect(component.deaths()).toHaveLength(3);
  });

  it('reloads the list after deleting a death', async () => {
    const component = create();
    await component.loadDeaths();
    component.openDelete(buildDeath('1'));

    expect(component.deleteOpen()).toBe(true);

    stub.listPigletDeaths = vi.fn(() => of(pageOf([buildDeath('2')])));
    await component.onDeleted(buildDeath('1'));

    expect(component.deleteOpen()).toBe(false);
    expect(notifications.success).toHaveBeenCalledWith(
      'Muerte de la cerda "C-001" eliminada correctamente',
    );
    expect(component.deaths()).toHaveLength(1);
  });

  it('reloads the list after updating a death', async () => {
    const component = create();
    await component.loadDeaths();
    component.openEdit(buildDeath('1'));

    expect(component.editOpen()).toBe(true);

    stub.listPigletDeaths = vi.fn(() => of(pageOf([buildDeath('1')])));
    await component.onUpdated();

    expect(component.editOpen()).toBe(false);
    expect(notifications.success).toHaveBeenCalledWith('Muerte actualizada correctamente');
    expect(stub.listPigletDeaths).toHaveBeenCalled();
  });

  it('hides the card actions for non-admins', () => {
    isAdmin.set(false);
    const fixture = TestBed.createComponent(PigletDeathsPage);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).not.toContain('Eliminar');
    expect(fixture.nativeElement.textContent).not.toContain('Editar');
  });

  it('applies the sow code filter when searching and returns to the first page', async () => {
    const component = create();
    await component.nextPage();
    component.filterForm.setValue({ sow_code: '  C-001  ', from: '', to: '' });

    await component.search();

    expect(stub.listPigletDeaths).toHaveBeenLastCalledWith({ sow_code: 'C-001' }, 1);
  });

  it('applies the date range filter when searching', async () => {
    const component = create();
    component.filterForm.setValue({ sow_code: '', from: '2026-04-01', to: '2026-04-30' });

    await component.search();

    expect(stub.listPigletDeaths).toHaveBeenLastCalledWith(
      { from: '2026-04-01', to: '2026-04-30' },
      1,
    );
  });

  it('rejects a range whose start is after its end', async () => {
    const component = create();
    component.filterForm.setValue({ sow_code: '', from: '2026-04-30', to: '2026-04-01' });

    await component.search();

    expect(component.rangeError()).toBeTruthy();
    expect(stub.listPigletDeaths).not.toHaveBeenCalled();
  });

  it('clears the filters and reloads without them', async () => {
    const component = create();
    component.filterForm.setValue({ sow_code: 'C-001', from: '', to: '' });
    await component.search();

    stub.listPigletDeaths = vi.fn(() => of(pageOf([buildDeath('1')])));
    await component.clearFilters();

    expect(stub.listPigletDeaths).toHaveBeenCalledWith({}, 1);
    expect(component.filterForm.getRawValue()).toEqual({ sow_code: '', from: '', to: '' });
  });

  it('moves between pages', async () => {
    stub.listPigletDeaths = vi.fn((_filters = {}, page = 1) =>
      of(pageOf([buildDeath('1')], 15, page)),
    );
    const component = create();
    await component.loadDeaths();

    expect(component.totalPages()).toBe(2);
    expect(component.total()).toBe(15);

    await component.nextPage();
    expect(stub.listPigletDeaths).toHaveBeenLastCalledWith({}, 2);

    await component.previousPage();
    expect(stub.listPigletDeaths).toHaveBeenLastCalledWith({}, 1);
  });
});
