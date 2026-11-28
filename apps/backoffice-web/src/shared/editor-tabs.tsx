import { useId, useRef, type ReactNode } from "react";
import { Tab, Tabs } from "@mui/material";

export function EditorTabs({
  label,
  value,
  onChange,
  disabled,
  panels,
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  disabled?: boolean;
  panels: { id: string; label: string; content: ReactNode }[];
}) {
  const prefix = useId();
  const invalidTarget = useRef<HTMLElement | null>(null);
  return (
    <div className="admin-editor-tabs">
      <Tabs
        aria-label={label}
        value={value}
        onChange={(_, next: string) => onChange(next)}
        variant="scrollable"
        scrollButtons="auto"
        allowScrollButtonsMobile
      >
        {panels.map((panel) => (
          <Tab
            key={panel.id}
            value={panel.id}
            label={panel.label}
            disabled={disabled}
            id={`${prefix}-tab-${panel.id}`}
            aria-controls={`${prefix}-panel-${panel.id}`}
          />
        ))}
      </Tabs>
      {panels.map((panel) => (
        <div
          key={panel.id}
          role="tabpanel"
          id={`${prefix}-panel-${panel.id}`}
          aria-labelledby={`${prefix}-tab-${panel.id}`}
          hidden={value !== panel.id}
          className="admin-editor-tab-panel"
          onInvalidCapture={(event) => {
            if (value === panel.id) return;
            event.preventDefault();
            if (invalidTarget.current) return;
            const target = event.target as HTMLElement;
            invalidTarget.current = target;
            onChange(panel.id);
            requestAnimationFrame(() => {
              target.focus();
              invalidTarget.current = null;
            });
          }}
        >
          {panel.content}
        </div>
      ))}
    </div>
  );
}
