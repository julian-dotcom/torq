import View from "./View";
import React, { useState } from "react";
import { AllViewsResponse, ViewInterface, ViewResponse } from "./types";
import { useGetTableViewsQuery } from "./viewsApiSlice";
import { ColumnMetaData } from "../table/types";

export function useView<T>(
  page: keyof AllViewsResponse,
  allColumns: Array<ColumnMetaData<T>>,
  defaultView: number,
  viewTemplate: ViewInterface<T>
): [View<T>, React.Dispatch<React.SetStateAction<number>>, boolean, Array<ViewResponse<T>>] {
  const allViews = useGetTableViewsQuery<{
    data: AllViewsResponse;
    isLoading: boolean;
    isFetching: boolean;
    isUninitialized: boolean;
    isSuccess: boolean;
  }>();

  const [updateCounter, viewUpdater] = useState(0);
  const [selectedView, setSelectedView] = useState(defaultView);
  const defualtView = { page: page, view: viewTemplate, id: undefined, viewOrder: 0 };

  if (allViews.data === undefined) {
    return [new View(defualtView, allColumns, updateCounter, viewUpdater), setSelectedView, true, [defualtView]];
  }

  const views = JSON.parse(JSON.stringify(allViews.data)) as typeof allViews.data;

  const all = views[page] ? (views[page] as Array<ViewResponse<T>>) : [defualtView];

  const view = new View(all[selectedView], allColumns, updateCounter, viewUpdater);

  return [view, setSelectedView, allViews.isSuccess, all];
}
