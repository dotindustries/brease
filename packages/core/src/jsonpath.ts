import jp from "jsonpath";

/**
 *
 * @param s the string to be validated
 * @returns boolean
 */
export const isJsonPath = (s: string) => {
  const parsed = jp.parse(s);
  return parsed.filter((e) => e.expression.type === "root").length === 1;
};

export const setValue = (o: any, pathExpression: string, value: any) => {
  jp.value(o, pathExpression, value);
};
