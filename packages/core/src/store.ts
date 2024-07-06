import { StoreApi, createStore } from "zustand/vanilla";
import hash from "object-hash";
import { immer } from "zustand/middleware/immer";
import { clone } from "lodash-es";
import { $setAction, ApplyFunction } from "./actions.js";
import {EvaluateRequest} from "@buf/dot_brease.bufbuild_es/brease/context/v1/models_pb.js";
import { EvaluationResult } from "@buf/dot_brease.bufbuild_es/brease/rule/v1/model_pb.js";

export type EvaluateInput<T> = Pick<
  EvaluateRequest,
  "overrideCode" | "overrideRules"
> & {
  object: T;
};

export type FunctionKeys = "$set" | string;

export type FunctionMap<T extends object> = {
  [key: FunctionKeys]: ApplyFunction<T, any>;
};

// Hack to convert union types to intersection types
export type UnionToIntersection<U> = (
  U extends any ? (k: U) => void : never
) extends (k: infer I) => void
  ? I
  : never;

// Hack type to simplify resolved TS types
export type Resolve<T> = {
  [K in keyof T]: T[K];
};

export type RuleStoreOptions<T extends object, F extends FunctionMap<T>> = {
  evaluateRulesFn: (
    input: EvaluateInput<T>,
  ) => Promise<EvaluationResult[]>;
  userDefinedActions?: F;
  overrideCode?: string;
  overrideRules?: EvaluateRequest['overrideRules']
};

export interface RulesStore<T extends object, F extends FunctionMap<T>> {
  isExecuting: boolean;
  lastHash?: string;
  rawActions: EvaluationResult[];
  executeRules: (object: T) => void;
  result:
    | Resolve<T & UnionToIntersection<Awaited<ReturnType<F[keyof F]>>>>
    | undefined;
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
          overrideCode: overrideCode ?? '',
          overrideRules: overrideRules ?? [],
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

        // apply builtIn actions
        // @ts-ignore -- is assignable to the constraint of type F, but F could be instantiated with a different subtype of constraint FunctionMap<T>
        const functions: F = {
          $set: $setAction,
          ...userDefinedActions,
        };

        // apply recognized rules
        const result = await applyActions(object, rawActions, functions);

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
  rawActions: EvaluationResult[],
  fns: F,
): Promise<Resolve<T & UnionToIntersection<ReturnType<F[keyof F]>>>> => {
  const copy = clone(obj);
  const parts = await Promise.all(
    Object.values(fns).flatMap((fn) => rawActions.map((a) => fn(a, obj))),
  );
  return Object.assign({}, copy, ...parts) as never;
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
