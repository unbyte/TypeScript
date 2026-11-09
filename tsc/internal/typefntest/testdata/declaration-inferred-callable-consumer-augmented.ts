import { result, fromArrow, fromGenerated, Generic, widened, type Open } from "./out/library";
declare module "./out/library" { interface Open { second: number } }
type Assert<T extends true> = T;
type Check0 = Assert<typeof result extends 2 ? true : false>;
type Check1 = Assert<typeof fromArrow extends 2 ? true : false>;
type Check2 = Assert<typeof fromGenerated extends 2 ? true : false>;
type Check3 = Assert<Generic<Open>["value"] extends 2 ? true : false>;
type Widened = Assert<number extends typeof widened ? true : false>;
