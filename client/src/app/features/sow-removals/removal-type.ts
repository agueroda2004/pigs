import { RemovalType } from '../../core/sow-removals/sow-removal.models';

export const REMOVAL_TYPE_LABELS: Record<RemovalType, string> = {
  Muerte: 'Muerte',
  Desecho: 'Desecho',
  Sacrificio: 'Sacrificio',
};

export const REMOVAL_TYPE_CLASSES: Record<RemovalType, string> = {
  Muerte: 'bg-danger/10 text-danger',
  Desecho: 'bg-warning/10 text-warning',
  Sacrificio: 'bg-primary/10 text-primary',
};
