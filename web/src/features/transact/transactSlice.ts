import React from "react";
import { createSlice } from "@reduxjs/toolkit";
import { ColumnMetaData } from "../table/Table";
import { SortByOptionType } from "../sidebar/sections/sort/SortSectionOld";

export interface ViewInterface {
  title: string;
  id?: number;
  saved: boolean;
  filters?: any;
  columns?: ColumnMetaData[];
  order?: SortByOptionType[];
}

export interface initialStateProps {
  views: ViewInterface[];
}

const initialState: initialStateProps = {
  views: [],
};

export const paymentsSlice = createSlice({
  name: "payments",
  initialState,
  reducers: {},
});
