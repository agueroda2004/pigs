import { Component, input } from '@angular/core';

import { Weaging } from '../../../core/weagings/weaging.models';

@Component({
  selector: 'app-weaging-card',
  styleUrl: './weaging-card.css',
  templateUrl: './weaging-card.html',
})
export class WeagingCard {
  readonly weaging = input.required<Weaging>();
  readonly sowCode = input<string | null>(null);
}
