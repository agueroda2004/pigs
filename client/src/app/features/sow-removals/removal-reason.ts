import { RemovalReason } from '../../core/sow-removals/sow-removal.models';

export const REMOVAL_REASON_LABELS: Record<RemovalReason, string> = {
  Edad_Paridad: 'Edad / Paridad',
  Fallo_Reproductivo: 'Fallo reproductivo',
  Baja_Productividad: 'Baja productividad',
  Problema_Locomotor: 'Problema locomotor',
  Enfermedad: 'Enfermedad',
  Muerte_Subita: 'Muerte súbita',
  Otro: 'Otro',
};
