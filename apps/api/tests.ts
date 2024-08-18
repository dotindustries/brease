import {
    ClientAction,
    ConditionKind,
    decodeClientRule,
    encodeClientRule,
    Environment,
    newClient,
    Struct
} from "@brease/core/src";
import t from 'tap'

const brease = newClient({
    accessToken: "asdf",
    baseUrl: Environment.Development,
    headers: {
        "x-org-id": "org_01h89qgxe5e7wregw6gb94d5p6"
    }
});
const contextId = 'asdf'
const sampleRule = {
    "id": "oyXnMAN-Qe3BA7CU3oAeU",
    "name": "New rule",
    "query": "pre_01httnykd6fjvt518c3yxvx3r8.prse_01hv6qqj1ve7zvpvq03ak1b3w8.contains(\"hobbit\")",
    "actions": [
        {
            "id": "oLoYA84KTLSKuUCOdGfTg",
            "kind": "hide-field",
            "target": "pre_01hv7ach32fd3vkzwv073t1qr4.prse_01hv7acy3hf3ct0y7b4bc9n4w0",
            "value": ""
        }
    ]
}
const sampleRule2 = {
    "id": "ouk_01h89qgxe5e7wregw6gb94d5p6",
    "name": "New rule 2",
    "query": "pre_01httnykd6fjvt518c3yxvx3r8.prse_01hv6qqj1ve7zvpvq03ak1b3w8.contains(\"hobbit\")",
    "actions": [
        {
            "id": "oLJYA84KTLSKuUCOdGfTg",
            "kind": "hide-field",
            "target": "pre_01hv7ach32fd3vkzwv073t1qr4.prse_01hv7acy3hf3ct0y7b4bc9n4w0",
            "value": ""
        }
    ]
}

// t.beforeEach(async _t => {
//     await new Promise(res => setTimeout(res, 1000))
// })

t.test("no rules in system", async () => {
    // curl equivalent:
    // curl -vv -X POST http://localhost:4400/brease.context.v1.ContextService/ListRules \
    //   -H "Authorization: Bearer asdf" \
    //   -H "x-org-id: org_01h89qgxe5e7wregw6gb94d5p6" \
    //   -H "Content-Type: application/json" --data '{"contextId": "asdf"}'
    const response = await brease.client.listRules({contextId})
    t.equal(response.rules.length, 0);
})

t.test("create rule", async () => {
    const actions: Array<ClientAction> = sampleRule.actions.map(a => ({
        kind: a.kind,
        target: {
            id: a.target,
            kind: 'jsonpath'
        },

    } satisfies ClientAction))
    const response = await brease.client.createRule({
        contextId,
        rule: encodeClientRule({
            id: sampleRule.id,
            actions,
            description: sampleRule.name,
            expression: {
                condition: {
                    kind: ConditionKind.cel,
                    base: {
                        case: 'key',
                        value: ""
                    },
                    value: sampleRule.query
                }
            }
        })
    })
    t.ok(response);
});

t.test("retrieve created rule", async () => {
    const response = await brease.client.listRules({contextId})
    t.equal(response.rules.length, 1);

    const clientRules = response.rules.map(r => decodeClientRule(r))
    t.equal(clientRules.length, response.rules.length);
    if (!('condition' in clientRules[0]!.expression)) {
        t.fail("failed to decode rule");
    } else {
        const condition = clientRules[0]!.expression.condition
        if (condition && 'base' in condition) {
         t.equal(condition.value, sampleRule.query);
        } else {
            t.fail("failed to decode condition");
        }
    }
});

t.test("raw evaluate rule", async () => {
    const obj = {
        pre_01httnykd6fjvt518c3yxvx3r8: {
            prse_01hv6qqj1ve7zvpvq03ak1b3w8: "hobbit"
        }
    }

    const response = await brease.client.evaluate({
        contextId,
        object: Struct.fromJson(obj),
    })
    t.ok(response);
    console.log(response);
    t.equal(response.results.length, 1)
    t.equal(response.results[0]!.action, "hide-field")


    const response2 = await brease.client.evaluate({
        contextId,
        object: Struct.fromJson(obj),
    })
    t.ok(response2)
})

t.test("create second rule", async () => {
    const actions: Array<ClientAction> = sampleRule2.actions.map(a => ({
        kind: a.kind,
        target: {
            id: a.target,
            kind: 'jsonpath'
        },

    } satisfies ClientAction))
    const response = await brease.client.createRule({
        contextId,
        rule: encodeClientRule({
            id: sampleRule2.id,
            actions,
            description: sampleRule2.name,
            expression: {
                condition: {
                    kind: ConditionKind.cel,
                    base: {
                        case: 'key',
                        value: ""
                    },
                    value: sampleRule2.query
                }
            }
        })
    })
    t.ok(response);
});

t.test("retrieve created rules", async () => {
    const response = await brease.client.listRules({contextId})
    t.equal(response.rules.length, 2);
    console.log(JSON.stringify(response.rules, null, 2));
    t.same(new Set(response.rules.map(r => r.id)), new Set([sampleRule2.id, sampleRule.id]));
});

t.test("raw evaluate rules", async () => {
    const obj = {
        pre_01httnykd6fjvt518c3yxvx3r8: {
            prse_01hv6qqj1ve7zvpvq03ak1b3w8: "hobbit"
        }
    }

    const response = await brease.client.evaluate({
        contextId,
        object: Struct.fromJson(obj),
    })
    t.ok(response);
    console.log(response);
    t.equal(response.results.length, 2)
    t.equal(response.results[0]!.action, "hide-field")
    t.equal(response.results[1]!.action, "hide-field")

})

t.test("delete rule", async () => {
    const response = await brease.client.deleteRule({
        contextId,
        ruleId: sampleRule.id
    })
    t.ok(response);
});

t.test("deleting one rule should not affect the other", async () => {
    const response = await brease.client.listRules({contextId})
    t.equal(response.rules.length, 1);
    t.equal(response.rules[0]!.id, sampleRule2.id);
});

t.test("delete a rule", async () => {
    const response = await brease.client.deleteRule({
        contextId,
        ruleId: sampleRule2.id
    })
    t.ok(response);
});

t.test("no rules after all delete", async () => {
    const response = await brease.client.listRules({contextId})
    t.equal(response.rules.length, 0);
});
