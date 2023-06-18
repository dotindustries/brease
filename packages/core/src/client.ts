import { BreaseSDK, Environment } from "@brease/sdk";

export type ClientOptions = {
  accessToken: string;
  refreshToken?: string;
  environment?: Environment;
};

export const newClient = ({
  environment,
  accessToken,
  refreshToken,
}: ClientOptions) => {
  const sdk = new BreaseSDK(refreshToken, accessToken);
  environment && sdk.setEnvironment(environment);
  return sdk;
};
