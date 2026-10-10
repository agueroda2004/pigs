import { Component, computed, input, output, signal } from '@angular/core';

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
  readonly canEdit = input(false);
  readonly editRequested = output<Service>();
  readonly deleteRequested = output<Service>();

  protected readonly stateLabel = computed(() => SERVICE_STATE_LABELS[this.service().state]);
  protected readonly stateClass = computed(() => SERVICE_STATE_CLASSES[this.service().state]);
  protected readonly manageable = computed(
    () => this.canEdit() && this.service().state === 'Confirmado',
  );

  protected readonly mountsExpanded = signal(false);
  protected readonly visibleMounts = computed(() =>
    this.mountsExpanded() ? this.service().mounts : this.service().mounts.slice(0, 1),
  );
  protected readonly remainingMounts = computed(() =>
    Math.max(0, this.service().mounts.length - 1),
  );

  protected toggleMounts(): void {
    this.mountsExpanded.update((expanded) => !expanded);
  }

  protected mountTypeLabel(type: MountType): string {
    return MOUNT_TYPE_LABELS[type];
  }

  protected mountTypeClass(type: MountType): string {
    return MOUNT_TYPE_CLASSES[type];
  }
}
