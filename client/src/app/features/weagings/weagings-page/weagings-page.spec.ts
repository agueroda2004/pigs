import { WritableSignal, signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { AuthService } from '../../../core/auth/auth.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { Weaging, WeagingFilters, WeagingPage } from '../../../core/weagings/weaging.models';
import { WeagingsService } from '../../../core/weagings/weagings.service';
import { WeagingsPage } from './weagings-page';

function buildWeaging(id: string): Weaging {
  return {
    id,
    farrowing_id: 'farrowing-1',
    sow_id: 'sow-1',
    sow_code: 'C-001',
    weaging_date: '2026-04-25',
    quantity: 8,
    total_weight: 120.5,
    destination: 'Nave 2',
    note: null,
  };
}

function buildPage(
  items: Weaging[] = [buildWeaging('1'), buildWeaging('2')],
  overrides: Partial<WeagingPage> = {},
): WeagingPage {
  return {
    items,
    total: items.length,
    page: 1,
    page_size: 10,
    total_pages: 1,
    ...overrides,
  };
}

class WeagingsStub {
  listWeagings = vi.fn((_filters: WeagingFilters = {}, _page = 1) => of(buildPage()));
  createWeaging = vi.fn(() => of(void 0));
  updateWeaging = vi.fn(() => of(void 0));
  deleteWeaging = vi.fn(() => of(void 0));
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

    expect(stub.listWeagings).toHaveBeenCalledWith({}, 1);
    expect(fixture.nativeElement.querySelectorAll('app-weaging-card')).toHaveLength(2);
  });

  it('renders the sow code returned by the API', async () => {
    const fixture = TestBed.createComponent(WeagingsPage);
    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('C-001');
  });

  it('shows the error state and retries', async () => {
    stub.listWeagings = vi.fn(() => throwError(() => new Error('failed')));
    const component = create();

    await component.loadWeagings();
    expect(component.error()).toBe(true);

    stub.listWeagings = vi.fn(() => of(buildPage([buildWeaging('1')])));
    await component.loadWeagings();

    expect(component.error()).toBe(false);
    expect(component.weagings()).toHaveLength(1);
  });

  it('returns to the first page and reloads after registering a weaging', async () => {
    const component = create();
    component.page.set(2);

    stub.listWeagings = vi.fn(() =>
      of(buildPage([buildWeaging('1'), buildWeaging('2'), buildWeaging('3')])),
    );
    await component.onCreated('C-001');

    expect(notifications.success).toHaveBeenCalledWith(
      'Destete de "C-001" registrado correctamente',
    );
    expect(component.page()).toBe(1);
    expect(component.weagings()).toHaveLength(3);
  });

  it('hides the register button for non-admins', () => {
    isAdmin.set(false);
    const fixture = TestBed.createComponent(WeagingsPage);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).not.toContain('Registrar destete');
  });

  it('hides the delete action for non-admins', async () => {
    isAdmin.set(false);
    const fixture = TestBed.createComponent(WeagingsPage);
    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).not.toContain('Eliminar');
  });

  it('opens and cancels the delete confirmation', () => {
    const component = create();

    component.openDelete(buildWeaging('1'));

    expect(component.deleteOpen()).toBe(true);
    expect(component.selectedWeaging()?.id).toBe('1');
    expect(component.selectedSowCode()).toBe('C-001');

    component.closeDelete();

    expect(component.deleteOpen()).toBe(false);
    expect(component.selectedWeaging()).toBeNull();
  });

  it('hides the edit action for non-admins', async () => {
    isAdmin.set(false);
    const fixture = TestBed.createComponent(WeagingsPage);
    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).not.toContain('Editar');
  });

  it('opens and cancels the edit modal', () => {
    const component = create();

    component.openEdit(buildWeaging('1'));

    expect(component.editOpen()).toBe(true);
    expect(component.editingWeaging()?.id).toBe('1');
    expect(component.editingSowCode()).toBe('C-001');

    component.closeEdit();

    expect(component.editOpen()).toBe(false);
    expect(component.editingWeaging()).toBeNull();
  });

  it('closes the modal, notifies and reloads the filtered list after updating', async () => {
    const component = create();
    component.filterForm.setValue({ sow_code: 'C-001', from: '', to: '' });
    await component.search();

    component.openEdit(buildWeaging('1'));
    stub.listWeagings = vi.fn(() => of(buildPage([buildWeaging('2')])));
    await component.onUpdated();

    expect(component.editOpen()).toBe(false);
    expect(component.editingWeaging()).toBeNull();
    expect(notifications.success).toHaveBeenCalledWith(
      'Destete de "C-001" actualizado correctamente',
    );
    expect(stub.listWeagings).toHaveBeenLastCalledWith({ sow_code: 'C-001' }, 1);
    expect(component.weagings()).toHaveLength(1);
  });

  it('closes the modal, notifies and reloads the filtered list after deleting', async () => {
    const component = create();
    component.filterForm.setValue({ sow_code: 'C-001', from: '', to: '' });
    await component.search();

    component.openDelete(buildWeaging('1'));
    stub.listWeagings = vi.fn(() => of(buildPage([buildWeaging('2')])));
    await component.onDeleted(buildWeaging('1'));

    expect(component.deleteOpen()).toBe(false);
    expect(component.selectedWeaging()).toBeNull();
    expect(notifications.success).toHaveBeenCalledWith(
      'Destete de "C-001" eliminado correctamente',
    );
    expect(stub.listWeagings).toHaveBeenLastCalledWith({ sow_code: 'C-001' }, 1);
    expect(component.weagings()).toHaveLength(1);
  });

  it('trims the sow code and returns to the first page when searching', async () => {
    const component = create();
    component.page.set(2);
    component.filterForm.setValue({ sow_code: '  C-001  ', from: '', to: '' });

    await component.search();

    expect(component.page()).toBe(1);
    expect(stub.listWeagings).toHaveBeenLastCalledWith({ sow_code: 'C-001' }, 1);
  });

  it('omits a blank sow code from the filters', async () => {
    const component = create();
    component.filterForm.setValue({ sow_code: '   ', from: '', to: '' });

    await component.search();

    expect(stub.listWeagings).toHaveBeenLastCalledWith({}, 1);
  });

  it('applies the date range filter when searching', async () => {
    const component = create();
    component.filterForm.setValue({ sow_code: '', from: '2026-04-01', to: '2026-04-30' });

    await component.search();

    expect(stub.listWeagings).toHaveBeenLastCalledWith(
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
    expect(stub.listWeagings).not.toHaveBeenCalled();
  });

  it('clears the filters and reloads the first page without them', async () => {
    const component = create();
    component.filterForm.setValue({ sow_code: 'C-001', from: '', to: '' });
    await component.search();

    stub.listWeagings = vi.fn(() => of(buildPage([buildWeaging('1')])));
    await component.clearFilters();

    expect(stub.listWeagings).toHaveBeenCalledWith({}, 1);
    expect(component.filterForm.getRawValue()).toEqual({ sow_code: '', from: '', to: '' });
  });

  it('moves between pages and keeps the filters', async () => {
    stub.listWeagings = vi.fn(() =>
      of(buildPage([buildWeaging('1')], { total: 12, total_pages: 2 })),
    );
    const component = create();
    component.appliedFilters.set({ sow_code: 'C-001' });
    await component.loadWeagings();

    await component.nextPage();
    expect(component.page()).toBe(2);
    expect(stub.listWeagings).toHaveBeenLastCalledWith({ sow_code: 'C-001' }, 2);

    await component.previousPage();
    expect(component.page()).toBe(1);
    expect(stub.listWeagings).toHaveBeenLastCalledWith({ sow_code: 'C-001' }, 1);
  });

  it('ignores navigation beyond the available pages', async () => {
    stub.listWeagings = vi.fn(() =>
      of(buildPage([buildWeaging('1')], { total: 12, total_pages: 2 })),
    );
    const component = create();
    await component.loadWeagings();

    await component.previousPage();
    expect(component.page()).toBe(1);

    await component.nextPage();
    await component.nextPage();
    expect(component.page()).toBe(2);
  });

  it('clamps an out-of-range page and reloads the last available page', async () => {
    const component = create();
    component.page.set(3);
    stub.listWeagings = vi.fn(() =>
      of(buildPage([buildWeaging('1')], { total: 11, page: 3, total_pages: 2 })),
    );

    await component.loadWeagings();

    expect(component.page()).toBe(2);
    expect(component.totalPages()).toBe(2);
    expect(component.weagings()).toHaveLength(1);
    expect(stub.listWeagings).toHaveBeenLastCalledWith({}, 2);
  });
});
