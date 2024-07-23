import {z} from "zod";
import {brease} from "~/server/brease";

import {createTRPCRouter, publicProcedure} from "~/server/api/trpc";
import {TRPCError} from "@trpc/server";
import {isJsonPath} from "@brease/core/src";

// Mocked DB
interface Post {
    id: number;
    name: string;
}

const posts: Post[] = [
    {
        id: 1,
        name: "Hello World",
    },
];

export const postRouter = createTRPCRouter({
    hello: publicProcedure
        .input(z.object({text: z.string()}))
        .query(({input}) => {
            const isIt = isJsonPath('$.a.b')
            return {
                greeting: `Hello ${input.text}`,
                isJsonPath: isIt
            };
        }),

    breaseTest: publicProcedure.query(async () => {
        try {
            return await brease.authClient.getToken({});
        } catch (e: any) {
            console.log(e.statusCode);
            console.log(e.detail);
            if (e.statusCode == 401) {
                throw new TRPCError({code: "UNAUTHORIZED", message: e.detail});
            }
        }
        return null;
    }),

    create: publicProcedure
        .input(z.object({name: z.string().min(1)}))
        .mutation(async ({input}) => {
            const post: Post = {
                id: posts.length + 1,
                name: input.name,
            };
            posts.push(post);
            return post;
        }),

    getLatest: publicProcedure.query(() => {
        return posts.at(-1) ?? null;
    }),
});
