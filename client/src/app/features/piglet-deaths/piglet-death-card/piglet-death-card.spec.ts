import { TestBed } from '@angular/core/testing';
import { vi } from 'vitest';

import { PigletDeath } from '../../../core/piglet-deaths/piglet-death.models';
import { PigletDeathCard } from './piglet-death-card';

function buildDeath(overrides: Partial<PigletDeath> = {}): PigletDeath {
  return {
    id: '1',
    farrowing_id: 'farrowing-1',
    sow_id: 'sow-1',
    sow_code: 'C-001',
    operator_id: 'operator-1',
    operator_name: 'Juan Pérez',
    death_date: '2026-04-25',
    quantity: 2,
    weight: 2.5,
    cause: 'Pata_abierta',
    turn: 'Madrugada_no_asistida',
    note: null,
    ...overrides,
  };
}

describe('PigletDeathCard', () => {
  const render = (
    death: PigletDeath,
    sowCode: string | null,
    canDelete = false,
    canEdit = false,
  ) => {
    const fixture = TestBed.createComponent(PigletDeathCard);
    fixture.componentRef.setInput('death', death);
    fixture.componentRef.setInput('sowCode', sowCode);
    fixture.componentRef.setInput('canDelete', canDelete);
    fixture.componentRef.setInput('canEdit', canEdit);
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

  it('renders the Sin Turno label', () => {
    const fixture = render(buildDeath({ turn: 'Sin Turno' }), 'C-001');

    expect(fixture.nativeElement.textContent).toContain('Turno: Sin Turno');
  });

  it('renders the Desaparecido cause label', () => {
    const fixture = render(buildDeath({ cause: 'Desaparecido' }), 'C-001');

    expect(fixture.nativeElement.textContent).toContain('Desaparecido');
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

  it('hides the delete action when the user cannot delete', () => {
    const fixture = render(buildDeath(), 'C-001');

    expect(fixture.nativeElement.querySelectorAll('button')).toHaveLength(0);
  });

  it('emits the death when the delete action is used', () => {
    const death = buildDeath();
    const fixture = render(death, 'C-001', true);
    const emitted = vi.fn();
    fixture.componentInstance.deleteRequested.subscribe(emitted);

    const button = Array.from<HTMLButtonElement>(
      fixture.nativeElement.querySelectorAll('button'),
    ).find((element) => element.textContent?.includes('Eliminar'));
    button?.click();

    expect(emitted).toHaveBeenCalledWith(death);
  });

  it('emits the death when the edit action is used', () => {
    const death = buildDeath();
    const fixture = render(death, 'C-001', false, true);
    const emitted = vi.fn();
    fixture.componentInstance.editRequested.subscribe(emitted);

    const button = Array.from<HTMLButtonElement>(
      fixture.nativeElement.querySelectorAll('button'),
    ).find((element) => element.textContent?.includes('Editar'));
    button?.click();

    expect(emitted).toHaveBeenCalledWith(death);
  });
});
