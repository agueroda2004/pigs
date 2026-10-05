import { TestBed } from '@angular/core/testing';

import { PigletFostering } from '../../../core/piglet-fosterings/piglet-fostering.models';
import { PigletFosteringCard } from './piglet-fostering-card';

function buildFostering(overrides: Partial<PigletFostering> = {}): PigletFostering {
  return {
    id: '1',
    donor_farrowing_id: 'farrowing-1',
    receiver_farrowing_id: 'farrowing-2',
    donor_sow_id: 'sow-1',
    receiver_sow_id: 'sow-2',
    movement_date: '2026-04-25',
    quantity: 2,
    note: null,
    created_at: '',
    updated_at: '',
    created_by: 'admin',
    updated_by: 'admin',
    ...overrides,
  };
}

describe('PigletFosteringCard', () => {
  const render = (fostering: PigletFostering, donor: string | null, receiver: string | null) => {
    const fixture = TestBed.createComponent(PigletFosteringCard);
    fixture.componentRef.setInput('fostering', fostering);
    fixture.componentRef.setInput('donorSowCode', donor);
    fixture.componentRef.setInput('receiverSowCode', receiver);
    fixture.detectChanges();
    return fixture;
  };

  beforeEach(() => TestBed.configureTestingModule({ imports: [PigletFosteringCard] }));

  it('renders the sow codes, date and quantity', () => {
    const fixture = render(buildFostering(), 'C-001', 'C-002');
    const text = fixture.nativeElement.textContent;

    expect(text).toContain('C-001');
    expect(text).toContain('C-002');
    expect(text).toContain('2026-04-25');
    expect(text).toContain('Cantidad: 2');
  });

  it('falls back when a sow code is unknown', () => {
    const fixture = render(buildFostering(), null, null);

    expect(fixture.nativeElement.textContent).toContain('sin identificar');
  });

  it('shows the note only when present', () => {
    const without = render(buildFostering(), 'C-001', 'C-002');
    expect(without.nativeElement.textContent).not.toContain('Nota');

    const withNote = render(buildFostering({ note: 'Camada numerosa' }), 'C-001', 'C-002');
    expect(withNote.nativeElement.textContent).toContain('Camada numerosa');
  });
});
