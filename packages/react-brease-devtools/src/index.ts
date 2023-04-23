"use client";

import * as devtools from "./devtools";

export const ReactBreaseDevtools: (typeof devtools)["ReactBreaseDevtools"] =
  process.env.NODE_ENV !== "development"
    ? function () {
        return null;
      }
    : devtools.ReactBreaseDevtools;

export const ReactBreaseDevtoolsPanel: (typeof devtools)["ReactBreaseDevtoolsPanel"] =
  process.env.NODE_ENV !== "development"
    ? (function () {
        return null;
      } as any)
    : devtools.ReactBreaseDevtoolsPanel;
