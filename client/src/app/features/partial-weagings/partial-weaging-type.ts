import { PartialWeagingType } from '../../core/partial-weagings/partial-weaging.models';

export const PARTIAL_WEAGING_TYPE_LABELS: Record<PartialWeagingType, string> = {
  Normal: 'Normal',
  Nodriza: 'Nodriza',
  Baja_Viabilidad: 'Baja viabilidad',
};

export const PARTIAL_WEAGING_TYPE_CLASSES: Record<PartialWeagingType, string> = {
  Normal: 'bg-muted text-muted-foreground',
  Nodriza: 'bg-info/10 text-info',
  Baja_Viabilidad: 'bg-warning/10 text-warning',
};
