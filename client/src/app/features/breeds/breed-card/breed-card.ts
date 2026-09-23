import { DatePipe } from '@angular/common';
import { Component, computed, input, output } from '@angular/core';

import { Breed } from '../../../core/breeds/breed.models';

@Component({
  selector: 'app-breed-card',
  imports: [DatePipe],
  styleUrl: './breed-card.css',
  templateUrl: './breed-card.html',
})
export class BreedCard {
  readonly breed = input.required<Breed>();
  readonly canEdit = input(false);
  readonly editRequested = output<Breed>();

  protected readonly statusLabel = computed(() => (this.breed().active ? 'Activa' : 'Inactiva'));

  protected readonly statusClass = computed(() =>
    this.breed().active ? 'bg-success/10 text-success' : 'bg-muted text-muted-foreground',
  );
}
