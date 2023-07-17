import { createTypedSetAction, getStore } from "@brease/core";
import type {
  AddRuleInput,
  EvaluateRulesInput,
  ReplaceRuleInput,
  ApiAddRuleResponse,
  ApiReplaceRuleResponse,
  ApiEvaluateRulesResponse,
  ApiAllRulesResponse,
  RuleStoreOptions,
  EvaluationResult,
  FunctionMap,
  UnionToIntersection,
  Resolve,
} from "@brease/core";
import { useContext, useEffect, useMemo } from "react";
import { BreaseContext } from "./provider.js";
import { useStore } from "zustand";

const useRuleStore = <T extends object, F extends FunctionMap<T>>(
  contextID: string,
  options: RuleStoreOptions<T, F>,
) => {
  return useStore(getStore<T, F>(contextID, options));
};

export default useRuleStore;

export const useRulesClient = () => {
  return useContext(BreaseContext)?.sdk;
};

export type UseRuleOptions<T extends object, F extends FunctionMap<T>> = {
  objectID?: string;
  cacheTtl?: number;
  userDefinedActions?: F;
  overrideCode?: string;
  overrideRules?: EvaluateRulesInput.OverrideRules;
};

export type ExecuteRulesOptions = Pick<
  EvaluateRulesInput.Model,
  "overrideCode" | "overrideRules"
>;

export type UseRulesOutput<T extends object, F extends FunctionMap<T>> = {
  isLoading: boolean;
  rawActions: EvaluationResult.Model[];
  executeRules: (object: T, opts?: ExecuteRulesOptions) => void;
  result:
    | Resolve<T & UnionToIntersection<Awaited<ReturnType<F[keyof F]>>>>
    | undefined;
};

export const useRules = <T extends object, F extends FunctionMap<T>>(
  contextID: string,
  obj: T,
  opts: UseRuleOptions<T, F>,
): UseRulesOutput<T, F> => {
  const { evaluateRules } = useRuleContext(contextID, opts.cacheTtl);
  const storeID = opts.objectID ? `${contextID}_${opts.objectID}` : contextID;
  const {
    executeRules,
    isExecuting: isLoading,
    rawActions,
    result,
  } = useRuleStore<T, F>(storeID, {
    evaluateRulesFn: evaluateRules,
    userDefinedActions: opts.userDefinedActions,
    overrideCode: opts.overrideCode,
    overrideRules: opts.overrideRules,
  });

  useEffect(() => {
    executeRules(obj);
  }, [obj]);

  return {
    isLoading,
    rawActions,
    result,
    executeRules,
  };
};

export type RuleContext = {
  evaluateRules: (
    input: EvaluateRulesInput.Model,
  ) => Promise<ApiEvaluateRulesResponse.Results | undefined>;
  addRule: (input: AddRuleInput.Model) => Promise<ApiAddRuleResponse.Model>;
  getAllRules: (
    compileCode?: boolean | undefined,
  ) => Promise<ApiAllRulesResponse.Model>;
  removeRule: (ruleID: string) => Promise<any>;
  replaceRule: (
    input: ReplaceRuleInput.Model,
  ) => Promise<ApiReplaceRuleResponse.Model>;
};

/**
 * Creates context scoped functions ready for execution.
 *
 * @param contextID the domain context of the ruleset
 * @param cacheTtl the number of milliseconds to cache the rule run results for. Defaults to Infinity
 * @returns
 */
export const useRuleContext = (
  contextID: string,
  cacheTtl?: number,
): RuleContext => {
  const { sdk: client, createEvaluateRules } = useContext(BreaseContext);

  const operations = useMemo(() => {
    const evaluateRules = createEvaluateRules(contextID, cacheTtl);

    const addRule = (input: AddRuleInput.Model) => {
      return client.Context.addRule(input, contextID);
    };

    const getAllRules = (compileCode?: boolean | undefined) => {
      return client.Context.getAllRules(contextID, {
        compileCode,
      });
    };

    const removeRule = (ruleID: string) => {
      return client.Context.removeRule(contextID, ruleID);
    };

    const replaceRule = (input: ReplaceRuleInput.Model) => {
      return client.Context.replaceRule(input, contextID, input.rule.id);
    };

    return {
      evaluateRules,
      addRule,
      getAllRules,
      removeRule,
      replaceRule,
    };
  }, [contextID]);

  return operations;
};
