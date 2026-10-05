import { Component, input } from '@angular/core';

import { PigletFostering } from '../../../core/piglet-fosterings/piglet-fostering.models';

@Component({
  selector: 'app-piglet-fostering-card',
  styleUrl: './piglet-fostering-card.css',
  templateUrl: './piglet-fostering-card.html',
})
export class PigletFosteringCard {
  readonly fostering = input.required<PigletFostering>();
  readonly donorSowCode = input<string | null>(null);
  readonly receiverSowCode = input<string | null>(null);
}
