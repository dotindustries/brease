import {
  And,
  ApiEvaluateRulesResponse,
  BreaseSDK,
  ConditionKey,
  Environment,
  EvaluateRulesInput,
  Expression,
  Or,
  Rule,
  Target,
} from "@brease/sdk";
import hash from "object-hash";
import { LRUCache } from "lru-cache";
import { cachified, CacheEntry } from "cachified";
import { encode } from "./encoder.js";

const cache = new LRUCache<string, CacheEntry>({ max: 1000 });

export type ClientOptions = {
  accessToken: string;
  refreshToken?: string;
  environment?: Environment;
};

export const newClient = ({
  environment,
  accessToken,
  refreshToken,
}: ClientOptions): {
  /**
   * An initialized brease SDK client
   */
  sdk: BreaseSDK;
  /**
   * Created an evaluateRules function with caching functionality within a specific contextID.
   * @param contextID
   * @param cacheTtl
   * @returns
   */
  createEvaluateRules: (
    contextID: string,
    cacheTtl?: number | undefined,
  ) => (
    input: EvaluateRulesInput.Model,
  ) => Promise<ApiEvaluateRulesResponse.Results | undefined>;
} => {
  const sdk = new BreaseSDK(refreshToken, accessToken);
  environment && sdk.setEnvironment(environment);

  const createEvaluateRules =
    (contextID: string, cacheTtl?: number) =>
    (input: EvaluateRulesInput.Model) => {
      return cachified({
        key: `${contextID}-${hash(input)}`,
        cache,
        async getFreshValue() {
          const { results } = await sdk.Context.evaluateRules(input, contextID);
          return results;
        },
        ttl: cacheTtl,
      });
    };

  return {
    sdk,
    createEvaluateRules,
  };
};

export type Json =
  | string
  | number
  | boolean
  | null
  | Json[]
  | { [key: string]: Json };

export type ClientAnd = {
  and?: ClientExpression;
};

export type ClientOr = {
  or?: ClientExpression;
};

export type ClientConditionKey = Omit<ConditionKey.Model, "value"> & {
  value?: Json;
};

export type ClientConditionRef = Omit<ConditionKey.Model, "value"> & {
  value?: Json;
};

export type ClientCondition = {
  condition?: ClientConditionKey | ClientConditionRef;
};

export type ClientExpression = ClientAnd | ClientOr | ClientCondition;

export type ClientTarget = Omit<Target.Model, "value"> & {
  value?: Json;
};

export type ClientRule = {
  action: Rule.Action;
  description?: Rule.Description;
  expression: ClientExpression;
  id: Rule.Id;
  target: ClientTarget;
};

export const encodeClientRule = (rule: ClientRule): Rule.Model => {
  const { expression, target, ...rest } = rule;
  const encodedTarget = encodeTarget(target);
  const encodedExpression = encodeExpression(expression);
  return Object.assign(rest, {
    expression: encodedExpression!, // the first expr is secured via ClientRule TS
    target: encodedTarget,
  });
};

const encodeTarget = (t: ClientTarget) => {
  const { value, ...rest } = t;
  return Object.assign(rest, { value: encode(value) });
};

const encodeExpression = (
  e?: ClientExpression,
): Expression.Model | undefined => {
  if (!e) return undefined;

  if ("and" in e) {
    return Object.assign(e, { and: encodeExpression(e.and) }) as And.Model; // find a better way to make TS happy
  } else if ("or" in e) {
    return Object.assign(e, { or: encodeExpression(e.or) }) as Or.Model; // find a better way to make TS happy
  } else if ("condition" in e) {
    return Object.assign(e, {
      condition: encodeCondition(e.condition),
    });
  }
  console.error("Unknown expression type:", e);
  return undefined;
};

const encodeCondition = (c?: ClientConditionKey | ClientConditionRef) => {
  if (!c) return undefined;

  if ("value" in c) {
    const { value, ...rest } = c;
    return Object.assign(rest, { value: encode(value) });
  }
  console.error("Unknown condition type:", c);
  return undefined;
};
