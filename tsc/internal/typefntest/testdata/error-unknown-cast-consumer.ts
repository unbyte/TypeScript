import {fromUnknown,guarded,eraseGuard} from "./library";
const first = fromUnknown<number>(null);
const second = guarded<number>(null);
const third = eraseGuard<number>("ignored result");
