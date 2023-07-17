import {
  ZodSchema,
  ZodTypeDef,
  ZodTypeAny,
  ZodRawShape,
  ZodType,
  ZodTypeWithoutNonValues,
} from "zod";
// Using from https://github.com/temporalio/sdk-typescript/blob/main/packages/common/src/encoding.ts
// Pasted with modifications from: https://raw.githubusercontent.com/anonyco/FastestSmallestTextEncoderDecoder/master/EncoderDecoderTogether.src.js
/* eslint no-fallthrough: 0 */

const fromCharCode = String.fromCharCode;
const encoderRegexp =
  /[\x80-\uD7ff\uDC00-\uFFFF]|[\uD800-\uDBFF][\uDC00-\uDFFF]?/g;
const tmpBufferU16 = new Uint16Array(32);

export class TextDecoder {
  decode(
    inputArrayOrBuffer: Uint8Array | ArrayBuffer | SharedArrayBuffer,
  ): string {
    const inputAs8 =
      inputArrayOrBuffer instanceof Uint8Array
        ? inputArrayOrBuffer
        : new Uint8Array(inputArrayOrBuffer);

    let resultingString = "",
      tmpStr = "",
      index = 0,
      nextEnd = 0,
      cp0 = 0,
      codePoint = 0,
      minBits = 0,
      cp1 = 0,
      pos = 0,
      tmp = -1;
    const len = inputAs8.length | 0;
    const lenMinus32 = (len - 32) | 0;
    // Note that tmp represents the 2nd half of a surrogate pair incase a surrogate gets divided between blocks
    for (; index < len; ) {
      nextEnd = index <= lenMinus32 ? 32 : (len - index) | 0;
      for (; pos < nextEnd; index = (index + 1) | 0, pos = (pos + 1) | 0) {
        cp0 = inputAs8[index] & 0xff;
        switch (cp0 >> 4) {
          case 15:
            cp1 = inputAs8[(index = (index + 1) | 0)] & 0xff;
            if (cp1 >> 6 !== 0b10 || 0b11110111 < cp0) {
              index = (index - 1) | 0;
              break;
            }
            codePoint = ((cp0 & 0b111) << 6) | (cp1 & 0b00111111);
            minBits = 5; // 20 ensures it never passes -> all invalid replacements
            cp0 = 0x100; //  keep track of th bit size
          case 14:
            cp1 = inputAs8[(index = (index + 1) | 0)] & 0xff;
            codePoint <<= 6;
            codePoint |= ((cp0 & 0b1111) << 6) | (cp1 & 0b00111111);
            minBits = cp1 >> 6 === 0b10 ? (minBits + 4) | 0 : 24; // 24 ensures it never passes -> all invalid replacements
            cp0 = (cp0 + 0x100) & 0x300; // keep track of th bit size
          case 13:
          case 12:
            cp1 = inputAs8[(index = (index + 1) | 0)] & 0xff;
            codePoint <<= 6;
            codePoint |= ((cp0 & 0b11111) << 6) | (cp1 & 0b00111111);
            minBits = (minBits + 7) | 0;

            // Now, process the code point
            if (
              index < len &&
              cp1 >> 6 === 0b10 &&
              codePoint >> minBits &&
              codePoint < 0x110000
            ) {
              cp0 = codePoint;
              codePoint = (codePoint - 0x10000) | 0;
              if (0 <= codePoint /*0xffff < codePoint*/) {
                // BMP code point
                //nextEnd = nextEnd - 1|0;

                tmp = ((codePoint >> 10) + 0xd800) | 0; // highSurrogate
                cp0 = ((codePoint & 0x3ff) + 0xdc00) | 0; // lowSurrogate (will be inserted later in the switch-statement)

                if (pos < 31) {
                  // notice 31 instead of 32
                  tmpBufferU16[pos] = tmp;
                  pos = (pos + 1) | 0;
                  tmp = -1;
                } else {
                  // else, we are at the end of the inputAs8 and let tmp0 be filled in later on
                  // NOTE that cp1 is being used as a temporary variable for the swapping of tmp with cp0
                  cp1 = tmp;
                  tmp = cp0;
                  cp0 = cp1;
                }
              } else nextEnd = (nextEnd + 1) | 0; // because we are advancing i without advancing pos
            } else {
              // invalid code point means replacing the whole thing with null replacement characters
              cp0 >>= 8;
              index = (index - cp0 - 1) | 0; // reset index  back to what it was before
              cp0 = 0xfffd;
            }

            // Finally, reset the variables for the next go-around
            minBits = 0;
            codePoint = 0;
            nextEnd = index <= lenMinus32 ? 32 : (len - index) | 0;
          /*case 11:
        case 10:
        case 9:
        case 8:
          codePoint ? codePoint = 0 : cp0 = 0xfffd; // fill with invalid replacement character
        case 7:
        case 6:
        case 5:
        case 4:
        case 3:
        case 2:
        case 1:
        case 0:
          tmpBufferU16[pos] = cp0;
          continue;*/
          default: // fill with invalid replacement character
            tmpBufferU16[pos] = cp0;
            continue;
          case 11:
          case 10:
          case 9:
          case 8:
        }
        tmpBufferU16[pos] = 0xfffd; // fill with invalid replacement character
      }
      tmpStr += fromCharCode(
        tmpBufferU16[0],
        tmpBufferU16[1],
        tmpBufferU16[2],
        tmpBufferU16[3],
        tmpBufferU16[4],
        tmpBufferU16[5],
        tmpBufferU16[6],
        tmpBufferU16[7],
        tmpBufferU16[8],
        tmpBufferU16[9],
        tmpBufferU16[10],
        tmpBufferU16[11],
        tmpBufferU16[12],
        tmpBufferU16[13],
        tmpBufferU16[14],
        tmpBufferU16[15],
        tmpBufferU16[16],
        tmpBufferU16[17],
        tmpBufferU16[18],
        tmpBufferU16[19],
        tmpBufferU16[20],
        tmpBufferU16[21],
        tmpBufferU16[22],
        tmpBufferU16[23],
        tmpBufferU16[24],
        tmpBufferU16[25],
        tmpBufferU16[26],
        tmpBufferU16[27],
        tmpBufferU16[28],
        tmpBufferU16[29],
        tmpBufferU16[30],
        tmpBufferU16[31],
      );
      if (pos < 32) tmpStr = tmpStr.slice(0, (pos - 32) | 0); //-(32-pos));
      if (index < len) {
        //fromCharCode.apply(0, tmpBufferU16 : Uint8Array ?  tmpBufferU16.subarray(0,pos) : tmpBufferU16.slice(0,pos));
        tmpBufferU16[0] = tmp;
        pos = ~tmp >>> 31; //tmp !== -1 ? 1 : 0;
        tmp = -1;

        if (tmpStr.length < resultingString.length) continue;
      } else if (tmp !== -1) {
        tmpStr += fromCharCode(tmp);
      }

      resultingString += tmpStr;
      tmpStr = "";
    }

    return resultingString;
  }
}

