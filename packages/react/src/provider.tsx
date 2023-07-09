import { PropsWithChildren, createContext, useMemo } from "react";
import {
  newClient,
  type ClientOptions,
  BreaseSDK,
  ApiEvaluateRulesResponse,
  EvaluateRulesInput,
} from "@brease/core";

export const BreaseContext = createContext<ReturnType<typeof newClient>>(
  createNoOpClient(),
);

export type BreaseProviderProps = PropsWithChildren<ClientOptions>;

export const BreaseProvider = ({
  children,
  ...clientOpts
}: BreaseProviderProps) => {
  const client = useMemo(() => {
    return newClient(clientOpts);
  }, [clientOpts]);

  return (
    <BreaseContext.Provider value={client}>{children}</BreaseContext.Provider>
  );
};

function createNoOpClient(): {
  sdk: BreaseSDK;
  createEvaluateRules: (
    contextID: string,
    cacheTtl?: number | undefined,
  ) => (
    input: EvaluateRulesInput.Model,
  ) => Promise<ApiEvaluateRulesResponse.Results | undefined>;
} {
  const sdk = new BreaseSDK(undefined, "");
  return {
    sdk,
    createEvaluateRules:
      (contextID: string) => async (input: EvaluateRulesInput.Model) => {
        const { results } = await sdk.Context.evaluateRules(input, contextID);
        return results;
      },
  };
}
