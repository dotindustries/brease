import {JSONPath} from "jsonpath-plus";

/**
 * Checks if a given string is a valid JSONPath expression.
 * @param expression - The string expression to be checked.
 * @returns True if the expression is a valid JSONPath, otherwise false.
 */
export const isJsonPath = (expression: string) => {
  try {
    JSONPath({ path: expression, json: {} });
    return true;
  } catch (e) {
    return false;
  }
};

/**
 * Sets a value at the given JSONPath of an object.
 * @param obj - The object in which the value will be set.
 * @param path - The JSONPath expression where the value will be set.
 * @param value - The value to be set.
 */
export const setValue = (obj: any, path: string, value: any) => {
  JSONPath({
    path,
    json: obj,
    resultType: 'value',
    wrap: false,
    callback: (_payload: any, _type: any, fullPayload) => {
      const target = fullPayload.pointer;
      const segments = target.split('/');
      segments.shift(); // Remove the first empty string segment
      let current = obj;

      for (let i = 0; i < segments.length - 1; i++) {
        const segment = segments[i];
        if (!current[segment]) {
          current[segment] = {};
        }
        current = current[segment];
      }
      current[segments[segments.length - 1]] = value;
    }
  });
};
