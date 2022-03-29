import { MouseEventHandler } from "react";
import classNames from "classnames";
import styled from "@emotion/styled";

const StyledButton = styled.button`
  background: ${(props: any) => (props.isOpen ? "#ECFAF8" : "none")};
  border: ${(props: any) => (props.isOpen ? "1px solid #DDF6F5" : "none")};
  border-radius: 2px;
  display: grid;
  grid-auto-flow: column;
  align-items: center;
  grid-column-gap: 5px;
  padding: 5px 10px;
  cursor: pointer;
  &.small,
  &.small .icon {
    font-size: var(--font-size-small);
  }
  @media only screen and (max-width: 900px) {
    &.small-tablet,
    &.small-tablet .icon {
      font-size: var(--font-size-small);
    }
    &.collapse-tablet {
      padding: 5px 0.5em;
      .text {
        display: none;
      }
    }
  }
`;

function DefaultButton(props: {
  text: string;
  icon?: any;
  onClick?: MouseEventHandler<HTMLButtonElement> | undefined;
  className?: string;
  isOpen?: boolean;
}) {
  return (
    <StyledButton
      //@ts-expect-error
      isOpen={props.isOpen}
      className={classNames("button", props.className)}
      onClick={props.onClick}
    >
      {props.icon && <div className="icon">{props.icon}</div>}
      <div className="text">{props.text}</div>
    </StyledButton>
  );
}

export default DefaultButton;
