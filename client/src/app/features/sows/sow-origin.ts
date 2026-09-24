import { SowOrigin } from '../../core/sows/sow.models';

export const SOW_ORIGIN_LABELS: Record<SowOrigin, string> = {
  Propio: 'Propio',
  Externo: 'Externo',
};

export const SOW_ORIGIN_CLASSES: Record<SowOrigin, string> = {
  Propio: 'bg-muted text-muted-foreground',
  Externo: 'bg-info/10 text-info',
};
