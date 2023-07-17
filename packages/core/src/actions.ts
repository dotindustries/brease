import { EvaluationResult } from "@brease/sdk";
import { isJsonPath, setValue } from "./jsonpath.js";

export type ApplyFunction<T extends object, R extends object> = (
  action: EvaluationResult.Model,
  obj: T,
) => Promise<R>;

export interface Action<T extends object, R extends object> {
  kind: string;
  apply: ApplyFunction<T, R>;
}

export const $setAction = async <T extends object, R extends object>(
  action: EvaluationResult.Model,
  obj: T,
) => {
  if (action.action !== "$set") return {} as R;
  const a: { [k: string]: any } = {};

  if (action.targetID && isJsonPath(action.targetID)) {
    setValue(a, action.targetID, action.value);
  } else if (action.targetID) {
    a[action.targetID] = action.value;
  }

  return a as R;
};

/**
 * A helper function to create a type-safe business rule $set action
 * @returns
 */
export const createTypedSetAction = <R extends object>() => {
  return $setAction<any, R>;
};

export const builtinActions = [$setAction];
