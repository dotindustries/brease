import { TRPCError } from "@trpc/server";
import { z } from "zod";
import { createTRPCRouter, publicProcedure } from "~/server/api/trpc";
import { brease } from "~/server/brease";

export const exampleRouter = createTRPCRouter({
  breaseWebToken: publicProcedure.query(async () => {
    try {
      return await brease.Auth.getToken();
    } catch (e: any) {
      console.log(e.statusCode);
      console.log(e.detail);
      if (e.statusCode == 401) {
        throw new TRPCError({ code: "UNAUTHORIZED", message: e.detail });
      }
    }
    return null;
  }),
});
