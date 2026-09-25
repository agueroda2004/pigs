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
import { SowsService } from '../../../core/sows/sows.service';
import { ServicesPage } from './services-page';

function buildService(id: string): Service {
  return {
    id,
    sow_id: 'sow-1',
    expected_farrowing_date: '2025-05-04',
    note: null,
    state: 'Confirmado',
    location: 'Nave 1',
    mounts: [],
    created_at: '2025-01-10T12:00:00',
    updated_at: '2025-01-10T12:00:00',
    created_by: 'admin',
    updated_by: 'admin',
  };
}

class ServicesStub {
  listServices = vi.fn(() => of([buildService('1'), buildService('2')]));
  createService = vi.fn(() => of(buildService('3')));
}

class SowsStub {
  listSows = vi.fn(() => of([{ id: 'sow-1', code: 'C-001' }]));
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
        { provide: SowsService, useValue: new SowsStub() },
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

  it('resolves the sow code from the lookup', async () => {
    const component = create();
    await component.loadLookups();

    expect(component.sowCode(buildService('1'))).toBe('C-001');
  });

  it('shows the error state and retries', async () => {
    stub.listServices = vi.fn(() => throwError(() => new Error('failed')));
    const component = create();

    await component.loadServices();
    expect(component.error()).toBe(true);

    stub.listServices = vi.fn(() => of([buildService('1')]));
    await component.loadServices();

    expect(component.error()).toBe(false);
    expect(component.services()).toHaveLength(1);
  });

  it('reloads the list and keeps the modal open after creating a service', async () => {
    const component = create();
    await component.loadServices();
    component.openModal();

    stub.listServices = vi.fn(() => of([buildService('1'), buildService('2'), buildService('3')]));
    await component.onCreated('C-001');

    expect(notifications.success).toHaveBeenCalledWith(
      'Servicio para la cerda "C-001" creado correctamente',
    );
    expect(component.modalOpen()).toBe(true);
    expect(component.services()).toHaveLength(3);
  });

  it('hides the create button for non-admins', () => {
    isAdmin.set(false);
    const fixture = TestBed.createComponent(ServicesPage);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).not.toContain('Crear servicio');
  });

  it('applies the filter values when searching', async () => {
    const component = create();
    component.filterForm.setValue({ sow_id: 'sow-1', state: 'Fallido' });

    await component.search();

    expect(stub.listServices).toHaveBeenLastCalledWith({ sow_id: 'sow-1', state: 'Fallido' });
  });

  it('clears the filters and reloads without them', async () => {
    const component = create();
    component.filterForm.setValue({ sow_id: 'sow-1', state: 'Fallido' });
    await component.search();

    stub.listServices = vi.fn(() => of([buildService('1')]));
    await component.clearFilters();

    expect(stub.listServices).toHaveBeenCalledWith({});
    expect(component.filterForm.getRawValue()).toEqual({ sow_id: '', state: '' });
  });

  it('shows a filtered empty message when filters match nothing', async () => {
    stub.listServices = vi.fn(() => of([]));
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
