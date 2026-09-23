import { Component, computed, input, output } from '@angular/core';

import { Boar } from '../../../core/boars/boar.models';
import { BOAR_ORIGIN_CLASSES, BOAR_ORIGIN_LABELS } from '../boar-origin';
import { BOAR_STATE_CLASSES, BOAR_STATE_LABELS } from '../boar-state';

@Component({
  selector: 'app-boar-card',
  styleUrl: './boar-card.css',
  templateUrl: './boar-card.html',
})
export class BoarCard {
  readonly boar = input.required<Boar>();
  readonly breedName = input<string | null>(null);
  readonly canEdit = input(false);
  readonly editRequested = output<Boar>();

  protected readonly stateLabel = computed(() => BOAR_STATE_LABELS[this.boar().state]);
  protected readonly stateClass = computed(() => BOAR_STATE_CLASSES[this.boar().state]);
  protected readonly originLabel = computed(() => BOAR_ORIGIN_LABELS[this.boar().origin]);
  protected readonly originClass = computed(() => BOAR_ORIGIN_CLASSES[this.boar().origin]);
  protected readonly activeLabel = computed(() => (this.boar().active ? 'Activo' : 'Inactivo'));
  protected readonly activeClass = computed(() =>
    this.boar().active ? 'bg-success/10 text-success' : 'bg-muted text-muted-foreground',
  );
}
