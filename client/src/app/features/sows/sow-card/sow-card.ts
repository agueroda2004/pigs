import { Component, computed, input, output } from '@angular/core';

import { Sow } from '../../../core/sows/sow.models';
import { SOW_ORIGIN_CLASSES, SOW_ORIGIN_LABELS } from '../sow-origin';
import { SOW_STATE_CLASSES, SOW_STATE_LABELS } from '../sow-state';

@Component({
  selector: 'app-sow-card',
  styleUrl: './sow-card.css',
  templateUrl: './sow-card.html',
})
export class SowCard {
  readonly sow = input.required<Sow>();
  readonly breedName = input<string | null>(null);
  readonly canEdit = input(false);
  readonly editRequested = output<Sow>();

  protected readonly stateLabel = computed(() => SOW_STATE_LABELS[this.sow().state]);
  protected readonly stateClass = computed(() => SOW_STATE_CLASSES[this.sow().state]);
  protected readonly originLabel = computed(() => SOW_ORIGIN_LABELS[this.sow().origin]);
  protected readonly originClass = computed(() => SOW_ORIGIN_CLASSES[this.sow().origin]);
  protected readonly activeLabel = computed(() => (this.sow().active ? 'Activa' : 'Inactiva'));
  protected readonly activeClass = computed(() =>
    this.sow().active ? 'bg-success/10 text-success' : 'bg-muted text-muted-foreground',
  );
}
