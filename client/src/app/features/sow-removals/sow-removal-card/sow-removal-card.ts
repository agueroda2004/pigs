import { Component, computed, input } from '@angular/core';

import { SowRemoval } from '../../../core/sow-removals/sow-removal.models';
import { SOW_STATE_LABELS } from '../../sows/sow-state';
import { REMOVAL_REASON_LABELS } from '../removal-reason';
import { REMOVAL_TYPE_CLASSES, REMOVAL_TYPE_LABELS } from '../removal-type';

@Component({
  selector: 'app-sow-removal-card',
  styleUrl: './sow-removal-card.css',
  templateUrl: './sow-removal-card.html',
})
export class SowRemovalCard {
  readonly removal = input.required<SowRemoval>();
  readonly sowCode = input<string | null>(null);

  protected readonly typeLabel = computed(() => REMOVAL_TYPE_LABELS[this.removal().type]);
  protected readonly typeClass = computed(() => REMOVAL_TYPE_CLASSES[this.removal().type]);
  protected readonly reasonLabel = computed(() => REMOVAL_REASON_LABELS[this.removal().reason]);
  protected readonly lastStateLabel = computed(() => SOW_STATE_LABELS[this.removal().last_state]);
}
