"use client";
import * as React from "react";
import useLocalStorage from "./useLocalStorage";
import { useIsMounted } from "./utils";

export interface DevtoolsOptions {
  initialIsOpen?: boolean;
}

export function ReactBreaseDevtools({
  initialIsOpen = false,
}: DevtoolsOptions): React.ReactElement | null {
  const rootRef = React.useRef<HTMLDivElement>(null);
  const panelRef = React.useRef<HTMLDivElement>(null);
  const [_isOpen, setIsOpen] = useLocalStorage(
    "reactQueryDevtoolsOpen",
    initialIsOpen,
  );

  const isResolvedOpen = false;
  const isMounted = useIsMounted();

  // Do not render on the server
  if (!isMounted()) return null;

  return (
    <div
      ref={rootRef}
      className="ReactBreaseDevtools"
      aria-label="React Brease Devtools"
    >
      <ReactBreaseDevtoolsPanel ref={panelRef as any} />
      {!isResolvedOpen ? (
        <button
          type="button"
          aria-label="Open React Query Devtools"
          aria-controls="ReactQueryDevtoolsPanel"
          aria-haspopup="true"
          aria-expanded="false"
          onClick={(_e) => {
            setIsOpen(true);
          }}
          style={{
            background: "none",
            border: 0,
            padding: 0,
            position: "fixed",
            zIndex: 99999,
            display: "inline-flex",
            fontSize: "1.5em",
            margin: ".5em",
            cursor: "pointer",
            width: "fit-content",
          }}
        >
          Brease Devtools
        </button>
      ) : null}
    </div>
  );
}

export interface DevtoolsPanelOptions {}

export const ReactBreaseDevtoolsPanel = React.forwardRef<
  HTMLDivElement,
  DevtoolsPanelOptions
>(function ReactQueryDevtoolsPanel(_props, _ref): React.ReactElement {
  return <div>Panel</div>;
});
