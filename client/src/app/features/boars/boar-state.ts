import { BoarState } from '../../core/boars/boar.models';

export const BOAR_STATE_LABELS: Record<BoarState, string> = {
  Vivo: 'Vivo',
  Muerto: 'Muerto',
  Desecho: 'Desecho',
  Sacrificado: 'Sacrificado',
};

export const BOAR_STATE_CLASSES: Record<BoarState, string> = {
  Vivo: 'bg-success/10 text-success',
  Muerto: 'bg-danger/10 text-danger',
  Desecho: 'bg-warning/10 text-warning',
  Sacrificado: 'bg-primary/10 text-primary',
};
