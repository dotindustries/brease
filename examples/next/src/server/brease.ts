import { BreaseSDK, newClient } from "@brease/core";
import { env } from "~/env.mjs";
import axios from "axios";

if (env.DEBUG) {
  axios.interceptors.request.use(
    (config) => {
      console.log("Outgoing Request Headers:", config.headers);
      return config;
    },
    (error) => {
      console.error("Error in request interceptor:", error);
      return Promise.reject(error);
    }
  );
}

const globalForBrease = globalThis as unknown as {
  brease: BreaseSDK | undefined;
};

export const brease =
  globalForBrease.brease ?? newClient({ accessToken: env.BREASE_TOKEN });

if (env.NODE_ENV !== "production") globalForBrease.brease = brease;
