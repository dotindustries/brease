# Multiple contexts

## Different contexts
Each context and their execution is scoped to its own store.
You can use as many different contexts in parallel as you need. For example:

```ts
const { result: order } = useRules("checkout", user);
const { result: order } = useRules("colors", user);
```

## Same context for multiple objects

If you have to run business rules for multiple objects of a context, you must add a distinct identifier:

```ts
const { result: order1 } = useRules("checkout", user1, { objectID: "u1" });
const { result: order2 } = useRules("checkout", user2, { objectID: "u2" });
```

Otherwise both `order1` and `order2` end up share the refer to `user2` with applied actions. They would use the same rules store and the last execution wins.
https://dotindustries.notion.site/Type-inheritance-0ac344df4a6c4821870c3fd6b374d0ed