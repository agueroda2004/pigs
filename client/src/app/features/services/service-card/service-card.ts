import { Component, computed, input } from '@angular/core';

import { MountType, Service } from '../../../core/services/service.models';
import { MOUNT_TYPE_CLASSES, MOUNT_TYPE_LABELS } from '../mount-type';
import { SERVICE_STATE_CLASSES, SERVICE_STATE_LABELS } from '../service-state';

@Component({
  selector: 'app-service-card',
  styleUrl: './service-card.css',
  templateUrl: './service-card.html',
})
export class ServiceCard {
  readonly service = input.required<Service>();
  readonly sowCode = input<string | null>(null);
  readonly boarCodes = input<Map<string, string>>(new Map());
  readonly operatorNames = input<Map<string, string>>(new Map());

  protected readonly stateLabel = computed(() => SERVICE_STATE_LABELS[this.service().state]);
  protected readonly stateClass = computed(() => SERVICE_STATE_CLASSES[this.service().state]);

  protected mountTypeLabel(type: MountType): string {
    return MOUNT_TYPE_LABELS[type];
  }

  protected mountTypeClass(type: MountType): string {
    return MOUNT_TYPE_CLASSES[type];
  }

  protected boarCode(id: string): string {
    return this.boarCodes().get(id) ?? 'Verraco';
  }

  protected operatorName(id: string): string {
    return this.operatorNames().get(id) ?? 'Operador';
  }
}
