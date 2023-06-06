import * as React from "react";
import { PropsWithChildren, createContext, useMemo } from "react";
import { newClient, type ClientOptions, BreaseSDK } from "@brease/core";

export const BreaseContext = createContext(new BreaseSDK(undefined, ""));

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
