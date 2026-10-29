import { WritableSignal, signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { AuthService } from '../../../core/auth/auth.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import {
  PartialWeaging,
  PartialWeagingFilters,
  PartialWeagingPage,
} from '../../../core/partial-weagings/partial-weaging.models';
import { PartialWeagingsService } from '../../../core/partial-weagings/partial-weagings.service';
import { PartialWeagingsPage } from './partial-weagings-page';

function buildWeaging(id: string): PartialWeaging {
  return {
    id,
    farrowing_id: 'farrowing-1',
    sow_id: 'sow-1',
    sow_code: 'C-001',
    weaging_date: '2026-04-25',
    quantity: 2,
    total_weight: 42.5,
    type: 'Normal',
    note: null,
  };
}

function pageOf(items: PartialWeaging[], total = items.length, page = 1): PartialWeagingPage {
  return {
    items,
    total,
    page,
    page_size: 10,
    total_pages: total > 0 ? Math.ceil(total / 10) : 0,
  };
}

class PartialWeagingsStub {
  listPartialWeagings = vi.fn((_filters: PartialWeagingFilters = {}, page = 1) =>
    of(pageOf([buildWeaging('1'), buildWeaging('2')], 2, page)),
  );
  createPartialWeaging = vi.fn(() => of(void 0));
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
        { provide: NotificationService, useValue: notifications },
        { provide: AuthService, useValue: { isAdmin } },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => TestBed.createComponent(PartialWeagingsPage).componentInstance as any;

  it('loads the first page on init and renders a card per weaging', async () => {
    const fixture = TestBed.createComponent(PartialWeagingsPage);

    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();

    expect(stub.listPartialWeagings).toHaveBeenCalledWith({}, 1);
    expect(fixture.nativeElement.querySelectorAll('app-partial-weaging-card')).toHaveLength(2);
  });

  it('shows the error state and retries', async () => {
    stub.listPartialWeagings = vi.fn(() => throwError(() => new Error('failed')));
    const component = create();

    await component.loadWeagings();
    expect(component.error()).toBe(true);

    stub.listPartialWeagings = vi.fn(() => of(pageOf([buildWeaging('1')])));
    await component.loadWeagings();

    expect(component.error()).toBe(false);
    expect(component.weagings()).toHaveLength(1);
  });

  it('reloads the list from the first page after registering a weaging', async () => {
    const component = create();
    component.page.set(2);
    stub.listPartialWeagings = vi.fn(() =>
      of(pageOf([buildWeaging('1'), buildWeaging('2'), buildWeaging('3')])),
    );

    await component.onCreated('C-001');

    expect(notifications.success).toHaveBeenCalledWith(
      'Destete parcial de "C-001" registrado correctamente',
    );
    expect(component.page()).toBe(1);
    expect(stub.listPartialWeagings).toHaveBeenLastCalledWith({}, 1);
    expect(component.weagings()).toHaveLength(3);
  });

