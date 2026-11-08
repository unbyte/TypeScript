import type {Waiting} from "./library.js";
type Pending<T>=Waiting<T>;
type InvalidInput=Pending<"invalid input">;
type InvalidApplication=Pending<1>;
