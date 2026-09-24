import { WritableSignal, signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { AuthService } from '../../../core/auth/auth.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { Operator } from '../../../core/operators/operator.models';
import { OperatorsService } from '../../../core/operators/operators.service';
import { OperatorsPage } from './operators-page';

function buildOperator(id: string): Operator {
  return {
    id,
    name: `Operator ${id}`,
    active: true,
    created_at: '2026-01-02T12:00:00',
    updated_at: '2026-01-02T12:00:00',
    created_by: 'admin',
    updated_by: 'admin',
  };
}

class OperatorsStub {
  listOperators = vi.fn(() => of([buildOperator('1'), buildOperator('2')]));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('OperatorsPage', () => {
  let stub: OperatorsStub;
  let notifications: NotificationsStub;
  let isAdmin: WritableSignal<boolean>;

  beforeEach(async () => {
    stub = new OperatorsStub();
    notifications = new NotificationsStub();
    isAdmin = signal(true);
    await TestBed.configureTestingModule({
      imports: [OperatorsPage],
      providers: [
        { provide: OperatorsService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
        { provide: AuthService, useValue: { isAdmin } },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => TestBed.createComponent(OperatorsPage).componentInstance as any;

  it('loads operators on init and renders a card per operator', async () => {
    const fixture = TestBed.createComponent(OperatorsPage);
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;

    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();

    expect(stub.listOperators).toHaveBeenCalled();
    expect(component.operators()).toHaveLength(2);
    expect(fixture.nativeElement.querySelectorAll('app-operator-card')).toHaveLength(2);
  });

  it('shows the error state and retries', async () => {
    stub.listOperators = vi.fn(() => throwError(() => new Error('failed')));
    const component = create();

    await component.loadOperators();
    expect(component.error()).toBe(true);

    stub.listOperators = vi.fn(() => of([buildOperator('1')]));
    await component.loadOperators();

    expect(component.error()).toBe(false);
    expect(component.operators()).toHaveLength(1);
  });

  it('reloads the list after creating an operator', async () => {
    const component = create();
    await component.loadOperators();

    stub.listOperators = vi.fn(() =>
      of([buildOperator('1'), buildOperator('2'), buildOperator('3')]),
    );
    await component.onCreated('Juan Pérez');

    expect(notifications.success).toHaveBeenCalledWith(
      'Operador "Juan Pérez" creado correctamente',
    );
    expect(component.modalOpen()).toBe(false);
    expect(component.operators()).toHaveLength(3);
  });

  it('reloads the list after updating an operator', async () => {
    const component = create();
    await component.loadOperators();

    stub.listOperators = vi.fn(() => of([buildOperator('9')]));
    await component.onUpdated(buildOperator('9'));

    expect(notifications.success).toHaveBeenCalledWith(
      'Operador "Operator 9" actualizado correctamente',
    );
    expect(component.editOpen()).toBe(false);
    expect(component.operators()).toHaveLength(1);
  });

  it('hides the create button for non-admins', () => {
    isAdmin.set(false);
    const fixture = TestBed.createComponent(OperatorsPage);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).not.toContain('Crear operador');
  });
});
