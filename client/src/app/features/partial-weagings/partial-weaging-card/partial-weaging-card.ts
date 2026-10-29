import { Component, computed, input, output } from '@angular/core';

import { PartialWeaging } from '../../../core/partial-weagings/partial-weaging.models';
import { PARTIAL_WEAGING_TYPE_CLASSES, PARTIAL_WEAGING_TYPE_LABELS } from '../partial-weaging-type';

@Component({
  selector: 'app-partial-weaging-card',
  styleUrl: './partial-weaging-card.css',
  templateUrl: './partial-weaging-card.html',
})
export class PartialWeagingCard {
  readonly weaging = input.required<PartialWeaging>();
  readonly sowCode = input<string | null>(null);
  readonly canEdit = input(false);
  readonly canDelete = input(false);
  readonly editRequested = output<PartialWeaging>();
  readonly deleteRequested = output<PartialWeaging>();

  protected readonly typeLabel = computed(() => PARTIAL_WEAGING_TYPE_LABELS[this.weaging().type]);
  protected readonly typeClass = computed(() => PARTIAL_WEAGING_TYPE_CLASSES[this.weaging().type]);
}
