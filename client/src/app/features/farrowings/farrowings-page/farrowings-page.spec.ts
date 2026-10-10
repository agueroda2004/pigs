import { WritableSignal, signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { Mock, vi } from 'vitest';

import { AuthService } from '../../../core/auth/auth.service';
import { Farrowing, FarrowingPage } from '../../../core/farrowings/farrowing.models';
import { FarrowingsService } from '../../../core/farrowings/farrowings.service';
import { MedicationsService } from '../../../core/medications/medications.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { OperatorsService } from '../../../core/operators/operators.service';
import { SowsService } from '../../../core/sows/sows.service';
import { FarrowingsPage } from './farrowings-page';

function buildFarrowing(id: string): Farrowing {
  return {
    id,
    sow_id: 'sow-1',
    sow_code: 'C-001',
    service_id: 'service-1',
    farrow_date: '2026-04-20',
    start_time: null,
    end_time: null,
    location: null,
    live_born: 10,
    stillborn: 1,
    mummified: 0,
    current_piglets: 10,
    litter_weight: null,
    stillborn_weight: null,
    is_manipulated: false,
    is_nurse: false,
    nurse_start_date: null,
    note: null,
    operators: [],
    medications: [],
  };
}

function buildPage(
  items: Farrowing[],
  page = 1,
  totalPages = 1,
  total = items.length,
): FarrowingPage {
  return { items, total, page, page_size: 10, total_pages: totalPages };
}

class FarrowingsStub {
  listFarrowings: Mock = vi.fn(() => of(buildPage([buildFarrowing('1'), buildFarrowing('2')])));
  createFarrowing = vi.fn(() => of(undefined));
  updateFarrowing = vi.fn(() => of(undefined));
  deleteFarrowing = vi.fn(() => of(undefined));
}

class SowsStub {
  listSowDropdown = vi.fn(() => of([]));
}

class OperatorsStub {
  listOperatorDropdown = vi.fn(() => of([]));
}

class MedicationsStub {
  listMedications = vi.fn(() => of([]));
  listMedicationOptions = vi.fn(() => of([]));
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
        { provide: OperatorsService, useValue: new OperatorsStub() },
        { provide: MedicationsService, useValue: new MedicationsStub() },
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

  it('shows the error state and retries', async () => {
    stub.listFarrowings = vi.fn(() => throwError(() => new Error('failed')));
    const component = create();

    await component.loadFarrowings();
    expect(component.error()).toBe(true);

    stub.listFarrowings = vi.fn(() => of(buildPage([buildFarrowing('1')])));
    await component.loadFarrowings();

    expect(component.error()).toBe(false);
    expect(component.farrowings()).toHaveLength(1);
  });

  it('reloads the list after registering a farrowing', async () => {
    const component = create();
    await component.loadFarrowings();
    component.openModal();

    stub.listFarrowings = vi.fn(() =>
      of(buildPage([buildFarrowing('1'), buildFarrowing('2'), buildFarrowing('3')])),
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

  it('applies the sow code filter when searching', async () => {
    const component = create();
    component.filterForm.setValue({ sow_code: '  C-001  ', from: '', to: '' });

    await component.search();

    expect(stub.listFarrowings).toHaveBeenLastCalledWith({ sow_code: 'C-001' }, 1);
  });

  it('applies the date range filter when searching', async () => {
    const component = create();
    component.filterForm.setValue({ sow_code: '', from: '2026-04-01', to: '2026-04-30' });

    await component.search();

    expect(stub.listFarrowings).toHaveBeenLastCalledWith(
      { from: '2026-04-01', to: '2026-04-30' },
      1,
    );
  });

  it('rejects a range whose start is after its end', async () => {
    const component = create();
    component.filterForm.setValue({ sow_code: '', from: '2026-04-30', to: '2026-04-01' });

    await component.search();

    expect(component.rangeError()).toBeTruthy();
    expect(stub.listFarrowings).not.toHaveBeenCalled();
  });

  it('resets to the first page when searching', async () => {
    stub.listFarrowings = vi.fn(() => of(buildPage([buildFarrowing('1')], 2, 3, 21)));
    const component = create();
    component.page.set(2);

    component.filterForm.setValue({ sow_code: 'C-001', from: '', to: '' });
    await component.search();

    expect(component.page()).toBe(1);
    expect(stub.listFarrowings).toHaveBeenLastCalledWith({ sow_code: 'C-001' }, 1);
  });

  it('clears the filters and reloads without them', async () => {
    const component = create();
    component.filterForm.setValue({ sow_code: 'C-001', from: '', to: '' });
    await component.search();

    stub.listFarrowings = vi.fn(() => of(buildPage([buildFarrowing('1')])));
    await component.clearFilters();

    expect(stub.listFarrowings).toHaveBeenCalledWith({}, 1);
    expect(component.filterForm.getRawValue()).toEqual({ sow_code: '', from: '', to: '' });
  });

  it('navigates between pages', async () => {
    stub.listFarrowings = vi.fn((_filters: unknown, page: number) =>
      of(buildPage([buildFarrowing('1')], page, 3, 21)),
    );
    const component = create();
    await component.loadFarrowings();

    await component.nextPage();
    expect(component.page()).toBe(2);
    expect(stub.listFarrowings).toHaveBeenLastCalledWith({}, 2);

    await component.previousPage();
    expect(component.page()).toBe(1);
    expect(stub.listFarrowings).toHaveBeenLastCalledWith({}, 1);
  });

  it('does not navigate past the first or last page', async () => {
    stub.listFarrowings = vi.fn((_filters: unknown, page: number) =>
      of(buildPage([buildFarrowing('1')], page, 2, 11)),
    );
    const component = create();
    await component.loadFarrowings();

    await component.previousPage();
    expect(component.page()).toBe(1);

    await component.nextPage();
    expect(component.page()).toBe(2);
    stub.listFarrowings.mockClear();

    await component.nextPage();
    expect(component.page()).toBe(2);
    expect(stub.listFarrowings).not.toHaveBeenCalled();
  });

  it('opens the edit modal for the selected farrowing', () => {
    const component = create();
    const farrowing = buildFarrowing('9');

    component.openEdit(farrowing);

    expect(component.editOpen()).toBe(true);
    expect(component.editingFarrowing()).toEqual(farrowing);
  });

  it('reloads the list after updating a farrowing', async () => {
    const component = create();
    await component.loadFarrowings();

    stub.listFarrowings = vi.fn(() => of(buildPage([buildFarrowing('9')])));
    await component.onUpdated();

    expect(notifications.success).toHaveBeenCalledWith('Parto actualizado correctamente');
    expect(component.editOpen()).toBe(false);
    expect(component.farrowings()).toHaveLength(1);
  });

  it('reloads the list after deleting a farrowing', async () => {
    const component = create();
    await component.loadFarrowings();

    stub.listFarrowings = vi.fn(() => of(buildPage([buildFarrowing('2')])));
    await component.onDeleted(buildFarrowing('9'));

    expect(notifications.success).toHaveBeenCalledWith(
      'Parto de la cerda "C-001" eliminado correctamente',
    );
    expect(component.deleteOpen()).toBe(false);
    expect(component.farrowings()).toHaveLength(1);
  });
});
