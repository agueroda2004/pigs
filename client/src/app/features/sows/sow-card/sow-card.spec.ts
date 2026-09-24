import { TestBed } from '@angular/core/testing';

import { Sow } from '../../../core/sows/sow.models';
import { SowCard } from './sow-card';

function buildSow(overrides: Partial<Sow> = {}): Sow {
  return {
    id: '1',
    code: 'C-001',
    location: 'Corral A',
    active: true,
    entry_date: '2026-01-10',
    birth_date: null,
    note: null,
    state: 'Viva',
    origin: 'Propio',
    parity: 3,
    breed_id: 'breed-1',
    created_at: '2026-01-02T12:00:00',
    updated_at: '2026-01-02T12:00:00',
    created_by: 'admin',
    updated_by: 'admin',
    ...overrides,
  };
}

describe('SowCard', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [SowCard] }).compileComponents();
  });

  it('renders the code, state, origin, active status, parity, breed and location', () => {
    const fixture = TestBed.createComponent(SowCard);
    fixture.componentRef.setInput('sow', buildSow());
    fixture.componentRef.setInput('breedName', 'Duroc');
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('C-001');
    expect(text).toContain('Viva');
    expect(text).toContain('Propio');
    expect(text).toContain('Activa');
    expect(text).toContain('Paridad: 3');
    expect(text).toContain('Duroc');
    expect(text).toContain('Corral A');
  });

  it('maps the state to its badge class', () => {
    const fixture = TestBed.createComponent(SowCard);
    fixture.componentRef.setInput('sow', buildSow({ state: 'Muerta' }));
    fixture.detectChanges();

    expect(fixture.nativeElement.querySelector('.text-danger')).not.toBeNull();
  });

  it('maps the external origin to its badge class', () => {
    const fixture = TestBed.createComponent(SowCard);
    fixture.componentRef.setInput('sow', buildSow({ origin: 'Externo' }));
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('Externo');
    expect(fixture.nativeElement.querySelector('.text-info')).not.toBeNull();
  });

  it('shows placeholders for missing breed and location', () => {
    const fixture = TestBed.createComponent(SowCard);
    fixture.componentRef.setInput('sow', buildSow({ location: null }));
    fixture.componentRef.setInput('breedName', null);
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('Sin raza');
    expect(text).toContain('Sin ubicación');
  });

  it('hides the edit button when editing is not allowed', () => {
    const fixture = TestBed.createComponent(SowCard);
    fixture.componentRef.setInput('sow', buildSow());
    fixture.componentRef.setInput('canEdit', false);
    fixture.detectChanges();

    expect(fixture.nativeElement.querySelector('button')).toBeNull();
  });

  it('emits the sow when the edit button is clicked', () => {
    const sow = buildSow();
    const fixture = TestBed.createComponent(SowCard);
    fixture.componentRef.setInput('sow', sow);
    fixture.componentRef.setInput('canEdit', true);
    fixture.detectChanges();

    const emitted: Sow[] = [];
    fixture.componentInstance.editRequested.subscribe((value) => emitted.push(value));
    (fixture.nativeElement.querySelector('button') as HTMLButtonElement).click();

    expect(emitted).toEqual([sow]);
  });
});