//////////////////////////////////////////////////////////////////////////////////////
function encoderReplacer(nonAsciiChars: string) {
  // make the UTF string into a binary UTF-8 encoded string
  let point = nonAsciiChars.charCodeAt(0) | 0;
  if (0xd800 <= point) {
    if (point <= 0xdbff) {
      const nextcode = nonAsciiChars.charCodeAt(1) | 0; // defaults to 0 when NaN, causing null replacement character

      if (0xdc00 <= nextcode && nextcode <= 0xdfff) {
        //point = ((point - 0xD800)<<10) + nextcode - 0xDC00 + 0x10000|0;
        point = ((point << 10) + nextcode - 0x35fdc00) | 0;
        if (point > 0xffff)
          return fromCharCode(
            (0x1e /*0b11110*/ << 3) | (point >> 18),
            (0x2 /*0b10*/ << 6) | ((point >> 12) & 0x3f) /*0b00111111*/,
            (0x2 /*0b10*/ << 6) | ((point >> 6) & 0x3f) /*0b00111111*/,
            (0x2 /*0b10*/ << 6) | (point & 0x3f) /*0b00111111*/,
          );
      } else point = 65533 /*0b1111111111111101*/; //return '\xEF\xBF\xBD';//fromCharCode(0xef, 0xbf, 0xbd);
    } else if (point <= 0xdfff) {
      point = 65533 /*0b1111111111111101*/; //return '\xEF\xBF\xBD';//fromCharCode(0xef, 0xbf, 0xbd);
    }
  }
  /*if (point <= 0x007f) return nonAsciiChars;
  else */ if (point <= 0x07ff) {
    return fromCharCode((0x6 << 5) | (point >> 6), (0x2 << 6) | (point & 0x3f));
  } else
    return fromCharCode(
      (0xe /*0b1110*/ << 4) | (point >> 12),
      (0x2 /*0b10*/ << 6) | ((point >> 6) & 0x3f) /*0b00111111*/,
      (0x2 /*0b10*/ << 6) | (point & 0x3f) /*0b00111111*/,
    );
}

