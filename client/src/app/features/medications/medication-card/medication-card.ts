import { DatePipe } from '@angular/common';
import { Component, computed, input, output } from '@angular/core';

import { Medication } from '../../../core/medications/medication.models';

@Component({
  selector: 'app-medication-card',
  imports: [DatePipe],
  styleUrl: './medication-card.css',
  templateUrl: './medication-card.html',
})
export class MedicationCard {
  readonly medication = input.required<Medication>();
  readonly canEdit = input(false);
  readonly editRequested = output<Medication>();

  protected readonly statusLabel = computed(() =>
    this.medication().active ? 'Activo' : 'Inactivo',
  );

  protected readonly statusClass = computed(() =>
    this.medication().active ? 'bg-success/10 text-success' : 'bg-muted text-muted-foreground',
  );
}
