import { PigletDeathCause } from '../../core/piglet-deaths/piglet-death.models';

export const PIGLET_DEATH_CAUSE_LABELS: Record<PigletDeathCause, string> = {
  Aplastado: 'Aplastado',
  Debilidad: 'Debilidad',
  Diarrea: 'Diarrea',
  Deformidad: 'Deformidad',
  Otro: 'Otro',
  Canibalismo: 'Canibalismo',
  Pata_abierta: 'Pata abierta',
  Bacteria: 'Bacteria',
  Reaccion_medicamento: 'Reacción a medicamento',
};

export const PIGLET_DEATH_CAUSE_CLASSES: Record<PigletDeathCause, string> = {
  Aplastado: 'bg-danger/10 text-danger',
  Debilidad: 'bg-warning/10 text-warning',
  Diarrea: 'bg-info/10 text-info',
  Deformidad: 'bg-warning/10 text-warning',
  Otro: 'bg-muted text-muted-foreground',
  Canibalismo: 'bg-danger/10 text-danger',
  Pata_abierta: 'bg-info/10 text-info',
  Bacteria: 'bg-danger/10 text-danger',
  Reaccion_medicamento: 'bg-warning/10 text-warning',
};