export class TextEncoder {
  public encode(inputString: string): Uint8Array {
    // 0xc0 => 0b11000000; 0xff => 0b11111111; 0xc0-0xff => 0b11xxxxxx
    // 0x80 => 0b10000000; 0xbf => 0b10111111; 0x80-0xbf => 0b10xxxxxx
    const encodedString = inputString === void 0 ? "" : "" + inputString,
      len = encodedString.length | 0;
    let result = new Uint8Array(((len << 1) + 8) | 0);
    let tmpResult: Uint8Array;
    let i = 0,
      pos = 0,
      point = 0,
      nextcode = 0;
    let upgradededArraySize = !Uint8Array; // normal arrays are auto-expanding
    for (i = 0; i < len; i = (i + 1) | 0, pos = (pos + 1) | 0) {
      point = encodedString.charCodeAt(i) | 0;
      if (point <= 0x007f) {
        result[pos] = point;
      } else if (point <= 0x07ff) {
        result[pos] = (0x6 << 5) | (point >> 6);
        result[(pos = (pos + 1) | 0)] = (0x2 << 6) | (point & 0x3f);
      } else {
        widenCheck: {
          if (0xd800 <= point) {
            if (point <= 0xdbff) {
              nextcode = encodedString.charCodeAt((i = (i + 1) | 0)) | 0; // defaults to 0 when NaN, causing null replacement character

              if (0xdc00 <= nextcode && nextcode <= 0xdfff) {
                //point = ((point - 0xD800)<<10) + nextcode - 0xDC00 + 0x10000|0;
                point = ((point << 10) + nextcode - 0x35fdc00) | 0;
                if (point > 0xffff) {
                  result[pos] = (0x1e /*0b11110*/ << 3) | (point >> 18);
                  result[(pos = (pos + 1) | 0)] =
                    (0x2 /*0b10*/ << 6) | ((point >> 12) & 0x3f) /*0b00111111*/;
                  result[(pos = (pos + 1) | 0)] =
                    (0x2 /*0b10*/ << 6) | ((point >> 6) & 0x3f) /*0b00111111*/;
                  result[(pos = (pos + 1) | 0)] =
                    (0x2 /*0b10*/ << 6) | (point & 0x3f) /*0b00111111*/;
                  continue;
                }
                break widenCheck;
              }
              point = 65533 /*0b1111111111111101*/; //return '\xEF\xBF\xBD';//fromCharCode(0xef, 0xbf, 0xbd);
            } else if (point <= 0xdfff) {
              point = 65533 /*0b1111111111111101*/; //return '\xEF\xBF\xBD';//fromCharCode(0xef, 0xbf, 0xbd);
            }
          }
          if (
            !upgradededArraySize &&
            i << 1 < pos &&
            i << 1 < ((pos - 7) | 0)
          ) {
            upgradededArraySize = true;
            tmpResult = new Uint8Array(len * 3);
            tmpResult.set(result);
            result = tmpResult;
          }
        }
        result[pos] = (0xe /*0b1110*/ << 4) | (point >> 12);
        result[(pos = (pos + 1) | 0)] =
          (0x2 /*0b10*/ << 6) | ((point >> 6) & 0x3f) /*0b00111111*/;
        result[(pos = (pos + 1) | 0)] =
          (0x2 /*0b10*/ << 6) | (point & 0x3f) /*0b00111111*/;
      }
    }
    return Uint8Array ? result.subarray(0, pos) : result.slice(0, pos);
  }

