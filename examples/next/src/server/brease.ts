import {BreaseClient, newClient} from "@brease/core";
import { env } from "~/env.mjs";

const globalForBrease = globalThis as unknown as {
  brease: BreaseClient | undefined;
};

export const brease =
  globalForBrease.brease ?? newClient({ accessToken: env.BREASE_TOKEN });

if (env.NODE_ENV !== "production") globalForBrease.brease = brease;
