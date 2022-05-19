// https://www.pluralsight.com/guides/using-d3.js-inside-a-react-app
import React from "react";
import * as d3 from "d3";
import { BaseType } from "d3";

export const useD3 = (renderChartFn: Function, dependencies: any) => {
  const ref = React.useRef();

  React.useEffect(() => {
    renderChartFn(d3.select(ref.current as unknown as BaseType));
    return () => {};
  }, dependencies);

  return ref;
};
