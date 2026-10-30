import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { AuthService } from '../../core/auth/auth.service';
import { FarrowingSchedule, FarrowingScheduleEntry } from '../../core/services/service.models';
import { ServicesService } from '../../core/services/services.service';

@Component({
  selector: 'app-user-home',
  styleUrl: './user-home.css',
  templateUrl: './user-home.html',
})
export class UserHome implements OnInit {
  private readonly auth = inject(AuthService);
  private readonly servicesService = inject(ServicesService);

  protected readonly user = this.auth.user;
  protected readonly schedule = signal<FarrowingSchedule | null>(null);
  protected readonly loading = signal(false);
  protected readonly error = signal(false);
  protected readonly filterCode = signal('');

  protected readonly hasFilter = computed(() => this.filterCode().trim().length > 0);

  protected readonly upcoming = computed(() => this.filterEntries(this.schedule()?.upcoming ?? []));

  protected readonly overdue = computed(() => this.filterEntries(this.schedule()?.overdue ?? []));

  async ngOnInit(): Promise<void> {
    await this.loadSchedule();
  }

  protected async loadSchedule(): Promise<void> {
    this.loading.set(true);
    this.error.set(false);
    try {
      this.schedule.set(await firstValueFrom(this.servicesService.getFarrowingSchedule()));
    } catch {
      this.error.set(true);
      this.schedule.set(null);
    } finally {
      this.loading.set(false);
    }
  }

  protected onFilterInput(event: Event): void {
    this.filterCode.set((event.target as HTMLInputElement).value);
  }

  protected clearFilter(): void {
    this.filterCode.set('');
  }

  protected averageGestation(entry: FarrowingScheduleEntry): string {
    if (entry.average_gestation_days === null) {
      return 'Sin historial';
    }
    return `${Number(entry.average_gestation_days.toFixed(1))} días`;
  }

  protected averageDate(entry: FarrowingScheduleEntry): string {
    return entry.average_expected_farrowing_date ?? 'Sin historial';
  }

  private filterEntries(entries: FarrowingScheduleEntry[]): FarrowingScheduleEntry[] {
    const term = this.filterCode().trim().toLowerCase();
    if (!term) {
      return entries;
    }
    return entries.filter((entry) => entry.sow_code.toLowerCase().includes(term));
  }
}
