import clone from "clone";
import { ColumnMetaData } from "features/table/types";
import { AndClause, Clause } from "../sidebar/sections/filter/filter";
import { OrderBy } from "../sidebar/sections/sort/SortSection";
import { AllViewsResponse, ViewInterface, ViewResponse } from "./types";

export default class View<T> {
  view: ViewInterface<T>;
  page: keyof AllViewsResponse;
  id: number | undefined;
  allColumns: Array<ColumnMetaData<T>>;
  renderCallback: React.Dispatch<React.SetStateAction<number>>;
  renderCount = 0;
  saved: boolean;

  constructor(
    viewResponse: ViewResponse<T>,
    allColumns: Array<ColumnMetaData<T>>,
    renderCount: number,
    render: React.Dispatch<React.SetStateAction<number>>,
    saved?: boolean
  ) {
    this.view = viewResponse.view;
    this.allColumns = allColumns;
    this.renderCallback = render;
    this.renderCount = renderCount;
    this.id = viewResponse.id;
    this.page = viewResponse.page;
    this.saved = saved || true;
  }

  get columns(): ColumnMetaData<T>[] {
    return this.view.columns;
  }

  get sortBy(): OrderBy[] | undefined {
    return this.view.sortBy;
  }

  get filters(): Clause {
    return this.view.filters || new AndClause();
  }

  set filters(filter) {
    this.view.filters = filter;
  }

  get groupBy(): "channels" | "peers" | undefined {
    return this.view.groupBy;
  }

  set groupBy(groupBy: "channels" | "peers" | undefined) {
    this.view.groupBy = groupBy;
    this.render();
  }

  //------------------ Columns ------------------

  addColumn = (column: ColumnMetaData<T>) => {
    this.view.columns.push(column);
    this.render();
  };

  updateColumn = (column: ColumnMetaData<T>, index: number) => {
    if (this.view.columns) {
      this.view.columns[index] = column;
      this.render();
    }
  };

  updateAllColumns = (columns: Array<ColumnMetaData<T>>) => {
    this.view.columns = columns;
    this.render();
  };

  moveColumn = (fromIndex: number, toIndex: number) => {
    if (this.view.columns) {
      const column = this.view.columns[fromIndex];
      this.view.columns.splice(fromIndex, 1);
      this.view.columns.splice(toIndex, 0, column);
      this.render();
    }
  };

  removeColumn = (index: number) => {
    if (this.view.columns) {
      this.view.columns.splice(index, 1);
      this.render();
    }
  };

  //------------------ SortBy ------------------

  addSortBy = (sortBy: OrderBy) => {
    this.view.sortBy = this.view.sortBy || [];
    this.view.sortBy = [...this.view.sortBy, sortBy];

    this.render();
  };

  updateSortBy = (update: OrderBy, index: number) => {
    if (this.view?.sortBy) {
      this.view.sortBy = [
        ...this.view.sortBy.slice(0, index),
        update,
        ...this.view.sortBy.slice(index + 1, this.view.sortBy.length),
      ];
    }

    this.render();
  };

  updateAllSortBy = (sortBy: Array<OrderBy>) => {
    this.view.sortBy = sortBy;

    this.render();
  };

  moveSortBy = (fromIndex: number, toIndex: number) => {
    if (this.view.sortBy) {
      const sortBy = clone(this.view.sortBy);
      const item = sortBy[fromIndex];
      sortBy.splice(fromIndex, 1);
      sortBy.splice(toIndex, 0, item);
      this.view.sortBy = sortBy;
    }
  };

  removeSortBy = (index: number) => {
    this.view.sortBy = this.view.sortBy || [];
    this.view.sortBy = [
      ...this.view.sortBy.slice(0, index),
      ...this.view.sortBy.slice(index + 1, this.view.sortBy.length),
    ];

    this.render();
  };

  //------------------ Filters ------------------
  // Filters are handled in the filters component. It can be refactored later, but for now it is fine.

  updateFilters = (filters: Clause) => {
    this.view.filters = filters;

    this.render();
  };

  //------------------ Render ------------------
  // This is a callback from the parent component to make react render. Used when the view is updated.
  public render() {
    this.renderCallback(this.renderCount + 1);
  }
}
