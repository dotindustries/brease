import {ClientAction, ConditionKind, encodeClientRule, Environment, newClient, Struct} from "@brease/core/src";
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

t.test("delete rule", async () => {
    const response = await brease.client.deleteRule({
        contextId,
        ruleId: sampleRule.id
    })
    t.ok(response);
});

t.test("list rules after delete", async () => {
    const response = await brease.client.listRules({contextId})
    t.equal(response.rules.length, 0);
});
