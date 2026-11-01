import { Abortion } from '../abortions/abortion.models';
import { Farrowing } from '../farrowings/farrowing.models';
import { PartialWeaging } from '../partial-weagings/partial-weaging.models';
import { PigletDeath } from '../piglet-deaths/piglet-death.models';
import { PigletFostering } from '../piglet-fosterings/piglet-fostering.models';
import { Service } from '../services/service.models';
import { Sow } from './sow.models';
import { Weaging } from '../weagings/weaging.models';

export type SowHistoryEvent =
  | { type: 'sow_entry'; date: string; data: Sow }
  | { type: 'service'; date: string; data: Service }
  | { type: 'abortion'; date: string; data: Abortion }
  | { type: 'farrowing'; date: string; data: Farrowing }
  | { type: 'piglet_death'; date: string; data: PigletDeath }
  | { type: 'piglet_fostering'; date: string; data: PigletFostering }
  | { type: 'partial_weaging'; date: string; data: PartialWeaging }
  | { type: 'weaging'; date: string; data: Weaging };

export interface SowHistory {
  sow: Sow;
  events: SowHistoryEvent[];
}
