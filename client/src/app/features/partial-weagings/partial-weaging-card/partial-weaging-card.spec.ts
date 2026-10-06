import { TestBed } from '@angular/core/testing';

import { PartialWeaging } from '../../../core/partial-weagings/partial-weaging.models';
import { PartialWeagingCard } from './partial-weaging-card';

function buildWeaging(overrides: Partial<PartialWeaging> = {}): PartialWeaging {
  return {
    id: '1',
    farrowing_id: 'farrowing-1',
    sow_id: 'sow-1',
    weaging_date: '2026-04-25',
    quantity: 2,
    total_weight: 42.5,
    type: 'Normal',
    note: null,
    created_at: '',
    updated_at: '',
    created_by: 'admin',
    updated_by: 'admin',
    ...overrides,
  };
}

describe('PartialWeagingCard', () => {
  const render = (weaging: PartialWeaging, sowCode: string | null) => {
    const fixture = TestBed.createComponent(PartialWeagingCard);
    fixture.componentRef.setInput('weaging', weaging);
    fixture.componentRef.setInput('sowCode', sowCode);
    fixture.detectChanges();
    return fixture;
  };

  beforeEach(() => TestBed.configureTestingModule({ imports: [PartialWeagingCard] }));

  it('renders the sow code, date, quantity and type', () => {
    const fixture = render(buildWeaging(), 'C-001');
    const text = fixture.nativeElement.textContent;

    expect(text).toContain('C-001');
    expect(text).toContain('2026-04-25');
    expect(text).toContain('Cantidad: 2');
    expect(text).toContain('Normal');
  });

  it('renders the type label for a nurse weaging', () => {
    const fixture = render(buildWeaging({ type: 'Nodriza' }), 'C-001');

    expect(fixture.nativeElement.textContent).toContain('Nodriza');
  });

  it('falls back when the sow code is unknown', () => {
    const fixture = render(buildWeaging(), null);

    expect(fixture.nativeElement.textContent).toContain('sin identificar');
  });

  it('shows the weight and note only when present', () => {
    const without = render(buildWeaging({ total_weight: null }), 'C-001');
    expect(without.nativeElement.textContent).not.toContain('Peso total');
    expect(without.nativeElement.textContent).not.toContain('Nota');

    const withValues = render(buildWeaging({ total_weight: 10, note: 'Camada numerosa' }), 'C-001');
    expect(withValues.nativeElement.textContent).toContain('Peso total: 10 kg');
    expect(withValues.nativeElement.textContent).toContain('Camada numerosa');
  });
});
