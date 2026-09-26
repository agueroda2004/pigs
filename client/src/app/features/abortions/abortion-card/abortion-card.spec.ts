import { TestBed } from '@angular/core/testing';

import { Abortion } from '../../../core/abortions/abortion.models';
import { AbortionCard } from './abortion-card';

function buildAbortion(overrides: Partial<Abortion> = {}): Abortion {
  return {
    id: '1',
    sow_id: 'sow-1',
    service_id: 'service-1',
    abortion_date: '2026-01-15',
    cause: 'Infeccioso',
    note: null,
    created_at: '',
    updated_at: '',
    created_by: 'admin',
    updated_by: 'admin',
    ...overrides,
  };
}

describe('AbortionCard', () => {
  const render = (abortion: Abortion, sowCode: string | null) => {
    const fixture = TestBed.createComponent(AbortionCard);
    fixture.componentRef.setInput('abortion', abortion);
    fixture.componentRef.setInput('sowCode', sowCode);
    fixture.detectChanges();
    return fixture;
  };

  beforeEach(() => TestBed.configureTestingModule({ imports: [AbortionCard] }));

  it('renders the sow code, date and cause', () => {
    const fixture = render(buildAbortion(), 'C-001');

    const text = fixture.nativeElement.textContent;
    expect(text).toContain('C-001');
    expect(text).toContain('2026-01-15');
    expect(text).toContain('Infeccioso');
  });

  it('falls back when the sow code is unknown', () => {
    const fixture = render(buildAbortion(), null);

    expect(fixture.nativeElement.textContent).toContain('sin identificar');
  });

  it('shows the note only when present', () => {
    const withoutNote = render(buildAbortion(), 'C-001');
    expect(withoutNote.nativeElement.textContent).not.toContain('Nota');

    const withNote = render(buildAbortion({ note: 'aborto espontáneo' }), 'C-001');
    expect(withNote.nativeElement.textContent).toContain('aborto espontáneo');
  });
});
