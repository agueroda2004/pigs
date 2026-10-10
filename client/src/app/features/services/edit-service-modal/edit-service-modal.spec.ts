import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { BoarsService } from '../../../core/boars/boars.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { OperatorsService } from '../../../core/operators/operators.service';
import { Service } from '../../../core/services/service.models';
import { ServicesService } from '../../../core/services/services.service';
import { EditServiceModal } from './edit-service-modal';

function buildService(overrides: Partial<Service> = {}): Service {
  return {
    id: '42',
    sow_id: 'sow-1',
    sow_code: 'C-001',
    expected_farrowing_date: '2026-05-04',
    note: null,
    state: 'Confirmado',
    location: 'Nave 1',
    mounts: [
      {
        id: 'm1',
        service_id: '42',
        boar_id: 'boar-1',
        boar_code: 'V-001',
        operator_id: 'operator-1',
        operator_name: 'Ana',
        mount_number: 1,
        mount_date: '2026-01-10',
        type: 'Artificial',
        note: null,
      },
    ],
    ...overrides,
  };
}

class ServicesStub {
  updateService = vi.fn((_id: string, _request: unknown) => of(undefined));
}

class BoarsStub {
  listBoarDropdown = vi.fn(() => of([{ id: 'boar-1', code: 'V-001', active: true }]));
}

class OperatorsStub {
  listOperatorDropdown = vi.fn(() => of([{ id: 'operator-1', name: 'Ana', active: true }]));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('EditServiceModal', () => {
  let stub: ServicesStub;
  let boars: BoarsStub;
  let operators: OperatorsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new ServicesStub();
    boars = new BoarsStub();
    operators = new OperatorsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [EditServiceModal],
      providers: [
        { provide: ServicesService, useValue: stub },
        { provide: BoarsService, useValue: boars },
        { provide: OperatorsService, useValue: operators },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = async (service: Service = buildService()): Promise<any> => {
    const fixture = TestBed.createComponent(EditServiceModal);
    fixture.componentRef.setInput('service', service);
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();
    await fixture.whenStable();
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return fixture.componentInstance as any;
  };

  it('prefills the form from the service', async () => {
    const component = await create();

    expect(component.form.getRawValue().location).toBe('Nave 1');
    expect(component.mounts.length).toBe(1);
    expect(component.mounts.at(0).getRawValue()).toMatchObject({
      id: 'm1',
      mount_date: '2026-01-10',
      boar_id: 'boar-1',
      operator_id: 'operator-1',
      type: 'Artificial',
    });
  });

  it('renders the first mount values when the service has several mounts', async () => {
    const service = buildService({
      mounts: [
        {
          id: 'm1',
          service_id: '42',
          boar_id: 'boar-1',
          boar_code: 'V-001',
          operator_id: 'operator-1',
          operator_name: 'Ana',
          mount_number: 1,
          mount_date: '2026-01-10',
          type: 'Artificial',
          note: null,
        },
        {
          id: 'm2',
          service_id: '42',
          boar_id: 'boar-1',
          boar_code: 'V-001',
          operator_id: 'operator-1',
          operator_name: 'Ana',
          mount_number: 2,
          mount_date: '2026-01-11',
          type: 'Artificial',
          note: null,
        },
      ],
    });
    const fixture = TestBed.createComponent(EditServiceModal);
    fixture.componentRef.setInput('service', service);
    fixture.detectChanges();
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('10 de Enero 2026');
    expect(text).toContain('11 de Enero 2026');
  });

  it('labels inactive boars and operators in the dropdowns', async () => {
    boars.listBoarDropdown = vi.fn(() => of([{ id: 'boar-1', code: 'V-001', active: false }]));
    operators.listOperatorDropdown = vi.fn(() =>
      of([{ id: 'operator-1', name: 'Ana', active: false }]),
    );
    const component = await create();

    await component.loadOptions();

    expect(component.boarOptions()).toEqual([{ value: 'boar-1', label: 'V-001 (inactivo)' }]);
    expect(component.operatorOptions()).toEqual([{ value: 'operator-1', label: 'Ana (inactivo)' }]);
  });

  it('does not call the API when nothing changed', async () => {
    const component = await create();
    const updated = vi.fn();
    component.updated.subscribe(updated);

    await component.submit();

    expect(stub.updateService).not.toHaveBeenCalled();
    expect(updated).toHaveBeenCalled();
  });

  it('sends only the changed mounts in the update array', async () => {
    const component = await create();
    component.mounts.at(0).patchValue({ mount_date: '2026-01-11' });

    await component.submit();

    expect(stub.updateService).toHaveBeenCalledWith('42', {
      location: 'Nave 1',
      note: '',
      mounts: {
        create: [],
        update: [
          {
            id: 'm1',
            boar_id: 'boar-1',
            operator_id: 'operator-1',
            mount_date: '2026-01-11',
            type: 'Artificial',
            note: '',
          },
        ],
        delete: [],
      },
    });
  });

  it('sends new mounts in the create array', async () => {
    const component = await create();
    component.addMount();
    component.mounts.at(1).patchValue({
      mount_date: '2026-01-11',
      boar_id: 'boar-1',
      operator_id: 'operator-1',
      type: 'Natural',
      note: '',
    });

    await component.submit();

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const request = stub.updateService.mock.calls[0][1] as any;
    expect(request.mounts.update).toHaveLength(0);
    expect(request.mounts.delete).toHaveLength(0);
    expect(request.mounts.create).toEqual([
      {
        boar_id: 'boar-1',
        operator_id: 'operator-1',
        mount_date: '2026-01-11',
        type: 'Natural',
        note: '',
      },
    ]);
  });

  it('sends removed mounts in the delete array', async () => {
    const service = buildService({
      mounts: [
        {
          id: 'm1',
          service_id: '42',
          boar_id: 'boar-1',
          boar_code: 'V-001',
          operator_id: 'operator-1',
          operator_name: 'Ana',
          mount_number: 1,
          mount_date: '2026-01-10',
          type: 'Artificial',
          note: null,
        },
        {
          id: 'm2',
          service_id: '42',
          boar_id: 'boar-1',
          boar_code: 'V-001',
          operator_id: 'operator-1',
          operator_name: 'Ana',
          mount_number: 2,
          mount_date: '2026-01-11',
          type: 'Artificial',
          note: null,
        },
      ],
    });
    const component = await create(service);
    component.removeMount(1);

    await component.submit();

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const request = stub.updateService.mock.calls[0][1] as any;
    expect(request.mounts.delete).toEqual(['m2']);
  });

  it('shows the server message when the update fails', async () => {
    stub.updateService = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 409,
            error: { error: 'Solo se pueden editar servicios en estado Confirmado' },
          }),
      ),
    );
    const component = await create();
    component.form.controls.location.setValue('Nave 2');

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith(
      'Solo se pueden editar servicios en estado Confirmado',
    );
  });

  it('shows a fallback when the service is missing', async () => {
    stub.updateService = vi.fn(() => throwError(() => new HttpErrorResponse({ status: 404 })));
    const component = await create();
    component.form.controls.location.setValue('Nave 2');

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith('El servicio no existe');
  });
});
