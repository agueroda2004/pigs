import { TestBed } from '@angular/core/testing';

import { Service } from '../../../core/services/service.models';
import { ServiceCard } from './service-card';

function buildService(overrides: Partial<Service> = {}): Service {
  return {
    id: '1',
    sow_id: 'sow-1',
    expected_farrowing_date: '2025-05-04',
    note: null,
    state: 'Confirmado',
    location: 'Nave 1',
    mounts: [
      {
        id: 'm1',
        service_id: '1',
        boar_id: 'boar-1',
        operator_id: 'operator-1',
        mount_number: 1,
        mount_date: '2025-01-10',
        type: 'Artificial',
        note: null,
        created_at: '2025-01-10T12:00:00',
        updated_at: '2025-01-10T12:00:00',
        created_by: 'admin',
        updated_by: 'admin',
      },
    ],
    created_at: '2025-01-10T12:00:00',
    updated_at: '2025-01-10T12:00:00',
    created_by: 'admin',
    updated_by: 'admin',
    ...overrides,
  };
}

describe('ServiceCard', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [ServiceCard] }).compileComponents();
  });

  it('renders the sow code, state, farrowing date and mounts', () => {
    const fixture = TestBed.createComponent(ServiceCard);
    fixture.componentRef.setInput('service', buildService());
    fixture.componentRef.setInput('sowCode', 'C-001');
    fixture.componentRef.setInput('boarCodes', new Map([['boar-1', 'V-001']]));
    fixture.componentRef.setInput('operatorNames', new Map([['operator-1', 'Ana']]));
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('C-001');
    expect(text).toContain('Confirmado');
    expect(text).toContain('2025-05-04');
    expect(text).toContain('Nave 1');
    expect(text).toContain('Monta 1');
    expect(text).toContain('V-001');
    expect(text).toContain('Ana');
    expect(text).toContain('Artificial');
  });

  it('maps the state to its badge class', () => {
    const fixture = TestBed.createComponent(ServiceCard);
    fixture.componentRef.setInput('service', buildService({ state: 'Fallido' }));
    fixture.detectChanges();

    expect(fixture.nativeElement.querySelector('.text-danger')).not.toBeNull();
  });

  it('shows placeholders when lookups are missing', () => {
    const fixture = TestBed.createComponent(ServiceCard);
    fixture.componentRef.setInput('service', buildService());
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('sin identificar');
    expect(text).toContain('Verraco');
    expect(text).toContain('Operador');
  });
});
