import type {
    ClientRule, EvaluateInput,
    EvaluateRequest,
    FunctionMap,
    ListRulesResponse,
    Client,
    Resolve, Result,
    RuleStoreOptions,
    UnionToIntersection,
    VersionedRule
} from "@brease/core";
import {ContextService, encodeClientRule, getStore} from "@brease/core";
import {useContext, useEffect, useMemo} from "react";
import {BreaseContext} from "./provider.js";
import {useStore} from "zustand";

const useRuleStore = <T extends object, F extends FunctionMap<T>>(
    contextID: string,
    options: RuleStoreOptions<T, F>,
) => {
    return useStore(getStore<T, F>(contextID, options));
};

export default useRuleStore;

export const useRulesClient = (): Client<typeof ContextService> => {
    return useContext(BreaseContext)?.client;
};

export type UseRuleOptions<T extends object, F extends FunctionMap<T>> = {
    objectID?: string;
    cacheTtl?: number;
    userDefinedActions?: F;
    overrideCode?: string;
    overrideRules?: Array<ClientRule>;
};

export type ExecuteRulesOptions = Pick<
    EvaluateRequest,
    "overrideCode" | "overrideRules"
>;

export type UseRulesOutput<T extends object, F extends FunctionMap<T>> = {
    isLoading: boolean;
    rawActions: Result[];
    executeRules: (object: T, opts?: ExecuteRulesOptions) => void;
    result:
        | Resolve<T & UnionToIntersection<Awaited<ReturnType<F[keyof F]>>>>
        | undefined;
};

export const useRules = <T extends object, F extends FunctionMap<T>>(
    contextID: string,
    obj: T,
    opts: UseRuleOptions<T, F>,
): UseRulesOutput<T, F> => {
    const {evaluateRules} = useRuleContext<T>(contextID, opts.cacheTtl);
    const storeID = opts.objectID ? `${contextID}_${opts.objectID}` : contextID;
    const {
        executeRules,
        isExecuting: isLoading,
        rawActions,
        result,
    } = useRuleStore<T, F>(storeID, {
        evaluateRulesFn: evaluateRules,
        userDefinedActions: opts.userDefinedActions,
        overrideCode: opts.overrideCode,
        overrideRules: opts.overrideRules,
    });

    useEffect(() => {
        executeRules(obj);
    }, [obj]);

    return {
        isLoading,
        rawActions,
        result,
        executeRules,
    };
};

export type RuleContext<T> = {
    evaluateRules: (
        input: EvaluateInput<T>,
    ) => Promise<Result[]>;
    addRule: (input: ClientRule) => Promise<VersionedRule>;
    getAllRules: (
        compileCode?: boolean | undefined,
    ) => Promise<ListRulesResponse>;
    removeRule: (ruleID: string) => Promise<any>;
    replaceRule: (input: ClientRule) => Promise<VersionedRule>;
};

/**
 * Creates context scoped functions ready for execution.
 *
 * @param contextID the domain context of the ruleset
 * @param cacheTtl the number of milliseconds to cache the rule run results for. Defaults to Infinity
 * @returns
 */
export const useRuleContext = <T>(
    contextID: string,
    cacheTtl?: number,
): RuleContext<T> => {
    const {client, createEvaluateRules} = useContext(BreaseContext);

    const operations = useMemo(() => {
        const evaluateRules = createEvaluateRules(contextID, cacheTtl);

        const addRule = (rule: ClientRule) => {
            return client.createRule(
                {
                    contextId: contextID,

                    rule: encodeClientRule(rule),
                }
            );
        };

        const getAllRules = (compileCode?: boolean | undefined) => {
            // TODO: decode rules instead of raw output?
            return client.listRules({
                contextId: contextID,
                compileCode,
            });
        };

        const removeRule = (ruleID: string) => {
            return client.deleteRule({
                contextId: contextID,
                ruleId: ruleID,
            });
        };

        const replaceRule = (rule: ClientRule) => {
            return client.updateRule(
                {rule: encodeClientRule(rule), contextId: contextID},
            );
        };

        return {
            evaluateRules,
            addRule,
            getAllRules,
            removeRule,
            replaceRule,
        };
    }, [contextID]);

    return operations;
};
