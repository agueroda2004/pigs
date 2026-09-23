import { BoarOrigin } from '../../core/boars/boar.models';

export const BOAR_ORIGIN_LABELS: Record<BoarOrigin, string> = {
  Propio: 'Propio',
  Externo: 'Externo',
};

export const BOAR_ORIGIN_CLASSES: Record<BoarOrigin, string> = {
  Propio: 'bg-muted text-muted-foreground',
  Externo: 'bg-info/10 text-info',
};
