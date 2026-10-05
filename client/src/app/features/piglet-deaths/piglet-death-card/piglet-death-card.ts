import { Component, computed, input } from '@angular/core';

import { PigletDeath } from '../../../core/piglet-deaths/piglet-death.models';
import { PIGLET_DEATH_CAUSE_CLASSES, PIGLET_DEATH_CAUSE_LABELS } from '../piglet-death-cause';
import { PIGLET_DEATH_TURN_LABELS } from '../piglet-death-turn';

@Component({
  selector: 'app-piglet-death-card',
  styleUrl: './piglet-death-card.css',
  templateUrl: './piglet-death-card.html',
})
export class PigletDeathCard {
  readonly death = input.required<PigletDeath>();
  readonly sowCode = input<string | null>(null);

  protected readonly causeLabel = computed(
    () => PIGLET_DEATH_CAUSE_LABELS[this.death().cause],
  );
  protected readonly causeClass = computed(
    () => PIGLET_DEATH_CAUSE_CLASSES[this.death().cause],
  );
  protected readonly turnLabel = computed(() => PIGLET_DEATH_TURN_LABELS[this.death().turn]);
}
