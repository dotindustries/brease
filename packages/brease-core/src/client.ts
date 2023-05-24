import { SDK } from "brease-sdk";

export const newClient = (serverURL?: string) => {
  const sdk = new SDK({ serverURL: serverURL });

  return sdk.contextID;
};
