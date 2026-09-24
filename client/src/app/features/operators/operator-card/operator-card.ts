import { DatePipe } from '@angular/common';
import { Component, computed, input, output } from '@angular/core';

import { Operator } from '../../../core/operators/operator.models';

@Component({
  selector: 'app-operator-card',
  imports: [DatePipe],
  styleUrl: './operator-card.css',
  templateUrl: './operator-card.html',
})
export class OperatorCard {
  readonly operator = input.required<Operator>();
  readonly canEdit = input(false);
  readonly editRequested = output<Operator>();

  protected readonly statusLabel = computed(() => (this.operator().active ? 'Activo' : 'Inactivo'));

  protected readonly statusClass = computed(() =>
    this.operator().active ? 'bg-success/10 text-success' : 'bg-muted text-muted-foreground',
  );
}
