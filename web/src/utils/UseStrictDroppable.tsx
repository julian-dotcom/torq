import { useEffect, useState } from "react";

// useStrictDroppable is a workaround for incorrect handling of React.StrictMode by react-beautiful-dnd
// https://github.com/atlassian/react-beautiful-dnd/issues/2396#issuecomment-1248018320
export const useStrictDroppable = (loading: boolean) => {
  const [enabled, setEnabled] = useState(false);

  useEffect(() => {
    let animation: number;

    if (!loading) {
      animation = requestAnimationFrame(() => setEnabled(true));
    }

    return () => {
      cancelAnimationFrame(animation);
      setEnabled(false);
    };
  }, [loading]);

  return [enabled];
};
