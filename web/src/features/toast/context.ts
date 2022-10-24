import { createContext } from "react";
import { addToastHandle } from "features/toast/Toasts";
import React from "react";
const ToastContext = createContext<React.MutableRefObject<addToastHandle | undefined> | null>(null);

// const ToastContext = createContext<addToastHandle>({
//   addToast: (_: string, __: toastCategory) => {
//     return;
//   },
// });
export default ToastContext;
