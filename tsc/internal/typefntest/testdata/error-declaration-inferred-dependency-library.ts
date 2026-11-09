import {read} from "./provider";
export const value = read();
export function result() { return read(); }
export class Box { value = read(); }
export default read();