  public encodeInto(
    inputString: string,
    u8Arr: Uint8Array,
  ): { written: number; read: number } {
    const encodedString =
      inputString === void 0
        ? ""
        : ("" + inputString).replace(encoderRegexp, encoderReplacer);
    let len = encodedString.length | 0,
      i = 0,
      char = 0,
      read = 0;
    const u8ArrLen = u8Arr.length | 0;
    const inputLength = inputString.length | 0;
    if (u8ArrLen < len) len = u8ArrLen;
    putChars: {
      for (; i < len; i = (i + 1) | 0) {
        char = encodedString.charCodeAt(i) | 0;
        switch (char >> 4) {
          case 0:
          case 1:
          case 2:
          case 3:
          case 4:
          case 5:
          case 6:
          case 7:
            read = (read + 1) | 0;
          // extension points:
          case 8:
          case 9:
          case 10:
          case 11:
            break;
          case 12:
          case 13:
            if (((i + 1) | 0) < u8ArrLen) {
              read = (read + 1) | 0;
              break;
            }
          case 14:
            if (((i + 2) | 0) < u8ArrLen) {
              //if (!(char === 0xEF && encodedString.substr(i+1|0,2) === "\xBF\xBD"))
              read = (read + 1) | 0;
              break;
            }
          case 15:
            if (((i + 3) | 0) < u8ArrLen) {
              read = (read + 1) | 0;
              break;
            }
          default:
            break putChars;
        }
        //read = read + ((char >> 6) !== 2) |0;
        u8Arr[i] = char;
      }
    }
    return { written: i, read: inputLength < read ? inputLength : read };
  }
}

/**
 * Encode a UTF-8 string into a Uint8Array
 */
export function encodeToUint8Array(s: string): Uint8Array {
  return TextEncoder.prototype.encode(s);
}

/**
 * Decode a Uint8Array into a UTF-8 string
 */
export function decodeFromUint8Array(a: Uint8Array): string {
  return TextDecoder.prototype.decode(a);
}

/*
https://gist.github.com/enepomnyaschih/54c437997f8202871278d0fdf68148ca

MIT License
Copyright (c) 2020 Egor Nepomnyaschih
Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

/*
// This constant can also be computed with the following algorithm:
const base64abc = [],
	A = "A".charCodeAt(0),
	a = "a".charCodeAt(0),
	n = "0".charCodeAt(0);
for (let i = 0; i < 26; ++i) {
	base64abc.push(String.fromCharCode(A + i));
}
for (let i = 0; i < 26; ++i) {
	base64abc.push(String.fromCharCode(a + i));
}
for (let i = 0; i < 10; ++i) {
	base64abc.push(String.fromCharCode(n + i));
}
base64abc.push("+");
base64abc.push("/");
*/
const base64abc = [
  "A",
  "B",
  "C",
  "D",
  "E",
  "F",
  "G",
  "H",
  "I",
  "J",
  "K",
  "L",
  "M",
  "N",
  "O",
  "P",
  "Q",
  "R",
  "S",
  "T",
  "U",
  "V",
  "W",
  "X",
  "Y",
  "Z",
  "a",
  "b",
  "c",
  "d",
  "e",
  "f",
  "g",
  "h",
  "i",
  "j",
  "k",
  "l",
  "m",
  "n",
  "o",
  "p",
  "q",
  "r",
  "s",
  "t",
  "u",
  "v",
  "w",
  "x",
  "y",
  "z",
  "0",
  "1",
  "2",
  "3",
  "4",
  "5",
  "6",
  "7",
  "8",
  "9",
  "+",
  "/",
];

