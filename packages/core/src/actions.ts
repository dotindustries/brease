import { EvaluationResult } from "@brease/sdk";

export type ApplyFunction<T> = (
  action: EvaluationResult.Model,
  obj: T,
) => Promise<T>;

export interface Action<T extends object> {
  kind: string;
  apply: ApplyFunction<T>;
}

export const $setAction: Action<any> = {
  kind: "$set",
  apply: async (action, obj) => {
    if (action.action !== "$set") return;

    if (action.targetID) {
      obj[action.targetID] = action.value;
    }

    return obj;
  },
};

/**
 * A helper function to create a type-safe business rule action
 * @param kind The action kind to look for in the rule evaluation results
 * @param apply The function to execute when an action in the rule evaluation results is found
 * @returns
 */
export const createActionHelper = <T extends object, R>(
  kind: string,
  apply: ApplyFunction<T>,
) => ({
  kind,
  apply,
});

export const builtinActions = [$setAction];
