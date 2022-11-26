import View from "./View";
import React, { useState } from "react";
import { ViewInterface } from "./types";

export function useView<T>(
  views: Array<ViewInterface<T>>,
  defaultView: number
): [View<T>, React.Dispatch<React.SetStateAction<number>>] {
  const [updateCounter, viewUpdater] = useState(0);
  const [selectedView, setSelectedView] = useState(defaultView);
  return [new View(views[selectedView], updateCounter, viewUpdater), setSelectedView];
}
