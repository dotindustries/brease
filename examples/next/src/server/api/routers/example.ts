import { z } from "zod";
import { createTRPCRouter, publicProcedure } from "~/server/api/trpc";
import { brease } from "~/server/brease";

export const exampleRouter = createTRPCRouter({
  breaseWebToken: publicProcedure.query(async () => {
    return await brease.Auth.getToken();
  }),
});
