import { StoreApi, createStore } from "zustand/vanilla";
import hash from "object-hash";
import { immer } from "zustand/middleware/immer";
import { clone } from "lodash-es";
import { EvaluationResult } from "@brease/sdk";
import { ApiEvaluateRulesResponse, EvaluateRulesInput } from "@brease/core";
import { Action, builtinActions } from "./actions.js";

export interface RulesStore<T> {
  isExecuting: boolean;
  lastHash?: string;
  rawActions: EvaluationResult.Model[];
  unknownActions: EvaluationResult.Model[] | undefined;
  executeRules: (object: T) => void;
  result: T | undefined;
}

const hasActionFilter = (
  a: EvaluationResult.Model,
): a is EvaluationResult.Model & {
  action: string;
} => Boolean(a.action);

export type EvaluateInput<T> = Pick<
  EvaluateRulesInput.Model,
  "overrideCode" | "overrideRules"
> & {
  object: T;
};

export type RuleStoreOptions<T extends object> = {
  evaluateRulesFn: (
    input: EvaluateRulesInput.Model,
  ) => Promise<ApiEvaluateRulesResponse.Results | undefined>;
  userDefinedActions?: Action<T>[];
  overrideCode?: string;
  overrideRules?: EvaluateRulesInput.OverrideRules;
};

const createRulesStore = <T extends object>({
  evaluateRulesFn,
  overrideCode,
  overrideRules,
  userDefinedActions,
}: RuleStoreOptions<T>) => {
  return createStore(
    immer<RulesStore<T>>((set, get) => ({
      isExecuting: false,
      rawActions: [],
      unknownActions: undefined,
      result: undefined,
      executeRules: async (object) => {
        if (get().isExecuting) {
          console.warn("Rules are executing, skipping to avoid duplicate run");
          return;
        }

        if (get().rawActions?.length !== 0) {
          console.warn("cleaning rules");
          set({ rawActions: [] });
        }

        const newHash = hash(object);
        if (get().lastHash === newHash) {
          console.warn("deduplicating, results already present in cache");
          return; // no change of data
        }

        set({ isExecuting: true, lastHash: newHash });

        const rawActions = await evaluateRulesFn({
          object: object as any,
          overrideCode,
          overrideRules,
        });

        // extract and apply builtIn actions
        const { knownActions, unknownActions } = findActions(
          rawActions,
          userDefinedActions,
        );

        // apply recognized rules
        const result = await applyActions(object, knownActions);

        set({
          rawActions,
          isExecuting: false,
          unknownActions,
          result,
        });
      },
    })),
  );
};

export const findActions = <T extends object>(
  rawActions: EvaluationResult.Model[] | undefined,
  userDefinedActions?: Action<T>[],
) => {
  if (!rawActions) return { knownActions: [], unknownActions: undefined };

  const builtInActionKinds = builtinActions.map((a) => a.kind);
  const customActionKinds = userDefinedActions?.map((a) => a.kind) ?? [];

  // allow the user defined actions to override the builtin functions
  // such as $set to provide increase type-safe on the results
  const notOverridenBuiltInActionKinds = builtInActionKinds.filter(
    (k) => !customActionKinds.includes(k),
  );

  const knownBuiltInActions = rawActions
    .filter(hasActionFilter)
    .filter((a) => notOverridenBuiltInActionKinds.includes(a.action))
    .map((a) => ({
      ...a,
      apply: builtinActions.find((ba) => ba.kind === a.action)!.apply,
      //                                                      ^ we already pre-filtered...
    }));
  const knownCustomActions =
    (userDefinedActions &&
      rawActions
        .filter(hasActionFilter)
        .filter((a) => customActionKinds.includes(a.action))
        .map((a) => ({
          ...a,
          apply: userDefinedActions.find((ba) => ba.kind === a.action)!.apply,
          //                                                          ^ we already pre-filtered...
        }))) ??
    [];

  const knownActions = knownBuiltInActions.concat(knownCustomActions);

  const unknownActions = rawActions
    ?.filter(hasActionFilter)
    .filter(
      (a) =>
        !notOverridenBuiltInActionKinds.includes(a.action) &&
        !customActionKinds.includes(a.action),
    );

  return { knownActions, unknownActions };
};

export type EvaluationResultWithAction<T extends object> =
  EvaluationResult.Model &
    Pick<Action<T>, "apply"> & {
      action: string;
    };

export type ApplyActionsResult<
  T extends object,
  R extends EvaluationResultWithAction<T>[],
> = T & UnionToIntersection<ReturnType<R[number]["apply"]>>;

type UnionToIntersection<U> = (U extends any ? (k: U) => void : never) extends (
  k: infer I,
) => void
  ? I
  : never;

export const applyActions = async <T extends object>(
  obj: T,
  results: EvaluationResultWithAction<T>[],
) => {
  let copy = clone(obj);
  for (const r of results) {
    // apply might modify the object itself
    // instead of returning a copy,
    // but just in case...
    copy = await r.apply(r, copy);
  }
  return copy;
};

const stores: Map<string, StoreApi<RulesStore<any>>> = new Map();

export const getStore = <T extends object>(
  contextID: string,
  opts: RuleStoreOptions<T>,
) => {
  let store = stores.get(contextID);
  if (!store) {
    store = createRulesStore<T>(opts);
    stores.set(contextID, store);
  }
  return store as StoreApi<RulesStore<T>>;
};
