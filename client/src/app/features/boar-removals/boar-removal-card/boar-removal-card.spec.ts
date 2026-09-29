import { TestBed } from '@angular/core/testing';
import { vi } from 'vitest';

import { BoarRemoval } from '../../../core/boar-removals/boar-removal.models';
import { BoarRemovalCard } from './boar-removal-card';

function buildRemoval(overrides: Partial<BoarRemoval> = {}): BoarRemoval {
  return {
    id: '1',
    boar_id: 'boar-1',
    removal_date: '2026-01-20',
    type: 'Muerte',
    reason: 'Enfermedad',
    note: null,
    last_state: 'Vivo',
    created_at: '',
    updated_at: '',
    created_by: 'admin',
    updated_by: 'admin',
    ...overrides,
  };
}

describe('BoarRemovalCard', () => {
  const render = (removal: BoarRemoval, boarCode: string | null) => {
    const fixture = TestBed.createComponent(BoarRemovalCard);
    fixture.componentRef.setInput('removal', removal);
    fixture.componentRef.setInput('boarCode', boarCode);
    fixture.detectChanges();
    return fixture;
  };

  beforeEach(() => TestBed.configureTestingModule({ imports: [BoarRemovalCard] }));

  it('renders the boar code, date, type, reason and last state', () => {
    const fixture = render(buildRemoval(), 'B-001');

    const text = fixture.nativeElement.textContent;
    expect(text).toContain('B-001');
    expect(text).toContain('2026-01-20');
    expect(text).toContain('Muerte');
    expect(text).toContain('Enfermedad');
    expect(text).toContain('Vivo');
  });

  it('falls back when the boar code is unknown', () => {
    const fixture = render(buildRemoval(), null);

    expect(fixture.nativeElement.textContent).toContain('sin identificar');
  });

  it('shows the note only when present', () => {
    const withoutNote = render(buildRemoval(), 'B-001');
    expect(withoutNote.nativeElement.textContent).not.toContain('Nota');

    const withNote = render(buildRemoval({ note: 'baja por enfermedad' }), 'B-001');
    expect(withNote.nativeElement.textContent).toContain('baja por enfermedad');
  });

  it('shows the edit button only when editable', () => {
    const readOnly = render(buildRemoval(), 'B-001');
    expect(readOnly.nativeElement.textContent).not.toContain('Editar');
    expect(readOnly.nativeElement.textContent).not.toContain('Eliminar');

    const fixture = TestBed.createComponent(BoarRemovalCard);
    fixture.componentRef.setInput('removal', buildRemoval());
    fixture.componentRef.setInput('boarCode', 'B-001');
    fixture.componentRef.setInput('editable', true);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('Editar');
    expect(fixture.nativeElement.textContent).toContain('Eliminar');
  });

  it('emits edit when the edit button is clicked', () => {
    const fixture = TestBed.createComponent(BoarRemovalCard);
    fixture.componentRef.setInput('removal', buildRemoval());
    fixture.componentRef.setInput('boarCode', 'B-001');
    fixture.componentRef.setInput('editable', true);
    fixture.detectChanges();
    const emitted = vi.fn();
    fixture.componentInstance.edit.subscribe(emitted);

    const button = fixture.nativeElement.querySelector('button') as HTMLButtonElement;
    button.click();

    expect(emitted).toHaveBeenCalled();
  });

  it('emits remove when the delete button is clicked', () => {
    const fixture = TestBed.createComponent(BoarRemovalCard);
    fixture.componentRef.setInput('removal', buildRemoval());
    fixture.componentRef.setInput('boarCode', 'B-001');
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
