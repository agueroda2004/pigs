import { TestBed } from '@angular/core/testing';

import { Breed } from '../../../core/breeds/breed.models';
import { BreedCard } from './breed-card';

function buildBreed(active = true): Breed {
  return {
    id: '1',
    name: 'Duroc',
    active,
    created_at: '2026-01-02T12:00:00',
    updated_at: '2026-01-02T12:00:00',
    created_by: 'admin',
    updated_by: 'admin',
  };
}

describe('BreedCard', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [BreedCard],
    }).compileComponents();
  });

  it('renders the name, status and created date', () => {
    const fixture = TestBed.createComponent(BreedCard);
    fixture.componentRef.setInput('breed', buildBreed(true));
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('Duroc');
    expect(text).toContain('Activa');
    expect(text).toContain('02/01/2026');
  });

  it('shows the inactive status for inactive breeds', () => {
    const fixture = TestBed.createComponent(BreedCard);
    fixture.componentRef.setInput('breed', buildBreed(false));
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('Inactiva');
  });

  it('hides the edit button when editing is not allowed', () => {
    const fixture = TestBed.createComponent(BreedCard);
    fixture.componentRef.setInput('breed', buildBreed());
    fixture.componentRef.setInput('canEdit', false);
    fixture.detectChanges();

    expect(fixture.nativeElement.querySelector('button')).toBeNull();
  });

  it('emits the breed when the edit button is clicked', () => {
    const breed = buildBreed();
    const fixture = TestBed.createComponent(BreedCard);
    fixture.componentRef.setInput('breed', breed);
    fixture.componentRef.setInput('canEdit', true);
    fixture.detectChanges();

    const emitted: Breed[] = [];
    fixture.componentInstance.editRequested.subscribe((value) => emitted.push(value));
    (fixture.nativeElement.querySelector('button') as HTMLButtonElement).click();

    expect(emitted).toEqual([breed]);
  });
});
