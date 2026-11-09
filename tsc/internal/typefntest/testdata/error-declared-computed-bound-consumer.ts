import { forward, checkedForward } from "./out/library";
forward({key:1}, "missing");
checkedForward({bad:1}, "bad");
checkedForward({bad:2}, "bad");
