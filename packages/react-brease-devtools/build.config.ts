import { defineBuildConfig } from "unbuild";

export default defineBuildConfig({
  // If entries is not provided, will be automatically inferred from package.json
  entries: [
    // default
    "./src/index",
    // "./src/useSyncExternalStore.native",
  ],

  // Change outDir, default is 'dist'
  outDir: "build",

  // Generates .d.ts declaration file
  declaration: true,
});