/*
// This constant can also be computed with the following algorithm:
const l = 256, base64codes = new Uint8Array(l);
for (let i = 0; i < l; ++i) {
	base64codes[i] = 255; // invalid character
}
base64abc.forEach((char, index) => {
	base64codes[char.charCodeAt(0)] = index;
});
base64codes["=".charCodeAt(0)] = 0; // ignored anyway, so we just need to prevent an error
*/
const base64codes = [
  255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
  255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
  255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 62, 255, 255,
  255, 63, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 255, 255, 255, 0, 255, 255,
  255, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
  21, 22, 23, 24, 25, 255, 255, 255, 255, 255, 255, 26, 27, 28, 29, 30, 31, 32,
  33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51,
];

function getBase64Code(charCode: number) {
  if (charCode >= base64codes.length) {
    throw new Error("Unable to parse base64 string.");
  }
  const code = base64codes[charCode];
  if (code === 255) {
    throw new Error("Unable to parse base64 string.");
  }
  return code;
}

export function bytesToBase64(bytes: number[] | Uint8Array) {
  let result = "",
    i,
    l = bytes.length;
  for (i = 2; i < l; i += 3) {
    result += base64abc[bytes[i - 2] >> 2];
    result += base64abc[((bytes[i - 2] & 0x03) << 4) | (bytes[i - 1] >> 4)];
    result += base64abc[((bytes[i - 1] & 0x0f) << 2) | (bytes[i] >> 6)];
    result += base64abc[bytes[i] & 0x3f];
  }
  if (i === l + 1) {
    // 1 octet yet to write
    result += base64abc[bytes[i - 2] >> 2];
    result += base64abc[(bytes[i - 2] & 0x03) << 4];
    result += "==";
  }
  if (i === l) {
    // 2 octets yet to write
    result += base64abc[bytes[i - 2] >> 2];
    result += base64abc[((bytes[i - 2] & 0x03) << 4) | (bytes[i - 1] >> 4)];
    result += base64abc[(bytes[i - 1] & 0x0f) << 2];
    result += "=";
  }
  return result;
}

export function base64ToBytes(str: string) {
  if (str.length % 4 !== 0) {
    throw new Error("Unable to parse base64 string.");
  }
  const index = str.indexOf("=");
  if (index !== -1 && index < str.length - 2) {
    throw new Error("Unable to parse base64 string.");
  }
  let missingOctets = str.endsWith("==") ? 2 : str.endsWith("=") ? 1 : 0,
    n = str.length,
    result = new Uint8Array(3 * (n / 4)),
    buffer: number;
  for (let i = 0, j = 0; i < n; i += 4, j += 3) {
    buffer =
      (getBase64Code(str.charCodeAt(i)) << 18) |
      (getBase64Code(str.charCodeAt(i + 1)) << 12) |
      (getBase64Code(str.charCodeAt(i + 2)) << 6) |
      getBase64Code(str.charCodeAt(i + 3));
    result[j] = buffer >> 16;
    result[j + 1] = (buffer >> 8) & 0xff;
    result[j + 2] = buffer & 0xff;
  }
  return result.subarray(0, result.length - missingOctets);
}

export function base64encode(
  str: string,
  encoder: {
    encode: (str: string) => Uint8Array | number[];
  } = new TextEncoder(),
) {
  return bytesToBase64(encoder.encode(str));
}

export function base64decode(
  str: string,
  decoder: { decode: (bytes: Uint8Array) => string } = new TextDecoder(),
) {
  return decoder.decode(base64ToBytes(str));
}

/**
 * Encode a any object into a base64 string
 */
export function encode(a: any): string {
  return base64encode(JSON.stringify(a));
}

/**
 * @internal
 */
type Schema<TData> = {
  parse: (data: unknown) => TData;
};

/**
 * Decode a base64 string into a any
 */
export function decode(s: string): any;
export function decode<T>(s: string, schema: Schema<T>): T;
export function decode<T>(s: string, schema?: Schema<T>): any {
  const objStr = base64decode(s);
  let obj: unknown;
  try {
    obj = JSON.parse(objStr);
  } catch (error) {
    obj = objStr;
  }

  if (schema) {
    return schema.parse(obj);
  }

  return obj;
}
