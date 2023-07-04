import { EvaluationResult } from "@brease/sdk";

export type ApplyFunction<T extends object, R extends object> = (
  action: EvaluationResult.Model,
  obj: T,
) => Promise<T | (T & R)>;

export interface Action<T extends object, R extends object> {
  kind: string;
  apply: ApplyFunction<T, R>;
}

export const $setAction = async <T extends object, R extends object>(
  action: T,
  obj: R,
) => {
  if (action.action !== "$set") return;

  if (action.targetID) {
    obj[action.targetID] = action.value;
  }

  return obj as T & R;
};

/**
 * A helper function to create a type-safe business rule action
 * @param kind The action kind to look for in the rule evaluation results
 * @param apply The function to execute when an action in the rule evaluation results is found
 * @returns
 */
export const createActionHelper = <T extends object, R extends object>(
  apply: ApplyFunction<T, R>,
) => apply;

export const builtinActions = [$setAction];
