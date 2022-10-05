import {render, waitFor, waitForElementToBeRemoved} from "@testing-library/react";
import { Provider } from "react-redux";
import { BrowserRouter } from "react-router-dom";
import App from "./App";
import { store } from "./store/store";

test("renders login page", async ()  => {
  const { getByText } = render(
    <BrowserRouter>
      <Provider store={store}>
        <App />
      </Provider>
    </BrowserRouter>
  );

  const loading = getByText(/loading/i);
  expect(loading).toBeInTheDocument();

  await waitForElementToBeRemoved(loading);

  await waitFor(() => {
    expect(getByText(/login/i)).toBeInTheDocument()
  });
});
