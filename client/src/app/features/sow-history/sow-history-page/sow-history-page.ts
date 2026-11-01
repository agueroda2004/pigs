import { HttpErrorResponse } from '@angular/common/http';
import { DatePipe, NgClass } from '@angular/common';
import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { SowHistory, SowHistoryEvent } from '../../../core/sows/sow-history.models';
import { SowsService } from '../../../core/sows/sows.service';

@Component({
  selector: 'app-sow-history-page',
  imports: [DatePipe, NgClass, ReactiveFormsModule],
  styleUrl: './sow-history-page.css',
  templateUrl: './sow-history-page.html',
})
export class SowHistoryPage {
  private readonly sowsService = inject(SowsService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly history = signal<SowHistory | null>(null);
  protected readonly loading = signal(false);
  protected readonly error = signal('');
  protected readonly searchForm = this.formBuilder.nonNullable.group({
    code: ['', [Validators.required, Validators.maxLength(50)]],
  });

  protected async search(): Promise<void> {
    if (this.searchForm.invalid) {
      this.searchForm.controls.code.markAsTouched();
      return;
    }

    const code = this.searchForm.controls.code.value.trim();
    if (!code) return;

    this.loading.set(true);
    this.error.set('');
    this.history.set(null);
    try {
      this.history.set(await firstValueFrom(this.sowsService.getSowHistory(code)));
    } catch (cause) {
      const response = cause as HttpErrorResponse;
      const message = response.error?.error;
      if (response.status === 404 && message !== 'Esta cerda ya no se encuentra en la granja') {
        this.error.set('No se encontró una cerda con ese código.');
      } else {
        this.error.set(message || 'No se pudo cargar el historial. Inténtelo de nuevo.');
      }
    } finally {
      this.loading.set(false);
    }
  }

  protected eventTitle(event: SowHistoryEvent): string {
    const titles: Record<SowHistoryEvent['type'], string> = {
      sow_entry: 'Ingreso a la granja',
      service: 'Servicio',
      abortion: 'Aborto',
      farrowing: 'Parto',
    };
    return titles[event.type];
  }

  protected eventBorderClass(type: SowHistoryEvent['type']): string {
    switch (type) {
      case 'sow_entry':
        return 'border-l-4 border-l-gray-400';
      case 'service':
        return 'border-l-4 border-l-green-500';
      case 'abortion':
        return 'border-l-4 border-l-red-500';
      case 'farrowing':
        return 'border-l-4 border-l-blue-500';
      default:
        return 'border-l-4 border-l-gray-300';
    }
  }
}
