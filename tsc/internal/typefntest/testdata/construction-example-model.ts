import type {ReadonlyPublic} from "./construction";
export declare const tag:unique symbol;
export declare class Secret { private brand; value:string; }
export interface Open { name?:string; }
export type OpenView = ReadonlyPublic<Open>;
export interface Callable {
  readonly leaf:Secret;
  (value:string):string;
  <T>(value:T):T;
  new(value:string):Secret;
}
export interface Indexed { [key:string]:Secret; [key:number]:Secret; fixed:Secret; }
