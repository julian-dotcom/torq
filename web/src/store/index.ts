import { createStore } from 'redux';

type storeType = { navHidden: boolean }

const initialState: storeType = {
  navHidden: false
}

const showNavReducer = (state: storeType = initialState, action: any) => {
  if (action.type === 'toggleNav') {
    return { navHidden: !state.navHidden }
  }

  return state;
};

const store = createStore(showNavReducer);

export default store;
