import { Component, computed, input, output } from '@angular/core';

import { BoarRemoval } from '../../../core/boar-removals/boar-removal.models';
import { BOAR_STATE_LABELS } from '../../boars/boar-state';
import { REMOVAL_REASON_LABELS } from '../removal-reason';
import { REMOVAL_TYPE_CLASSES, REMOVAL_TYPE_LABELS } from '../removal-type';

@Component({
  selector: 'app-boar-removal-card',
  styleUrl: './boar-removal-card.css',
  templateUrl: './boar-removal-card.html',
})
export class BoarRemovalCard {
  readonly removal = input.required<BoarRemoval>();
  readonly boarCode = input<string | null>(null);
  readonly editable = input(false);
  readonly edit = output<void>();
  readonly remove = output<void>();

  protected readonly typeLabel = computed(() => REMOVAL_TYPE_LABELS[this.removal().type]);
  protected readonly typeClass = computed(() => REMOVAL_TYPE_CLASSES[this.removal().type]);
  protected readonly reasonLabel = computed(() => REMOVAL_REASON_LABELS[this.removal().reason]);
  protected readonly lastStateLabel = computed(() => BOAR_STATE_LABELS[this.removal().last_state]);
}
