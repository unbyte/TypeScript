import { read, type Open } from "./provider";
const specialized = read<Open>;
export const count = specialized();
