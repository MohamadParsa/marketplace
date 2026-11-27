import { useId, useState, useEffect, type ReactNode } from "react";
import { useLocation } from "react-router-dom";
import { Button } from "@mui/material";

export function FilterPanel({
  children,
  onSearch,
  title = "جست‌وجو و فیلتر",
  className = "",
}: {
  children: ReactNode;
  onSearch?: () => void;
  title?: string;
  className?: string;
}) {
  const [open, setOpen] = useState(false);
  const { pathname } = useLocation();
  useEffect(() => setOpen(false), [pathname]);
  const id = useId();
  return (
    <section className={`admin-filter-panel ${className}`}>
      <Button
        className="admin-filter-toggle"
        aria-expanded={open}
        aria-controls={id}
        onClick={() => setOpen(!open)}
      >
        <span>{title}</span>
        <span aria-hidden="true">{open ? "−" : "+"}</span>
      </Button>
      <div id={id} hidden={!open} className="admin-filter-body">
        {onSearch ? (
          <form
            onSubmit={(event) => {
              event.preventDefault();
              onSearch();
            }}
          >
            {children}
            <Button type="submit" variant="contained">
              جست‌وجو
            </Button>
          </form>
        ) : (
          children
        )}
      </div>
    </section>
  );
}
