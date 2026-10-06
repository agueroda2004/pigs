import { TestBed } from '@angular/core/testing';

import { Boar } from '../../../core/boars/boar.models';
import { BoarCard } from './boar-card';

function buildBoar(overrides: Partial<Boar> = {}): Boar {
  return {
    id: '1',
    code: 'B-001',
    location: 'Corral A',
    active: true,
    entry_date: '2026-01-10',
    birth_date: null,
    note: null,
    state: 'Vivo',
    origin: 'Propio',
    breed_id: 'breed-1',
    ...overrides,
  };
}

describe('BoarCard', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [BoarCard] }).compileComponents();
  });

  it('renders the code, state, origin, active status, breed and location', () => {
    const fixture = TestBed.createComponent(BoarCard);
    fixture.componentRef.setInput('boar', buildBoar());
    fixture.componentRef.setInput('breedName', 'Duroc');
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('B-001');
    expect(text).toContain('Vivo');
    expect(text).toContain('Propio');
    expect(text).toContain('Activo');
    expect(text).toContain('Duroc');
    expect(text).toContain('Corral A');
  });

  it('maps the state to its badge class', () => {
    const fixture = TestBed.createComponent(BoarCard);
    fixture.componentRef.setInput('boar', buildBoar({ state: 'Muerto' }));
    fixture.detectChanges();

    expect(fixture.nativeElement.querySelector('.text-danger')).not.toBeNull();
  });

  it('maps the external origin to its badge class', () => {
    const fixture = TestBed.createComponent(BoarCard);
    fixture.componentRef.setInput('boar', buildBoar({ origin: 'Externo' }));
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('Externo');
    expect(fixture.nativeElement.querySelector('.text-info')).not.toBeNull();
  });

  it('shows placeholders for missing breed and location', () => {
    const fixture = TestBed.createComponent(BoarCard);
    fixture.componentRef.setInput('boar', buildBoar({ location: null }));
    fixture.componentRef.setInput('breedName', null);
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('Sin raza');
    expect(text).toContain('Sin ubicación');
  });

  it('hides the edit button when editing is not allowed', () => {
    const fixture = TestBed.createComponent(BoarCard);
    fixture.componentRef.setInput('boar', buildBoar());
    fixture.componentRef.setInput('canEdit', false);
    fixture.detectChanges();

    expect(fixture.nativeElement.querySelector('button')).toBeNull();
  });

  it('emits the boar when the edit button is clicked', () => {
    const boar = buildBoar();
    const fixture = TestBed.createComponent(BoarCard);
    fixture.componentRef.setInput('boar', boar);
    fixture.componentRef.setInput('canEdit', true);
    fixture.detectChanges();

    const emitted: Boar[] = [];
    fixture.componentInstance.editRequested.subscribe((value) => emitted.push(value));
    (fixture.nativeElement.querySelector('button') as HTMLButtonElement).click();

    expect(emitted).toEqual([boar]);
  });

  it('emits the boar when the delete button is clicked', () => {
    const boar = buildBoar();
    const fixture = TestBed.createComponent(BoarCard);
    fixture.componentRef.setInput('boar', boar);
    fixture.componentRef.setInput('canEdit', true);
    fixture.detectChanges();

    const emitted: Boar[] = [];
    fixture.componentInstance.deleteRequested.subscribe((value) => emitted.push(value));
    const buttons = fixture.nativeElement.querySelectorAll('button');
    (buttons[buttons.length - 1] as HTMLButtonElement).click();

    expect(emitted).toEqual([boar]);
  });
});
