import { TestBed } from '@angular/core/testing';

import { Breed } from '../../../core/breeds/breed.models';
import { BreedCard } from './breed-card';

function buildBreed(active = true): Breed {
  return {
    id: '1',
    name: 'Duroc',
    active,
  };
}

describe('BreedCard', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [BreedCard],
    }).compileComponents();
  });

  it('renders the name and status', () => {
    const fixture = TestBed.createComponent(BreedCard);
    fixture.componentRef.setInput('breed', buildBreed(true));
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('Duroc');
    expect(text).toContain('Activa');
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

  it('emits the breed when the delete button is clicked', () => {
    const breed = buildBreed();
    const fixture = TestBed.createComponent(BreedCard);
    fixture.componentRef.setInput('breed', breed);
    fixture.componentRef.setInput('canEdit', true);
    fixture.detectChanges();

    const emitted: Breed[] = [];
    fixture.componentInstance.deleteRequested.subscribe((value) => emitted.push(value));
    const buttons = fixture.nativeElement.querySelectorAll('button');
    (buttons[buttons.length - 1] as HTMLButtonElement).click();

    expect(emitted).toEqual([breed]);
  });
});
