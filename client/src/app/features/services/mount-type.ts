import { MountType } from '../../core/services/service.models';

export const MOUNT_TYPE_LABELS: Record<MountType, string> = {
  Natural: 'Natural',
  Artificial: 'Artificial',
};

export const MOUNT_TYPE_CLASSES: Record<MountType, string> = {
  Natural: 'bg-success/10 text-success',
  Artificial: 'bg-info/10 text-info',
};
