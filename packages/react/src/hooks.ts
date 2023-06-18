import type {
  AddRuleInput,
  EvaluateRulesInput,
  ReplaceRuleInput,
  ApiAddRuleResponse,
  ApiReplaceRuleResponse,
  ApiEvaluateRulesResponse,
  ApiAllRulesResponse,
} from "@brease/core";
import { useContext, useMemo } from "react";
import { BreaseContext } from "./provider";

export const useRulesClient = () => {
  return useContext(BreaseContext);
};

export const useRuleContext = (
  contextID: string
): {
  evaluateRules: (
    input: EvaluateRulesInput.Model
  ) => Promise<ApiEvaluateRulesResponse.Model>;
  addRule: (input: AddRuleInput.Model) => Promise<ApiAddRuleResponse.Model>;
  getAllRules: (
    compileCode?: boolean | undefined
  ) => Promise<ApiAllRulesResponse.Model>;
  removeRule: (ruleID: string) => Promise<any>;
  replaceRule: (
    input: ReplaceRuleInput.Model
  ) => Promise<ApiReplaceRuleResponse.Model>;
} => {
  const client = useRulesClient();

  const operations = useMemo(() => {
    const evaluateRules = (input: EvaluateRulesInput.Model) => {
      return client.Context.evaluateRules(input, contextID);
    };

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