  it('hides the register button for non-admins', () => {
    isAdmin.set(false);
    const fixture = TestBed.createComponent(PartialWeagingsPage);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).not.toContain('Registrar destete');
  });

  it('opens the delete modal with the sow code from the response', () => {
    const component = create();
    const weaging = buildWeaging('1');

    component.openDelete(weaging);

    expect(component.deleteOpen()).toBe(true);
    expect(component.selectedWeaging()).toEqual(weaging);
    expect(component.selectedSowCode()).toBe('C-001');
  });

  it('closes the delete modal and clears the selection', () => {
    const component = create();
    component.openDelete(buildWeaging('1'));

    component.closeDelete();

    expect(component.deleteOpen()).toBe(false);
    expect(component.selectedWeaging()).toBeNull();
  });

  it('reloads the current page after deleting a weaging', async () => {
    const component = create();
    component.page.set(2);
    stub.listPartialWeagings = vi.fn(() => of(pageOf([buildWeaging('2')], 15, 2)));

    await component.onDeleted(buildWeaging('1'));

    expect(notifications.success).toHaveBeenCalledWith(
      'Destete parcial de "C-001" eliminado correctamente',
    );
    expect(component.deleteOpen()).toBe(false);
    expect(component.selectedWeaging()).toBeNull();
    expect(stub.listPartialWeagings).toHaveBeenLastCalledWith({}, 2);
    expect(component.weagings()).toHaveLength(1);
  });

  it('hides the delete action for non-admins', () => {
    isAdmin.set(false);
    const fixture = TestBed.createComponent(PartialWeagingsPage);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).not.toContain('Eliminar');
  });

  it('opens the edit modal with the sow code from the response', () => {
    const component = create();
    const weaging = buildWeaging('1');

    component.openEdit(weaging);

    expect(component.editOpen()).toBe(true);
    expect(component.editingWeaging()).toEqual(weaging);
    expect(component.editingSowCode()).toBe('C-001');
  });

  it('closes the edit modal and clears the selection', () => {
    const component = create();
    component.openEdit(buildWeaging('1'));

    component.closeEdit();

    expect(component.editOpen()).toBe(false);
    expect(component.editingWeaging()).toBeNull();
  });

  it('reloads the current page after updating a weaging', async () => {
    const component = create();
    component.page.set(2);
    component.openEdit(buildWeaging('1'));
    stub.listPartialWeagings = vi.fn(() => of(pageOf([buildWeaging('1')], 15, 2)));

    await component.onUpdated();

    expect(notifications.success).toHaveBeenCalledWith(
      'Destete parcial de "C-001" actualizado correctamente',
    );
    expect(component.editOpen()).toBe(false);
    expect(component.editingWeaging()).toBeNull();
    expect(stub.listPartialWeagings).toHaveBeenLastCalledWith({}, 2);
    expect(component.weagings()).toHaveLength(1);
  });

  it('hides the edit action for non-admins', () => {
    isAdmin.set(false);
    const fixture = TestBed.createComponent(PartialWeagingsPage);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).not.toContain('Editar');
  });

  it('applies the trimmed sow code filter and resets the page', async () => {
    const component = create();
    component.page.set(3);
    component.filterForm.setValue({ sow_code: ' C-001 ', from: '', to: '' });

    await component.search();

    expect(stub.listPartialWeagings).toHaveBeenLastCalledWith({ sow_code: 'C-001' }, 1);
    expect(component.page()).toBe(1);
  });

  it('applies the date range filter when searching', async () => {
    const component = create();
    component.filterForm.setValue({ sow_code: '', from: '2026-04-01', to: '2026-04-30' });

    await component.search();

    expect(stub.listPartialWeagings).toHaveBeenLastCalledWith(
      {
        from: '2026-04-01',
        to: '2026-04-30',
      },
      1,
    );
  });

  it('rejects a sow code longer than 50 characters', async () => {
    const component = create();
    component.filterForm.setValue({ sow_code: 'A'.repeat(51), from: '', to: '' });

    await component.search();

    expect(component.filterForm.controls.sow_code.hasError('maxlength')).toBe(true);
    expect(stub.listPartialWeagings).not.toHaveBeenCalled();
  });

  it('rejects a range whose start is after its end', async () => {
    const component = create();
    component.filterForm.setValue({ sow_code: '', from: '2026-04-30', to: '2026-04-01' });

    await component.search();

    expect(component.rangeError()).toBeTruthy();
    expect(stub.listPartialWeagings).not.toHaveBeenCalled();
  });

  it('clears the filters and reloads from the first page without them', async () => {
    const component = create();
    component.filterForm.setValue({ sow_code: 'C-001', from: '', to: '' });
    await component.search();

    await component.clearFilters();

    expect(stub.listPartialWeagings).toHaveBeenLastCalledWith({}, 1);
    expect(component.filterForm.getRawValue()).toEqual({ sow_code: '', from: '', to: '' });
  });

  it('moves between pages', async () => {
    stub.listPartialWeagings = vi.fn((_filters: PartialWeagingFilters = {}, page = 1) =>
      of(pageOf([buildWeaging(String(page))], 25, page)),
    );
    const component = create();
    await component.loadWeagings();

    expect(component.totalPages()).toBe(3);
    expect(component.total()).toBe(25);

    await component.nextPage();
    expect(component.page()).toBe(2);
    expect(stub.listPartialWeagings).toHaveBeenLastCalledWith({}, 2);

    await component.previousPage();
    expect(component.page()).toBe(1);
    expect(stub.listPartialWeagings).toHaveBeenLastCalledWith({}, 1);
  });

  it('clamps the page when the server returns fewer pages', async () => {
    stub.listPartialWeagings = vi.fn(() => of(pageOf([buildWeaging('1')], 15, 1)));
    const component = create();
    component.page.set(3);

    await component.loadWeagings();

    expect(component.page()).toBe(2);
    expect(stub.listPartialWeagings).toHaveBeenLastCalledWith({}, 2);
  });

  it('renders pagination controls when there is more than one page', async () => {
    stub.listPartialWeagings = vi.fn(() => of(pageOf([buildWeaging('1')], 25, 1)));
    const fixture = TestBed.createComponent(PartialWeagingsPage);

    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('Página 1 de 3 (25 destetes)');
  });
});
