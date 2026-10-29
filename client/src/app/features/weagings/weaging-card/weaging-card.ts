import { Component, input, output } from '@angular/core';

import { Weaging } from '../../../core/weagings/weaging.models';

@Component({
  selector: 'app-weaging-card',
  styleUrl: './weaging-card.css',
  templateUrl: './weaging-card.html',
})
export class WeagingCard {
  readonly weaging = input.required<Weaging>();
  readonly sowCode = input<string | null>(null);
  readonly canEdit = input(false);
  readonly canDelete = input(false);
  readonly editRequested = output<Weaging>();
  readonly deleteRequested = output<Weaging>();
}
