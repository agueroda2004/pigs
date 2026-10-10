import { WritableSignal, signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { Abortion } from '../../../core/abortions/abortion.models';
import { AbortionsService } from '../../../core/abortions/abortions.service';
import { AuthService } from '../../../core/auth/auth.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { AbortionsPage } from './abortions-page';

function buildAbortion(id: string): Abortion {
  return {
    id,
    sow_id: 'sow-1',
    sow_code: 'C-001',
    service_id: 'service-1',
    abortion_date: '2026-01-15',
    cause: 'Infeccioso',
    note: null,
  };
}

function pageOf(items: Abortion[], total = items.length, totalPages = 1) {
  return { items, total, page: 1, page_size: 10, total_pages: totalPages };
}

class AbortionsStub {
  listAbortions = vi.fn((_filters?: unknown, _page?: number) =>
    of(pageOf([buildAbortion('1'), buildAbortion('2')])),
  );
  createAbortion = vi.fn(() => of(undefined));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('AbortionsPage', () => {
  let stub: AbortionsStub;
  let notifications: NotificationsStub;
  let isAdmin: WritableSignal<boolean>;

  beforeEach(async () => {
    stub = new AbortionsStub();
    notifications = new NotificationsStub();
    isAdmin = signal(true);
    await TestBed.configureTestingModule({
      imports: [AbortionsPage],
      providers: [
        { provide: AbortionsService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
        { provide: AuthService, useValue: { isAdmin } },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => TestBed.createComponent(AbortionsPage).componentInstance as any;

  it('loads abortions on init and renders a card per abortion', async () => {
    const fixture = TestBed.createComponent(AbortionsPage);

    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();

    expect(stub.listAbortions).toHaveBeenCalled();
    expect(fixture.nativeElement.querySelectorAll('app-abortion-card')).toHaveLength(2);
  });

  it('shows the error state and retries', async () => {
    stub.listAbortions = vi.fn(() => throwError(() => new Error('failed')));
    const component = create();

    await component.loadAbortions();
    expect(component.error()).toBe(true);

    stub.listAbortions = vi.fn(() => of(pageOf([buildAbortion('1')])));
    await component.loadAbortions();

    expect(component.error()).toBe(false);
    expect(component.abortions()).toHaveLength(1);
  });

  it('reloads the list and keeps the modal open after registering an abortion', async () => {
    const component = create();
    await component.loadAbortions();
    component.openModal();

    stub.listAbortions = vi.fn(() =>
      of(pageOf([buildAbortion('1'), buildAbortion('2'), buildAbortion('3')])),
    );
    await component.onCreated('C-001');

    expect(notifications.success).toHaveBeenCalledWith(
      'Aborto de la cerda "C-001" registrado correctamente',
    );
    expect(component.modalOpen()).toBe(true);
    expect(component.abortions()).toHaveLength(3);
    expect(stub.listAbortions).toHaveBeenLastCalledWith({}, 1);
  });

  it('opens the edit modal for the selected abortion', () => {
    const component = create();
    const abortion = buildAbortion('9');

    component.openEdit(abortion);

    expect(component.editOpen()).toBe(true);
    expect(component.editingAbortion()).toEqual(abortion);
  });

  it('reloads the list after updating an abortion', async () => {
    const component = create();
    await component.loadAbortions();

    stub.listAbortions = vi.fn(() => of(pageOf([buildAbortion('2')])));
    await component.onUpdated();

    expect(notifications.success).toHaveBeenCalledWith('Aborto actualizado correctamente');
    expect(component.editOpen()).toBe(false);
    expect(component.abortions()).toHaveLength(1);
  });

  it('opens the delete modal for the selected abortion', () => {
    const component = create();
    const abortion = buildAbortion('9');

    component.openDelete(abortion);

    expect(component.deleteOpen()).toBe(true);
    expect(component.deletingAbortion()).toEqual(abortion);
  });

  it('reloads the list after deleting an abortion', async () => {
    const component = create();
    await component.loadAbortions();

    stub.listAbortions = vi.fn(() => of(pageOf([buildAbortion('2')])));
    await component.onDeleted(buildAbortion('9'));

    expect(notifications.success).toHaveBeenCalledWith(
      'Aborto de la cerda "C-001" eliminado correctamente',
    );
    expect(component.deleteOpen()).toBe(false);
    expect(component.abortions()).toHaveLength(1);
  });

  it('hides the register button for non-admins', () => {
    isAdmin.set(false);
    const fixture = TestBed.createComponent(AbortionsPage);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).not.toContain('Registrar aborto');
  });

  it('applies the sow code filter when searching', async () => {
    const component = create();
    component.filterForm.setValue({ sow_code: 'C-001' });

    await component.search();

    expect(stub.listAbortions).toHaveBeenLastCalledWith({ sow_code: 'C-001' }, 1);
  });

  it('resets to the first page when searching', async () => {
    const component = create();
    component.page.set(3);
    component.filterForm.setValue({ sow_code: 'C-001' });

    await component.search();

    expect(component.page()).toBe(1);
    expect(stub.listAbortions).toHaveBeenLastCalledWith({ sow_code: 'C-001' }, 1);
  });

  it('clears the filters and reloads without them', async () => {
    const component = create();
    component.filterForm.setValue({ sow_code: 'C-001' });
    await component.search();

    stub.listAbortions = vi.fn(() => of(pageOf([buildAbortion('1')])));
    await component.clearFilters();

    expect(stub.listAbortions).toHaveBeenCalledWith({}, 1);
    expect(component.filterForm.getRawValue()).toEqual({ sow_code: '' });
  });

  it('paginates forward and backward', async () => {
    stub.listAbortions = vi.fn((_filters, page) =>
      of(pageOf([buildAbortion(String(page))], 25, 3)),
    );
    const component = create();
    await component.loadAbortions();

    expect(component.page()).toBe(1);
    expect(component.totalPages()).toBe(3);

    await component.nextPage();
    expect(component.page()).toBe(2);
    expect(stub.listAbortions).toHaveBeenLastCalledWith({}, 2);

    await component.previousPage();
    expect(component.page()).toBe(1);
    expect(stub.listAbortions).toHaveBeenLastCalledWith({}, 1);
  });
});
