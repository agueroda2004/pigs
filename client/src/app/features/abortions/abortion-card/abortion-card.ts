import { Component, computed, input, output } from '@angular/core';

import { Abortion } from '../../../core/abortions/abortion.models';
import { ABORTION_CAUSE_CLASSES, ABORTION_CAUSE_LABELS } from '../abortion-cause';

@Component({
  selector: 'app-abortion-card',
  styleUrl: './abortion-card.css',
  templateUrl: './abortion-card.html',
})
export class AbortionCard {
  readonly abortion = input.required<Abortion>();
  readonly sowCode = input<string | null>(null);
  readonly canEdit = input(false);
  readonly editRequested = output<Abortion>();
  readonly deleteRequested = output<Abortion>();

  protected readonly causeLabel = computed(() => ABORTION_CAUSE_LABELS[this.abortion().cause]);
  protected readonly causeClass = computed(() => ABORTION_CAUSE_CLASSES[this.abortion().cause]);
}
