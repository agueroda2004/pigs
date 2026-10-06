import { TestBed } from '@angular/core/testing';

import { Weaging } from '../../../core/weagings/weaging.models';
import { WeagingCard } from './weaging-card';

function buildWeaging(overrides: Partial<Weaging> = {}): Weaging {
  return {
    id: '1',
    farrowing_id: 'farrowing-1',
    sow_id: 'sow-1',
    weaging_date: '2026-04-25',
    quantity: 8,
    total_weight: 120.5,
    destination: 'Nave 2',
    note: null,
    created_at: '',
    updated_at: '',
    created_by: 'admin',
    updated_by: 'admin',
    ...overrides,
  };
}

describe('WeagingCard', () => {
  const render = (weaging: Weaging, sowCode: string | null) => {
    const fixture = TestBed.createComponent(WeagingCard);
    fixture.componentRef.setInput('weaging', weaging);
    fixture.componentRef.setInput('sowCode', sowCode);
    fixture.detectChanges();
    return fixture;
  };

  beforeEach(() => TestBed.configureTestingModule({ imports: [WeagingCard] }));

  it('renders the sow code, date, quantity, weight and destination', () => {
    const fixture = render(buildWeaging(), 'C-001');
    const text = fixture.nativeElement.textContent;

    expect(text).toContain('C-001');
    expect(text).toContain('2026-04-25');
    expect(text).toContain('Cantidad: 8');
    expect(text).toContain('Peso total: 120.5 kg');
    expect(text).toContain('Destino: Nave 2');
  });

  it('falls back when the sow code is unknown', () => {
    const fixture = render(buildWeaging(), null);

    expect(fixture.nativeElement.textContent).toContain('sin identificar');
  });

  it('shows the weight, destination and note only when present', () => {
    const without = render(buildWeaging({ total_weight: null, destination: null }), 'C-001');
    expect(without.nativeElement.textContent).not.toContain('Peso total');
    expect(without.nativeElement.textContent).not.toContain('Destino');
    expect(without.nativeElement.textContent).not.toContain('Nota');

    const withExtras = render(
      buildWeaging({ total_weight: 10, destination: 'Nave 1', note: 'Camada numerosa' }),
      'C-001',
    );
    expect(withExtras.nativeElement.textContent).toContain('Peso total: 10 kg');
    expect(withExtras.nativeElement.textContent).toContain('Destino: Nave 1');
    expect(withExtras.nativeElement.textContent).toContain('Camada numerosa');
  });
});
