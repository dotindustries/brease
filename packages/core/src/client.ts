import hash from "object-hash";
import {LRUCache} from "lru-cache";
import {CacheEntry, cachified} from "cachified";
import {encode, encodeToUint8Array} from "./encoder.js";
import {createPromiseClient, Interceptor, PromiseClient} from "@connectrpc/connect";
import {AuthService} from "@buf/dot_brease.connectrpc_es/brease/auth/v1/service_connect.js";
import {createConnectTransport} from "@connectrpc/connect-web";
import {ContextService} from "@buf/dot_brease.connectrpc_es/brease/context/v1/service_connect.js";
import {
    Action,
    And,
    Condition,
    Expression,
    Or,
    Rule, RuleRef,
    Target
} from "@buf/dot_brease.bufbuild_es/brease/rule/v1/model_pb.js";
import {JsonValue, Struct} from "@bufbuild/protobuf";
import {Result} from "./store.js";

// const logger: Interceptor = (next) => async (req) => {
//   return await next(req);
// };

const authInterceptor: (token: string, tokenScheme: 'JWT' | 'Bearer') => Interceptor = (token, tokenScheme) => (next) => async (req) => {
    req.header.set("Authorization", `${tokenScheme} ${token}`);
    return await next(req);
};

const headersInterceptor: (headers: Record<string, string>) => Interceptor = (headers) => (next) => async (req) => {
    for (const [key, value] of Object.entries(headers)) {
        req.header.set(key, value);
    }
    return await next(req);
};

const cache = new LRUCache<string, CacheEntry>({max: 1000});

export enum Environment {
    Development = "http://localhost:4400",
    Production = "https://api.brease.run",
}

export type ClientOptions = {
    accessToken: string;
    tokenType?: 'Bearer' | 'JWT';
    refreshToken?: string;
    baseUrl?: Environment | (string & {});
    headers?: Record<string, string>;
};

export type BreaseClient = {
    /**
     * An initialized brease SDK client
     */
    client: PromiseClient<typeof ContextService>;
    authClient: PromiseClient<typeof AuthService>;
    /**
     * Created an evaluateRules function with caching functionality within a specific contextID.
     * @param contextID
     * @param cacheTtl
     * @returns
     */
    createEvaluateRules: (
        contextID: string,
        cacheTtl?: number | undefined,
    ) => (
        input: any,
    ) => Promise<Result[]>;
}

export const newClient = (opts: ClientOptions): BreaseClient => {
    const baseUrl = opts.baseUrl ?? Environment.Production;
    const transport = createConnectTransport({
        baseUrl,
        interceptors: [
            // logger,
            authInterceptor(opts.accessToken, opts.tokenType ?? 'Bearer'),
            ...(opts.headers ? [headersInterceptor(opts.headers)] : []),
        ],
    })
    const authClient = createPromiseClient(AuthService, transport)
    const client = createPromiseClient(ContextService, transport)

    const createEvaluateRules =
        (contextID: string, cacheTtl?: number) =>
            (input: JsonValue) => {
                return cachified({
                    key: `${contextID}-${hash(input)}`,
                    cache,
                    async getFreshValue() {
                        const {results} = await client.evaluate({
                            contextId: contextID,
                            object: Struct.fromJson(input),
                            // overrideRules: [],
                            // overrideCode: ''
                        });
                        return results.map(({by, target, action}) => ({
                            action,
                            target: target &&{
                                id: target.id,
                                kind: target.kind,
                            },
                            by: by && {
                                id: by.id,
                                description: by.description
                            }
                        } satisfies Result));
                    },
                    ttl: cacheTtl,
                });
            };

    return {
        client,
        authClient,
        createEvaluateRules,
    };
};

export type Json =
    | string
    | number
    | boolean
    | null
    | Json[]
    | { [key: string]: Json };

export type ClientAnd = {
    and?: Array<ClientExpression>;
};

export type ClientOr = {
    or?: Array<ClientExpression>;
};

export type ClientConditionKey = Pick<Condition, "kind" | "base"> & {
    value?: Json;
};

export type ClientConditionRef = Pick<Condition, "kind" | "base"> & {
    value?: Json;
};

export type ClientCondition = {
    condition?: ClientConditionKey | ClientConditionRef;
};

export type ClientExpression = ClientAnd | ClientOr | ClientCondition

export type ClientTarget = Pick<Target, "id" | "kind"> & {
    value?: Json;
};

export type ClientAction = Pick<Action, 'kind'> & {
    target: ClientTarget;
}

export type ClientRule = {
    actions: Array<ClientAction>;
    description?: string;
    expression: ClientExpression;
    id: string;
};

export type ClientRuleRef = Pick<RuleRef, "id" | "description">

export const encodeClientRule = (rule: ClientRule): Rule => {
    const {expression, actions, ...rest} = rule;
    const encodedExpression = encodeExpression(expression);
    return new Rule({
        ...rest,
        expression: encodedExpression!, // the first expr is secured via ClientRule TS
        actions: actions.map(action => new Action({
            kind: action.kind,
            target: encodeTarget(action.target),
        }))
    })
};

const encodeTarget = (t: ClientTarget): Target => {
    const {value, ...rest} = t;
    return new Target({
        ...rest,
        value: encodeToUint8Array(encode(value)),
    })
};

const encodeExpression = (
    e?: ClientExpression,
): Expression | undefined => {
    if (!e) return undefined;

    if ("and" in e && e.and) {
        return new Expression({
            expr: {
                value: new And({
                    expression: e.and.map(a => encodeExpression(a)!) // find a better way to make TS happy
                }),
                case: 'and'
            }
        })
    } else if ("or" in e && e.or) {
        return new Expression({
            expr: {
                value: new Or({
                    expression: e.or.map(o => encodeExpression(o)!) // find a better way to make TS happy
                }),
                case: 'or'
            }
        })
    } else if ("condition" in e && e.condition) {
        const condition = encodeCondition(e.condition);
        if (!condition) return undefined;

        return new Expression({
            expr: {
                value: condition,
                case: 'condition'
            }
        })
    }
    console.error("Unknown expression type:", e);
    return undefined;
};

const encodeCondition = (c?: ClientConditionKey | ClientConditionRef) => {
    if (!c) return undefined;

    if ("value" in c) {
        const {value, ...rest} = c;
        return new Condition({
            ...rest,
            value: encodeToUint8Array(encode(value)),
        })
    }
    console.error("Unknown condition type:", c);
    return undefined;
};
