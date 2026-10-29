import { WritableSignal, signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { AuthService } from '../../../core/auth/auth.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import {
  PigletFostering,
  PigletFosteringPage,
} from '../../../core/piglet-fosterings/piglet-fostering.models';
import { PigletFosteringsService } from '../../../core/piglet-fosterings/piglet-fosterings.service';
import { PigletFosteringsPage } from './piglet-fosterings-page';

function buildFostering(id: string, donorCode = 'C-001', receiverCode = 'C-002'): PigletFostering {
  return {
    id,
    donor_farrowing_id: 'farrowing-1',
    receiver_farrowing_id: 'farrowing-2',
    donor_sow_id: 'sow-1',
    donor_sow_code: donorCode,
    receiver_sow_id: 'sow-2',
    receiver_sow_code: receiverCode,
    movement_date: '2026-04-25',
    quantity: 2,
    note: null,
  };
}

function pageOf(
  items: PigletFostering[],
  total = items.length,
  page = 1,
  totalPages = total > 0 ? Math.ceil(total / 10) : 0,
): PigletFosteringPage {
  return { items, total, page, page_size: 10, total_pages: totalPages };
}

class PigletFosteringsStub {
  listPigletFosterings = vi.fn((_filters = {}, _page = 1) =>
    of(pageOf([buildFostering('1'), buildFostering('2')])),
  );
  createPigletFostering = vi.fn(() => of(undefined));
  updatePigletFostering = vi.fn(() => of(undefined));
  deletePigletFostering = vi.fn(() => of(undefined));
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
        { provide: NotificationService, useValue: notifications },
        { provide: AuthService, useValue: { isAdmin } },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => TestBed.createComponent(PigletFosteringsPage).componentInstance as any;

  it('loads the first page on init and renders a card per fostering', async () => {
    const fixture = TestBed.createComponent(PigletFosteringsPage);

    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();

    expect(stub.listPigletFosterings).toHaveBeenCalledWith({}, 1);
    expect(fixture.nativeElement.querySelectorAll('app-piglet-fostering-card')).toHaveLength(2);
  });

  it('shows the error state and retries', async () => {
    stub.listPigletFosterings = vi.fn(() => throwError(() => new Error('failed')));
    const component = create();

    await component.loadFosterings();
    expect(component.error()).toBe(true);

    stub.listPigletFosterings = vi.fn(() => of(pageOf([buildFostering('1')])));
    await component.loadFosterings();

    expect(component.error()).toBe(false);
    expect(component.fosterings()).toHaveLength(1);
  });

  it('stores the page metadata returned by the server', async () => {
    stub.listPigletFosterings = vi.fn(() => of(pageOf([buildFostering('1')], 25, 1, 3)));
    const component = create();

    await component.loadFosterings();

    expect(component.total()).toBe(25);
    expect(component.totalPages()).toBe(3);
  });

  it('moves to the next page and blocks going below the first one', async () => {
    stub.listPigletFosterings = vi.fn(() => of(pageOf([buildFostering('1')], 25, 1, 3)));
    const component = create();
    await component.loadFosterings();

    await component.previousPage();
    expect(component.page()).toBe(1);

    await component.nextPage();
    expect(component.page()).toBe(2);
    expect(stub.listPigletFosterings).toHaveBeenLastCalledWith({}, 2);
  });

  it('does not go past the last page', async () => {
    stub.listPigletFosterings = vi.fn(() => of(pageOf([buildFostering('1')], 10, 1, 1)));
    const component = create();
    await component.loadFosterings();

    await component.nextPage();

    expect(component.page()).toBe(1);
  });

  it('falls back to the last page when the requested page no longer exists', async () => {
    stub.listPigletFosterings = vi.fn((_filters = {}, page = 1) =>
      page === 3 ? of(pageOf([], 5, 3, 1)) : of(pageOf([buildFostering('1')], 5, 1, 1)),
    );
    const component = create();
    await component.loadFosterings();
    component.page.set(3);

    await component.loadFosterings();

    expect(component.page()).toBe(1);
    expect(component.fosterings()).toHaveLength(1);
  });

