import { Component, computed, input } from '@angular/core';

import { Farrowing } from '../../../core/farrowings/farrowing.models';

@Component({
  selector: 'app-farrowing-card',
  styleUrl: './farrowing-card.css',
  templateUrl: './farrowing-card.html',
})
export class FarrowingCard {
  readonly farrowing = input.required<Farrowing>();
  readonly sowCode = input<string | null>(null);

  protected readonly totalBorn = computed(
    () => this.farrowing().live_born + this.farrowing().stillborn + this.farrowing().mummified,
  );

  protected readonly timeRange = computed(() => {
    const { start_time, end_time } = this.farrowing();
    if (start_time && end_time) {
      return `${start_time} → ${end_time}`;
    }
    return start_time ?? end_time ?? null;
  });

  protected readonly duration = computed(() => {
    const { start_time, end_time } = this.farrowing();
    if (!start_time || !end_time) {
      return null;
    }
    const start = toMinutes(start_time);
    const end = toMinutes(end_time);
    if (start === null || end === null) {
      return null;
    }
    const diff = end >= start ? end - start : end + 24 * 60 - start;
    const hours = Math.floor(diff / 60);
    const minutes = diff % 60;
    return minutes === 0 ? `${hours} h` : `${hours} h ${minutes} min`;
  });
}

// toMinutes converts an HH:MM string into its total minutes.
// It returns null when the value does not match the HH:MM format.
function toMinutes(value: string): number | null {
  const match = /^([01]\d|2[0-3]):([0-5]\d)$/.exec(value);
  if (!match) {
    return null;
  }
  return Number(match[1]) * 60 + Number(match[2]);
}
