// PARAMETERS testing
// run with npx tsx packages/core/src/parameters.ts

import { encode, decode } from "./encoder.js";
import { isEqual } from "lodash-es";

const testEncode = (input: any) => {
  const output = encode(input);
  console.log({
    action: "encode",
    input,
    output,
    type: typeof output,
  });
  return output;
};

const testDecode = (input: string, test?: any) => {
  const output = decode(input);
  const equals = isEqual(output, test);
  console.log({
    action: "decode",
    input,
    output,
    type: typeof output,
    test,
    equals,
  });
  if (!equals) {
    throw new Error(
      `shit, not the same: '${input}' and '${JSON.stringify(test)}'`,
    );
  }
};

const te9 = testEncode(9);
const tegov = testEncode(".gov");
const teLiveFish = testEncode("Livefish");
const teUA = testEncode(["Ukraine"]);
const teR = testEncode(["red"]);
const teB = testEncode(["blue"]);
const teRG = testEncode(["red", "green"]);
const teRegex = testEncode("\\d{2}/\\d{2}/20\\d{2}");
const teAb = testEncode({ a: "b" });
const teExample = testEncode("example");

testDecode(te9, 9);
testDecode(tegov, ".gov");
testDecode(teLiveFish, "Livefish");
testDecode(teUA, ["Ukraine"]);
testDecode(teR, ["red"]);
testDecode(teB, ["blue"]);
testDecode(teRG, ["red", "green"]);
testDecode(teRegex, "\\d{2}/\\d{2}/20\\d{2}");
testDecode(teAb, { a: "b" });
testDecode(teExample, "example");
