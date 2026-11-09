import {} from "./out/provider";
import { forwarded, fromAnnotation, fromImport, nested, Box, result, direct, arrow, nativeCopy, widened } from "./out/library";
declare module "./out/provider" { interface Open { second: number } }
type Assert<T extends true> = T;
type Check0 = Assert<typeof forwarded extends 2 ? true : false>;
type Check1 = Assert<typeof fromAnnotation extends 2 ? true : false>;
type Check2 = Assert<typeof fromImport extends 2 ? true : false>;
type Check3 = Assert<typeof nested.value extends 2 ? true : false>;
type Check4 = Assert<Box["value"] extends 2 ? true : false>;
type Check5 = Assert<ReturnType<typeof result> extends 2 ? true : false>;
type Native = Assert<typeof nativeCopy extends 1 ? true : false>;
type Widened = Assert<number extends typeof widened ? true : false>;
type Check6 = Assert<ReturnType<typeof direct> extends 2 ? true : false>;
type Check7 = Assert<ReturnType<typeof arrow> extends 2 ? true : false>;
type Check8 = Assert<Box["current"] extends 2 ? true : false>;
