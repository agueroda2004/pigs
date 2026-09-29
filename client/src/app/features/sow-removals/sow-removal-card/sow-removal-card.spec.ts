import { TestBed } from '@angular/core/testing';
import { vi } from 'vitest';

import { SowRemoval } from '../../../core/sow-removals/sow-removal.models';
import { SowRemovalCard } from './sow-removal-card';

function buildRemoval(overrides: Partial<SowRemoval> = {}): SowRemoval {
  return {
    id: '1',
    sow_id: 'sow-1',
    removal_date: '2026-01-20',
    type: 'Muerte',
    reason: 'Enfermedad',
    note: null,
    last_state: 'Gestando',
    created_at: '',
    updated_at: '',
    created_by: 'admin',
    updated_by: 'admin',
    ...overrides,
  };
}

describe('SowRemovalCard', () => {
  const render = (removal: SowRemoval, sowCode: string | null) => {
    const fixture = TestBed.createComponent(SowRemovalCard);
    fixture.componentRef.setInput('removal', removal);
    fixture.componentRef.setInput('sowCode', sowCode);
    fixture.detectChanges();
    return fixture;
  };

  beforeEach(() => TestBed.configureTestingModule({ imports: [SowRemovalCard] }));

  it('renders the sow code, date, type, reason and last state', () => {
    const fixture = render(buildRemoval(), 'C-001');

    const text = fixture.nativeElement.textContent;
    expect(text).toContain('C-001');
    expect(text).toContain('2026-01-20');
    expect(text).toContain('Muerte');
    expect(text).toContain('Enfermedad');
    expect(text).toContain('Gestando');
  });

  it('falls back when the sow code is unknown', () => {
    const fixture = render(buildRemoval(), null);

    expect(fixture.nativeElement.textContent).toContain('sin identificar');
  });

  it('shows the note only when present', () => {
    const withoutNote = render(buildRemoval(), 'C-001');
    expect(withoutNote.nativeElement.textContent).not.toContain('Nota');

    const withNote = render(buildRemoval({ note: 'baja por enfermedad' }), 'C-001');
    expect(withNote.nativeElement.textContent).toContain('baja por enfermedad');
  });

  it('shows the edit button only when editable', () => {
    const readOnly = render(buildRemoval(), 'C-001');
    expect(readOnly.nativeElement.textContent).not.toContain('Editar');

    const fixture = TestBed.createComponent(SowRemovalCard);
    fixture.componentRef.setInput('removal', buildRemoval());
    fixture.componentRef.setInput('sowCode', 'C-001');
    fixture.componentRef.setInput('editable', true);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('Editar');
  });

  it('emits remove when the delete button is clicked', () => {
    const fixture = TestBed.createComponent(SowRemovalCard);
    fixture.componentRef.setInput('removal', buildRemoval());
    fixture.componentRef.setInput('sowCode', 'C-001');
    fixture.componentRef.setInput('editable', true);
    fixture.detectChanges();
    const emitted = vi.fn();
    fixture.componentInstance.remove.subscribe(emitted);

    const button = Array.from(
      fixture.nativeElement.querySelectorAll('button') as NodeListOf<HTMLButtonElement>,
    ).find((candidate) => candidate.textContent?.includes('Eliminar'));
    button?.click();

    expect(emitted).toHaveBeenCalled();
  });
});
