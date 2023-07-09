import {
  ApiEvaluateRulesResponse,
  BreaseSDK,
  Environment,
  EvaluateRulesInput,
} from "@brease/sdk";
import hash from "object-hash";
import { LRUCache } from "lru-cache";
import { cachified, CacheEntry } from "cachified";

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
