import { SowState } from '../../core/sows/sow.models';

export const SOW_STATE_LABELS: Record<SowState, string> = {
  Viva: 'Viva',
  Muerta: 'Muerta',
  Desecho: 'Desecho',
  Sacrificada: 'Sacrificada',
  Abortada: 'Abortada',
  Gestando: 'Gestando',
  Lactando: 'Lactando',
  Destetada: 'Destetada',
};

export const SOW_STATE_CLASSES: Record<SowState, string> = {
  Viva: 'bg-success/10 text-success',
  Muerta: 'bg-danger/10 text-danger',
  Desecho: 'bg-warning/10 text-warning',
  Sacrificada: 'bg-primary/10 text-primary',
  Abortada: 'bg-danger/10 text-danger',
  Gestando: 'bg-info/10 text-info',
  Lactando: 'bg-success/10 text-success',
  Destetada: 'bg-muted text-muted-foreground',
};
