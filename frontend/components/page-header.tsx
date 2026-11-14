import React from "react";
import Link from "next/link";
import { ChevronRight } from "lucide-react";
import { cn } from "@/lib/utils";

export interface PageBreadcrumbItem {
  label: string;
  href?: string;
}

export interface PageHeaderProps {
  /** Workspace root shown leftmost in the hierarchy trail ("ORGANIZER" or "SPONSOR"). */
  workspace?: "organizer" | "sponsor";
  /** Hierarchy trail items after the workspace root. */
  breadcrumbs?: PageBreadcrumbItem[];
  /** Optional tags/badges rendered between the hierarchy and heading. */
  tags?: React.ReactNode;
  /** Page heading text (rendered as h1 without icons). */
  title: string;
  /** Optional subtitle/description below the heading. */
  description?: React.ReactNode;
  /** Optional action buttons or controls aligned to the right on wider screens. */
  actions?: React.ReactNode;
  className?: string;
}

const WORKSPACE_ROOT: Record<
  NonNullable<PageHeaderProps["workspace"]>,
  PageBreadcrumbItem
> = {
  organizer: { label: "Organizer", href: "/organizer" },
  sponsor: { label: "Sponsor", href: "/sponsor" },
};

const breadcrumbTextClass =
  "text-[10px] font-bold tracking-wider uppercase leading-none";

export function PageHeader({
  workspace,
  breadcrumbs,
  tags,
  title,
  description,
  actions,
  className,
}: PageHeaderProps) {
  const trail: PageBreadcrumbItem[] = [];
  if (workspace) {
    trail.push(WORKSPACE_ROOT[workspace]);
  }
  if (breadcrumbs?.length) {
    trail.push(...breadcrumbs);
  }

  return (
    <header className={cn("flex flex-col gap-3", className)}>
      {trail.length > 0 ? (
        <nav aria-label="Breadcrumb">
          <ol className="flex flex-wrap items-center gap-1.5">
            {trail.map((item, index) => {
              const isLast = index === trail.length - 1;
              const isSoleWorkspaceRoot = trail.length === 1 && Boolean(workspace);
              const showLink = Boolean(item.href) && !isLast;

              return (
                <React.Fragment key={`${item.label}-${index}`}>
                  {index > 0 ? (
                    <li aria-hidden="true" className="flex items-center">
                      <ChevronRight className="text-muted-foreground/60 size-3 shrink-0" />
                    </li>
                  ) : null}
                  <li className="flex min-w-0 items-center">
                    {showLink && item.href ? (
                      <Link
                        href={item.href}
                        className={cn(
                          breadcrumbTextClass,
                          "text-muted-foreground hover:text-foreground max-w-48 truncate transition-colors sm:max-w-xs"
                        )}
                      >
                        {item.label}
                      </Link>
                    ) : (
                      <span
                        aria-current={isLast ? "page" : undefined}
                        className={cn(
                          breadcrumbTextClass,
                          "max-w-48 truncate sm:max-w-xs",
                          isLast && !isSoleWorkspaceRoot
                            ? "text-foreground"
                            : "text-muted-foreground"
                        )}
                      >
                        {item.label}
                      </span>
                    )}
                  </li>
                </React.Fragment>
              );
            })}
          </ol>
        </nav>
      ) : null}

      {tags ? (
        <div className="flex flex-wrap items-center gap-2">{tags}</div>
      ) : null}

      <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div className="flex min-w-0 flex-1 flex-col gap-1.5">
          <h1 className="text-2xl font-bold tracking-tight wrap-anywhere sm:text-3xl">
            {title}
          </h1>
          {description ? (
            <p className="text-muted-foreground text-sm wrap-anywhere sm:text-base">
              {description}
            </p>
          ) : null}
        </div>
        {actions ? (
          <div className="flex shrink-0 flex-wrap items-center gap-2">
            {actions}
          </div>
        ) : null}
      </div>
    </header>
  );
}