  it('reloads the first page after registering a fostering', async () => {
    const component = create();
    await component.loadFosterings();
    component.page.set(2);
    component.openModal();

    stub.listPigletFosterings = vi.fn(() =>
      of(pageOf([buildFostering('1'), buildFostering('2'), buildFostering('3')])),
    );
    await component.onCreated({ donor: 'C-001', receiver: 'C-002' });

    expect(notifications.success).toHaveBeenCalledWith(
      'Traslado de "C-001" a "C-002" registrado correctamente',
    );
    expect(component.page()).toBe(1);
    expect(component.fosterings()).toHaveLength(3);
  });

  it('hides the register button for non-admins', () => {
    isAdmin.set(false);
    const fixture = TestBed.createComponent(PigletFosteringsPage);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).not.toContain('Registrar traslado');
  });

  it('opens the edit modal with the selected fostering', () => {
    const component = create();
    const fostering = buildFostering('1');

    component.openEditModal(fostering);

    expect(component.editingFostering()).toBe(fostering);
    expect(component.editOpen()).toBe(true);
    expect(component.modalOpen()).toBe(false);
  });

  it('reloads the list and closes the edit modal after updating a fostering', async () => {
    const component = create();
    await component.loadFosterings();
    component.openEditModal(buildFostering('1'));
    stub.listPigletFosterings = vi.fn(() => of(pageOf([buildFostering('1')])));

    await component.onUpdated();

    expect(notifications.success).toHaveBeenCalledWith('Traslado actualizado correctamente');
    expect(component.editingFostering()).toBeNull();
    expect(component.editOpen()).toBe(false);
    expect(stub.listPigletFosterings).toHaveBeenCalled();
  });

  it('opens and closes the delete confirmation', () => {
    const component = create();
    const fostering = buildFostering('1');

    component.openDelete(fostering);

    expect(component.deletingFostering()).toBe(fostering);
    expect(component.deleteOpen()).toBe(true);

    component.closeDelete();

    expect(component.deleteOpen()).toBe(false);
  });

  it('reloads the list and closes the modal after deleting a fostering', async () => {
    const component = create();
    await component.loadFosterings();
    component.openDelete(buildFostering('1'));
    stub.listPigletFosterings = vi.fn(() => of(pageOf([buildFostering('2')])));

    await component.onDeleted(buildFostering('1'));

    expect(notifications.success).toHaveBeenCalledWith(
      'Traslado de "C-001" a "C-002" eliminado correctamente',
    );
    expect(component.deleteOpen()).toBe(false);
    expect(component.deletingFostering()).toBeNull();
    expect(component.fosterings()).toHaveLength(1);
  });

  it('applies the sow code filter and resets to the first page', async () => {
    const component = create();
    component.page.set(3);
    component.filterForm.setValue({ sow_code: ' C-001 ', from: '', to: '' });

    await component.search();

    expect(stub.listPigletFosterings).toHaveBeenLastCalledWith({ sow_code: 'C-001' }, 1);
    expect(component.page()).toBe(1);
  });

  it('applies the date range filter when searching', async () => {
    const component = create();
    component.filterForm.setValue({ sow_code: '', from: '2026-04-01', to: '2026-04-30' });

    await component.search();

    expect(stub.listPigletFosterings).toHaveBeenLastCalledWith(
      {
        from: '2026-04-01',
        to: '2026-04-30',
      },
      1,
    );
  });

  it('rejects a range whose start is after its end', async () => {
    const component = create();
    component.filterForm.setValue({ sow_code: '', from: '2026-04-30', to: '2026-04-01' });

    await component.search();

    expect(component.rangeError()).toBeTruthy();
    expect(stub.listPigletFosterings).not.toHaveBeenCalled();
  });

  it('clears the filters and reloads without them', async () => {
    const component = create();
    component.filterForm.setValue({ sow_code: 'C-001', from: '', to: '' });
    await component.search();

    stub.listPigletFosterings = vi.fn(() => of(pageOf([buildFostering('1')])));
    await component.clearFilters();

    expect(stub.listPigletFosterings).toHaveBeenCalledWith({}, 1);
    expect(component.filterForm.getRawValue()).toEqual({ sow_code: '', from: '', to: '' });
  });
});
