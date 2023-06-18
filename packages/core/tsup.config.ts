import { defineConfig } from "tsup";

export default defineConfig({
  dts: {
    resolve: true,
  },
  clean: true,
  sourcemap: true,
  shims: true,
  format: ["esm", "cjs"],
  treeshake: "smallest",
  external: ["zustand"],
});
