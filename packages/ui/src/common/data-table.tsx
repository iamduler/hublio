"use client";

import * as React from "react";
import { cn } from "../lib/utils";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../ui/table";
import { EmptyState } from "../ui/empty-state";
import { LoadingState } from "../ui/loading-state";
import { Pagination } from "./pagination";
import { SortIcon, toggleSort, type SortState } from "./sort-icon";

export type DataTableColumn<T> = {
  id: string;
  header: React.ReactNode;
  cell: (row: T, index: number) => React.ReactNode;
  sortable?: boolean;
  /** Used for client-side sort when `mode="client"`. */
  sortAccessor?: (row: T) => string | number | boolean | null | undefined;
  className?: string;
  headerClassName?: string;
};

export type DataTableProps<T> = {
  columns: DataTableColumn<T>[];
  rows: T[];
  getRowId: (row: T) => string;
  /** Toolbar: search + FilterSelects etc. */
  toolbar?: React.ReactNode;
  /** Optional bulk-action bar between toolbar and table. */
  bulkBar?: React.ReactNode;
  /** Empty state when no rows after filtering. */
  emptyTitle?: React.ReactNode;
  emptyDescription?: React.ReactNode;
  emptyAction?: React.ReactNode;
  loading?: boolean;
  /** Client sorts/slices `rows`; server expects `rows` already paged and uses `total`. */
  mode?: "client" | "server";
  sort?: SortState;
  onSortChange?: (sort: SortState) => void;
  page?: number;
  pageSize?: number;
  onPageChange?: (page: number) => void;
  /** Required for server mode; optional for client (defaults to filtered length). */
  total?: number;
  /** Replace default Pagination (e.g. cursor “Load more”). Rendered top + bottom when set. */
  pagination?: React.ReactNode;
  /** Hide both Pagination slots. */
  hidePagination?: boolean;
  onRowClick?: (row: T) => void;
  className?: string;
  paginationShowingLabel?: (from: number, to: number, total: number) => string;
  prevLabel?: string;
  nextLabel?: string;
};

function compareValues(
  a: string | number | boolean | null | undefined,
  b: string | number | boolean | null | undefined,
): number {
  if (a == null && b == null) return 0;
  if (a == null) return 1;
  if (b == null) return -1;
  if (typeof a === "boolean" && typeof b === "boolean") {
    return Number(a) - Number(b);
  }
  if (typeof a === "number" && typeof b === "number") return a - b;
  return String(a).localeCompare(String(b), undefined, {
    numeric: true,
    sensitivity: "base",
  });
}

export function DataTable<T>({
  columns,
  rows,
  getRowId,
  toolbar,
  bulkBar,
  emptyTitle = "No results",
  emptyDescription,
  emptyAction,
  loading,
  mode = "client",
  sort: sortProp,
  onSortChange,
  page: pageProp = 1,
  pageSize = 10,
  onPageChange,
  total: totalProp,
  pagination: paginationSlot,
  hidePagination,
  onRowClick,
  className,
  paginationShowingLabel,
  prevLabel,
  nextLabel,
}: DataTableProps<T>) {
  const [internalSort, setInternalSort] = React.useState<SortState>(null);
  const [internalPage, setInternalPage] = React.useState(1);

  const sort = sortProp !== undefined ? sortProp : internalSort;
  const page = onPageChange ? pageProp : internalPage;

  const setSort = (next: SortState) => {
    if (onSortChange) onSortChange(next);
    else setInternalSort(next);
  };

  const setPage = (next: number) => {
    if (onPageChange) onPageChange(next);
    else setInternalPage(next);
  };

  const processed = React.useMemo(() => {
    if (mode === "server") return rows;
    let list = [...rows];
    if (sort) {
      const col = columns.find((c) => c.id === sort.column);
      const accessor = col?.sortAccessor;
      if (accessor) {
        list.sort((ra, rb) => {
          const cmp = compareValues(accessor(ra), accessor(rb));
          return sort.direction === "asc" ? cmp : -cmp;
        });
      }
    }
    return list;
  }, [rows, mode, sort, columns]);

  const total =
    mode === "server"
      ? (totalProp ?? rows.length)
      : (totalProp ?? processed.length);

  const pageRows =
    mode === "server"
      ? processed
      : processed.slice((page - 1) * pageSize, page * pageSize);

  React.useEffect(() => {
    if (mode === "client" && page > 1) {
      const maxPage = Math.max(1, Math.ceil(total / pageSize));
      if (page > maxPage) setPage(maxPage);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps -- only clamp when total shrinks
  }, [total, pageSize, mode]);

  const from = total === 0 ? 0 : (page - 1) * pageSize + 1;
  const to = Math.min(page * pageSize, total);
  const showingLabel = paginationShowingLabel?.(from, to, total);

  const paginationNode =
    hidePagination || total === 0
      ? null
      : (paginationSlot ?? (
        <Pagination
          page={page}
          total={total}
          perPage={pageSize}
          onChange={setPage}
          showingLabel={showingLabel}
          prevLabel={prevLabel}
          nextLabel={nextLabel}
        />
      ));

  return (
    <div
      className={cn(
        "overflow-hidden rounded-md border border-(--line) bg-(--white)",
        className,
      )}
    >
      {toolbar ? (
        <div className="flex flex-col gap-2 border-b border-(--line) px-4 py-3 sm:flex-row sm:flex-wrap sm:items-center">
          {toolbar}
        </div>
      ) : null}
      {bulkBar}
      {paginationNode}
      {loading ? (
        <LoadingState className="py-16" />
      ) : pageRows.length === 0 ? (
        <EmptyState
          size="sm"
          title={emptyTitle}
          description={emptyDescription}
          action={emptyAction}
          className="py-12"
        />
      ) : (
        <Table>
          <TableHeader>
            <TableRow className="hover:bg-transparent">
              {columns.map((col) => {
                const active = sort?.column === col.id;
                return (
                  <TableHead
                    key={col.id}
                    className={cn(
                      "h-9 bg-(--surface) px-4 text-[11px] font-semibold tracking-wider text-(--muted-clr)",
                      col.headerClassName,
                    )}
                  >
                    {col.sortable ? (
                      <button
                        type="button"
                        className="uppercase inline-flex items-center gap-1 transition-colors hover:text-(--ink)"
                        onClick={() => setSort(toggleSort(sort, col.id))}
                      >
                        {col.header}
                        <SortIcon
                          active={!!active}
                          direction={active ? sort?.direction : undefined}
                        />
                      </button>
                    ) : (
                      col.header
                    )}
                  </TableHead>
                );
              })}
            </TableRow>
          </TableHeader>
          <TableBody>
            {pageRows.map((row, index) => (
              <TableRow
                key={getRowId(row)}
                className={cn(
                  "group transition-colors hover:bg-(--surface)",
                  onRowClick && "cursor-pointer",
                )}
                onClick={onRowClick ? () => onRowClick(row) : undefined}
              >
                {columns.map((col) => (
                  <TableCell
                    key={col.id}
                    className={cn("px-4 py-3", col.className)}
                  >
                    {col.cell(row, index)}
                  </TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
      {paginationNode && !loading && pageRows.length > 0 ? paginationNode : null}
    </div>
  );
}
