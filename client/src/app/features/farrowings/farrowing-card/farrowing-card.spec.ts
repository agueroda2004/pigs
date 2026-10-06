import { TestBed } from '@angular/core/testing';

import { Farrowing } from '../../../core/farrowings/farrowing.models';
import { FarrowingCard } from './farrowing-card';

function buildFarrowing(overrides: Partial<Farrowing> = {}): Farrowing {
  return {
    id: '1',
    sow_id: 'sow-1',
    service_id: 'service-1',
    farrow_date: '2026-04-20',
    start_time: '22:00',
    end_time: '02:00',
    location: 'Corral 3',
    live_born: 10,
    stillborn: 1,
    mummified: 2,
    current_piglets: 8,
    litter_weight: 15.5,
    stillborn_weight: null,
    is_manipulated: false,
    is_nurse: false,
    nurse_start_date: null,
    note: null,
    operators: [],
    medications: [],
    created_at: '',
    updated_at: '',
    created_by: 'admin',
    updated_by: 'admin',
    ...overrides,
  };
}

describe('FarrowingCard', () => {
  const render = (farrowing: Farrowing, sowCode: string | null) => {
    const fixture = TestBed.createComponent(FarrowingCard);
    fixture.componentRef.setInput('farrowing', farrowing);
    fixture.componentRef.setInput('sowCode', sowCode);
    fixture.detectChanges();
    return fixture;
  };

  beforeEach(() => TestBed.configureTestingModule({ imports: [FarrowingCard] }));

  it('renders the sow code, date, counts and total born', () => {
    const fixture = render(buildFarrowing(), 'C-001');
    const text = fixture.nativeElement.textContent;

    expect(text).toContain('C-001');
    expect(text).toContain('2026-04-20');
    expect(text).toContain('10');
    expect(text).toContain('1');
    expect(text).toContain('2');
    expect(text).toContain('13');
  });

  it('renders the current piglets balance', () => {
    const fixture = render(buildFarrowing({ current_piglets: 7 }), 'C-001');
    const text = fixture.nativeElement.textContent;

    expect(text).toContain('Actuales');
    expect(text).toContain('7');
  });

  it('renders the time range and location', () => {
    const fixture = render(buildFarrowing(), 'C-001');
    const text = fixture.nativeElement.textContent;

    expect(text).toContain('22:00 → 02:00');
    expect(text).toContain('Corral 3');
  });

  it('computes the duration for a same-day range', () => {
    const fixture = render(buildFarrowing({ start_time: '08:00', end_time: '12:30' }), 'C-001');

    expect(fixture.nativeElement.textContent).toContain('Duración: 4 h 30 min');
  });

  it('computes the duration treating a range that crosses midnight', () => {
    const fixture = render(buildFarrowing({ start_time: '22:00', end_time: '02:00' }), 'C-001');

    expect(fixture.nativeElement.textContent).toContain('Duración: 4 h');
  });

  it('hides the duration when either time is missing', () => {
    const fixture = render(buildFarrowing({ start_time: '22:00', end_time: null }), 'C-001');

    expect(fixture.nativeElement.textContent).not.toContain('Duración');
  });

  it('falls back when the sow code is unknown', () => {
    const fixture = render(buildFarrowing(), null);

    expect(fixture.nativeElement.textContent).toContain('sin identificar');
  });

  it('shows the manipulated badge only when the farrowing is manipulated', () => {
    const without = render(buildFarrowing(), 'C-001');
    expect(without.nativeElement.textContent).not.toContain('Manipulado');

    const withBadge = render(buildFarrowing({ is_manipulated: true }), 'C-001');
    expect(withBadge.nativeElement.textContent).toContain('Manipulado');
  });

  it('shows the nurse badge and start date only for a nurse farrowing', () => {
    const without = render(buildFarrowing(), 'C-001');
    expect(without.nativeElement.textContent).not.toContain('Nodriza');

    const withNurse = render(
      buildFarrowing({ is_nurse: true, nurse_start_date: '2026-04-25' }),
      'C-001',
    );
    expect(withNurse.nativeElement.textContent).toContain('Nodriza');
    expect(withNurse.nativeElement.textContent).toContain('Inicio nodriza: 2026-04-25');
  });

  it('shows optional weights and note only when present', () => {
    const without = render(buildFarrowing(), 'C-001');
    expect(without.nativeElement.textContent).not.toContain('Nota');

    const withExtras = render(
      buildFarrowing({ stillborn_weight: 2.5, note: 'Parto sin complicaciones' }),
      'C-001',
    );
    expect(withExtras.nativeElement.textContent).toContain('Peso nacidos muertos: 2.5 kg');
    expect(withExtras.nativeElement.textContent).toContain('Parto sin complicaciones');
  });
});
