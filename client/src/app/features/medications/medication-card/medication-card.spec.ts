import { TestBed } from '@angular/core/testing';

import { Medication } from '../../../core/medications/medication.models';
import { MedicationCard } from './medication-card';

function buildMedication(active = true): Medication {
  return {
    id: '1',
    name: 'Ivermectina',
    active,
    created_at: '2026-01-02T12:00:00',
    updated_at: '2026-01-02T12:00:00',
    created_by: 'admin',
    updated_by: 'admin',
  };
}

describe('MedicationCard', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [MedicationCard],
    }).compileComponents();
  });

  it('renders the name, status and created date', () => {
    const fixture = TestBed.createComponent(MedicationCard);
    fixture.componentRef.setInput('medication', buildMedication(true));
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('Ivermectina');
    expect(text).toContain('Activo');
    expect(text).toContain('02/01/2026');
  });

  it('shows the inactive status for inactive medications', () => {
    const fixture = TestBed.createComponent(MedicationCard);
    fixture.componentRef.setInput('medication', buildMedication(false));
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('Inactivo');
  });

  it('hides the edit button when editing is not allowed', () => {
    const fixture = TestBed.createComponent(MedicationCard);
    fixture.componentRef.setInput('medication', buildMedication());
    fixture.componentRef.setInput('canEdit', false);
    fixture.detectChanges();

    expect(fixture.nativeElement.querySelector('button')).toBeNull();
  });

  it('emits the medication when the edit button is clicked', () => {
    const medication = buildMedication();
    const fixture = TestBed.createComponent(MedicationCard);
    fixture.componentRef.setInput('medication', medication);
    fixture.componentRef.setInput('canEdit', true);
    fixture.detectChanges();

    const emitted: Medication[] = [];
    fixture.componentInstance.editRequested.subscribe((value) => emitted.push(value));
    (fixture.nativeElement.querySelector('button') as HTMLButtonElement).click();

    expect(emitted).toEqual([medication]);
  });
});
