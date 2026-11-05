export interface SowHistorySow {
  code: string;
  entry_date: string;
  birth_date: string | null;
  state: string;
  parity: number;
  breed_name: string;
  age: string;
}

export interface SowHistoryServiceMount {
  boar_code: string;
  operator_name: string;
  mount_date: string;
  type: string;
}

export interface SowHistoryService {
  sow_code: string;
  note: string | null;
  state: string;
  location: string | null;
  mounts: SowHistoryServiceMount[];
}

export interface SowHistoryAbortion {
  sow_code: string;
  abortion_date: string;
  cause: string;
  note: string | null;
}

export interface SowHistoryFarrowingOperator {
  operator_name: string;
}

export interface SowHistoryFarrowingMedication {
  medication_name: string;
  dose: number;
  applied_by_name: string;
}

export interface SowHistoryFarrowing {
  farrow_date: string;
  start_time: string | null;
  end_time: string | null;
  duration: string | null;
  location: string | null;
  live_born: number;
  mummified: number;
  stillborn: number;
  litter_weight: number | null;
  stillborn_weight: number | null;
  is_manipulated: boolean;
  note: string | null;
  is_nurse: boolean;
  operators: SowHistoryFarrowingOperator[];
  medications: SowHistoryFarrowingMedication[];
}

export interface SowHistoryPigletDeath {
  operator_name: string;
  death_date: string;
  quantity: number;
  weight: number | null;
  cause: string;
  turn: string;
  note: string | null;
}

export interface SowHistoryPigletFostering {
  sow_id: string;
  role: 'donor' | 'receiver';
  movement_date: string;
  quantity: number;
  note: string | null;
}

export interface SowHistoryPartialWeaging {
  weaging_date: string;
  quantity: number;
  total_weight: number | null;
  type: string;
  note: string | null;
}

export interface SowHistoryWeaging {
  weaging_date: string;
  quantity: number;
  total_weight: number | null;
  note: string | null;
}

export type SowHistoryEvent =
  | { type: 'sow_entry'; date: string; data: SowHistorySow }
  | { type: 'service'; date: string; data: SowHistoryService }
  | { type: 'abortion'; date: string; data: SowHistoryAbortion }
  | { type: 'farrowing'; date: string; data: SowHistoryFarrowing }
  | { type: 'piglet_death'; date: string; data: SowHistoryPigletDeath }
  | { type: 'piglet_fostering'; date: string; data: SowHistoryPigletFostering }
  | { type: 'partial_weaging'; date: string; data: SowHistoryPartialWeaging }
  | { type: 'weaging'; date: string; data: SowHistoryWeaging };

export interface SowHistoryAverages {
  average_live_born: number | null;
  average_stillborn: number | null;
  average_mummified: number | null;
  average_live_born_weight: number | null;
  average_litter_weight: number | null;
  average_stillborn_weight: number | null;
  mortality_rate_percentage: number | null;
  average_deaths_per_farrowing: number | null;
  average_dead_piglet_weight: number | null;
  average_weaned_piglets: number | null;
  average_weaned_piglet_weight: number | null;
  average_gestation_days: number | null;
  average_weaning_age_days: number | null;
  manipulated_farrowings: number;
  total_farrowings: number;
  manipulated_text: string;
  average_farrowing_duration: string | null;
  average_farrowing_duration_formatted: string | null;
}

export interface SowHistory {
  sow: SowHistorySow;
  events: SowHistoryEvent[];
  averages: SowHistoryAverages;
}
