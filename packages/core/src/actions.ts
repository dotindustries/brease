import { EvaluationResult } from "@buf/dot_brease.bufbuild_es/brease/rule/v1/model_pb.js";
import { isJsonPath, setValue } from "./jsonpath.js";
import {Result} from "./store.js";

export type ApplyFunction<T extends object, R extends object> = (
  action: Result,
  obj: T,
) => Promise<R>;

export interface ClientActionFn<T extends object, R extends object> {
  kind: string;
  apply: ApplyFunction<T, R>;
}

export const $setAction = async <T extends object, R extends object>(
  action: EvaluationResult,
  _obj: T,
) => {
  if (action.action !== "$set") return {} as R;
  const a: { [k: string]: any } = {};

  if (action.target?.id && isJsonPath(action.target?.id)) {
    setValue(a, action.target?.id, action.target.value);
  } else if (action.target?.id) {
    a[action.target?.id] = action.target.value;
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
