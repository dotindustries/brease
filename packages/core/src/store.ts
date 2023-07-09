import { StoreApi, createStore } from "zustand/vanilla";
import hash from "object-hash";
import { immer } from "zustand/middleware/immer";
import { clone } from "lodash-es";
import { EvaluationResult } from "@brease/sdk";
import { ApiEvaluateRulesResponse, EvaluateRulesInput } from "@brease/core";
import { $setAction, ApplyFunction } from "./actions.js";

export type EvaluateInput<T> = Pick<
  EvaluateRulesInput.Model,
  "overrideCode" | "overrideRules"
> & {
  object: T;
};

export type FunctionKeys = "$set" | string;

export type FunctionMap<T extends object> = {
  [key: FunctionKeys]: ApplyFunction<T, any>;
};

// Helper type to convert union types to intersection types
export type UnionToIntersection<U> = (
  U extends any ? (k: U) => void : never
) extends (k: infer I) => void
  ? I
  : never;

export type RuleStoreOptions<T extends object, F extends FunctionMap<T>> = {
  evaluateRulesFn: (
    input: EvaluateRulesInput.Model,
  ) => Promise<ApiEvaluateRulesResponse.Results | undefined>;
  userDefinedActions?: F;
  overrideCode?: string;
  overrideRules?: EvaluateRulesInput.OverrideRules;
};

export interface RulesStore<T extends object, F extends FunctionMap<T>> {
  isExecuting: boolean;
  lastHash?: string;
  rawActions: EvaluationResult.Model[];
  executeRules: (object: T) => void;
  result: Awaited<T & UnionToIntersection<ReturnType<F[keyof F]>>> | undefined;
}

const createRulesStore = <T extends object, F extends FunctionMap<T>>({
  evaluateRulesFn,
  overrideCode,
  overrideRules,
  userDefinedActions,
}: RuleStoreOptions<T, F>) => {
  return createStore(
    immer<RulesStore<T, F>>((set, get) => ({
      isExecuting: false,
      rawActions: [],
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

        // nothing to execute
        if (!rawActions) {
          set({
            rawActions,
            isExecuting: false,
            result: undefined,
          });
          return;
        }

        // extract and apply builtIn actions
        const actions: F = {
          $set: $setAction,
          ...userDefinedActions,
        };

        // apply recognized rules
        const result = await applyActions(object, rawActions, actions);

        set({
          rawActions,
          isExecuting: false,
          result,
        });
      },
    })),
  );
};

export const applyActions = async <T extends object, F extends FunctionMap<T>>(
  obj: T,
  rawActions: EvaluationResult.Model[],
  fns: F,
) => {
  let copy = clone(obj);
  for (const action of rawActions) {
    if (!action.action) continue;
    const fn = fns[action.action];
    if (!fn) continue;
    const extension = await fn(action, copy);
    Object.assign(copy, extension);
  }
  return copy as T & UnionToIntersection<ReturnType<F[keyof F]>>;
};

const stores: Map<string, StoreApi<RulesStore<any, any>>> = new Map();

export const getStore = <T extends object, F extends FunctionMap<T>>(
  contextID: string,
  opts: RuleStoreOptions<T, F>,
) => {
  let store = stores.get(contextID);
  if (!store) {
    store = createRulesStore<T, F>(opts);
    stores.set(contextID, store);
  }
  return store as StoreApi<RulesStore<T, F>>;
};
