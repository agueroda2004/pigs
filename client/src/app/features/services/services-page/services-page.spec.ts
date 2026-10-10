import { WritableSignal, signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { AuthService } from '../../../core/auth/auth.service';
import { BoarsService } from '../../../core/boars/boars.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { OperatorsService } from '../../../core/operators/operators.service';
import { Service } from '../../../core/services/service.models';
import { ServicesService } from '../../../core/services/services.service';
import { ServicesPage } from './services-page';

function buildService(id: string): Service {
  return {
    id,
    sow_id: 'sow-1',
    sow_code: 'C-001',
    expected_farrowing_date: '2025-05-04',
    note: null,
    state: 'Confirmado',
    location: 'Nave 1',
    mounts: [],
  };
}

function pageOf(items: Service[], total = items.length, totalPages = 1) {
  return { items, total, page: 1, page_size: 10, total_pages: totalPages };
}

class ServicesStub {
  listServices = vi.fn((_filters?: unknown, _page?: number) =>
    of(pageOf([buildService('1'), buildService('2')])),
  );
  createService = vi.fn(() => of(undefined));
  updateService = vi.fn(() => of(undefined));
  deleteService = vi.fn(() => of(undefined));
}

class BoarsStub {
  listBoars = vi.fn(() => of([{ id: 'boar-1', code: 'V-001' }]));
}

class OperatorsStub {
  listOperators = vi.fn(() => of([{ id: 'operator-1', name: 'Ana' }]));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('ServicesPage', () => {
  let stub: ServicesStub;
  let notifications: NotificationsStub;
  let isAdmin: WritableSignal<boolean>;

  beforeEach(async () => {
    stub = new ServicesStub();
    notifications = new NotificationsStub();
    isAdmin = signal(true);
    await TestBed.configureTestingModule({
      imports: [ServicesPage],
      providers: [
        { provide: ServicesService, useValue: stub },
        { provide: BoarsService, useValue: new BoarsStub() },
        { provide: OperatorsService, useValue: new OperatorsStub() },
        { provide: NotificationService, useValue: notifications },
        { provide: AuthService, useValue: { isAdmin } },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => TestBed.createComponent(ServicesPage).componentInstance as any;

  it('loads services on init and renders a card per service', async () => {
    const fixture = TestBed.createComponent(ServicesPage);
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;

    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();

    expect(stub.listServices).toHaveBeenCalled();
    expect(component.services()).toHaveLength(2);
    expect(fixture.nativeElement.querySelectorAll('app-service-card')).toHaveLength(2);
  });

  it('shows the error state and retries', async () => {
    stub.listServices = vi.fn(() => throwError(() => new Error('failed')));
    const component = create();

    await component.loadServices();
    expect(component.error()).toBe(true);

    stub.listServices = vi.fn(() => of(pageOf([buildService('1')])));
    await component.loadServices();

    expect(component.error()).toBe(false);
    expect(component.services()).toHaveLength(1);
  });

  it('reloads the list and keeps the modal open after creating a service', async () => {
    const component = create();
    await component.loadServices();
    component.openModal();

    stub.listServices = vi.fn(() =>
      of(pageOf([buildService('1'), buildService('2'), buildService('3')])),
    );
    await component.onCreated('C-001');

    expect(notifications.success).toHaveBeenCalledWith(
      'Servicio para la cerda "C-001" creado correctamente',
    );
    expect(component.modalOpen()).toBe(true);
    expect(component.services()).toHaveLength(3);
    expect(stub.listServices).toHaveBeenLastCalledWith({}, 1);
  });

  it('opens the edit modal for the selected service', () => {
    const component = create();
    const service = buildService('9');

    component.openEdit(service);

    expect(component.editOpen()).toBe(true);
    expect(component.editingService()).toEqual(service);
  });

  it('reloads the list after updating a service', async () => {
    const component = create();
    await component.loadServices();

    stub.listServices = vi.fn(() => of(pageOf([buildService('2')])));
    await component.onUpdated();

    expect(notifications.success).toHaveBeenCalledWith('Servicio actualizado correctamente');
    expect(component.editOpen()).toBe(false);
    expect(component.services()).toHaveLength(1);
  });

  it('opens the delete modal for the selected service', () => {
    const component = create();
    const service = buildService('9');

    component.openDelete(service);

    expect(component.deleteOpen()).toBe(true);
    expect(component.deletingService()).toEqual(service);
  });

  it('reloads the list after deleting a service', async () => {
    const component = create();
    await component.loadServices();

    stub.listServices = vi.fn(() => of(pageOf([buildService('2')])));
    await component.onDeleted(buildService('9'));

    expect(notifications.success).toHaveBeenCalledWith(
      'Servicio de la cerda "C-001" eliminado correctamente',
    );
    expect(component.deleteOpen()).toBe(false);
    expect(component.services()).toHaveLength(1);
  });

  it('hides the create button for non-admins', () => {
    isAdmin.set(false);
    const fixture = TestBed.createComponent(ServicesPage);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).not.toContain('Crear servicio');
  });

  it('applies the filter values when searching', async () => {
    const component = create();
    component.filterForm.setValue({ sow_code: 'C-001', state: 'Fallido' });

    await component.search();

    expect(stub.listServices).toHaveBeenLastCalledWith({ sow_code: 'C-001', state: 'Fallido' }, 1);
  });

  it('clears the filters and reloads without them', async () => {
    const component = create();
    component.filterForm.setValue({ sow_code: 'C-001', state: 'Fallido' });
    await component.search();

    stub.listServices = vi.fn(() => of(pageOf([buildService('1')])));
    await component.clearFilters();

    expect(stub.listServices).toHaveBeenCalledWith({}, 1);
    expect(component.filterForm.getRawValue()).toEqual({ sow_code: '', state: '' });
  });

  it('paginates forward and backward', async () => {
    stub.listServices = vi.fn((_filters, page) => of(pageOf([buildService(String(page))], 25, 3)));
    const component = create();
    await component.loadServices();

    expect(component.page()).toBe(1);
    expect(component.totalPages()).toBe(3);

    await component.nextPage();
    expect(component.page()).toBe(2);
    expect(stub.listServices).toHaveBeenLastCalledWith({}, 2);

    await component.previousPage();
    expect(component.page()).toBe(1);
    expect(stub.listServices).toHaveBeenLastCalledWith({}, 1);
  });

  it('shows a filtered empty message when filters match nothing', async () => {
    stub.listServices = vi.fn(() => of(pageOf([], 0, 0)));
    const fixture = TestBed.createComponent(ServicesPage);
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;
    fixture.detectChanges();
    await fixture.whenStable();

    component.filterForm.patchValue({ state: 'Fallido' });
    await component.search();
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain(
      'No hay servicios que coincidan con los filtros',
    );
  });
});
