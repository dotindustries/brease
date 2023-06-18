import { BreaseSDK, newClient } from "@brease/core";
import { env } from "~/env.mjs";

const globalForBrease = globalThis as unknown as {
  brease: BreaseSDK | undefined;
};

export const brease =
  globalForBrease.brease ?? newClient({ accessToken: env.BREASE_TOKEN });

if (env.NODE_ENV !== "production") globalForBrease.brease = brease;
