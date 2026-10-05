import { TestBed } from '@angular/core/testing';

import { PigletDeath } from '../../../core/piglet-deaths/piglet-death.models';
import { PigletDeathCard } from './piglet-death-card';

function buildDeath(overrides: Partial<PigletDeath> = {}): PigletDeath {
  return {
    id: '1',
    farrowing_id: 'farrowing-1',
    sow_id: 'sow-1',
    operator_id: 'operator-1',
    operator_name: 'Juan Pérez',
    death_date: '2026-04-25',
    quantity: 2,
    weight: 2.5,
    cause: 'Pata_abierta',
    turn: 'Madrugada_no_asistida',
    note: null,
    created_at: '',
    updated_at: '',
    created_by: 'admin',
    updated_by: 'admin',
    ...overrides,
  };
}

describe('PigletDeathCard', () => {
  const render = (death: PigletDeath, sowCode: string | null) => {
    const fixture = TestBed.createComponent(PigletDeathCard);
    fixture.componentRef.setInput('death', death);
    fixture.componentRef.setInput('sowCode', sowCode);
    fixture.detectChanges();
    return fixture;
  };

  beforeEach(() => TestBed.configureTestingModule({ imports: [PigletDeathCard] }));

  it('renders the sow code, date, quantity, operator and labels', () => {
    const fixture = render(buildDeath(), 'C-001');
    const text = fixture.nativeElement.textContent;

    expect(text).toContain('Cerda C-001');
    expect(text).toContain('2026-04-25');
    expect(text).toContain('Cantidad: 2');
    expect(text).toContain('Peso: 2.5 kg');
    expect(text).toContain('Operador: Juan Pérez');
    expect(text).toContain('Pata abierta');
    expect(text).toContain('Madrugada no asistida');
  });

  it('falls back when the sow code is unknown', () => {
    const fixture = render(buildDeath(), null);

    expect(fixture.nativeElement.textContent).toContain('sin identificar');
  });

  it('hides the weight when it is null', () => {
    const fixture = render(buildDeath({ weight: null }), 'C-001');

    expect(fixture.nativeElement.textContent).not.toContain('Peso');
  });

  it('shows the note only when present', () => {
    const without = render(buildDeath(), 'C-001');
    expect(without.nativeElement.textContent).not.toContain('Nota');

    const withNote = render(buildDeath({ note: 'Sin complicaciones' }), 'C-001');
    expect(withNote.nativeElement.textContent).toContain('Sin complicaciones');
  });
});
