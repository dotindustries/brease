import { PropsWithChildren, createContext, useMemo } from "react";
import {
  newClient,
  type ClientOptions, BreaseClient,
} from "@brease/core";

export const BreaseContext = createContext<BreaseClient>(
  newClient({accessToken: ''}),
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

