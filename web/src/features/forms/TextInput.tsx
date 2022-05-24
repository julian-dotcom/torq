import styles from "./textInput.module.scss";
import React from "react";

interface textInputProps {
  label?: string;
  value?: string;
  placeholder?: string;
  onChange?: (value: string) => void;
}
function TextInput({ label, value, placeholder, onChange }: textInputProps) {
  const [localValue, setLocalValue] = React.useState("" as string | undefined);
  React.useEffect(() => {
    setLocalValue(value);
  }, [value]);
  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement | HTMLInputElement>) => {
    setLocalValue(e.target.value);
    onChange && onChange(e.target.value);
  };
  return (
    <div style={{ marginBottom: "var(--form-margin-bottom)" }}>
      <div style={{ marginBottom: "var(--form-label-margin-bottom)" }}>
        <span>{label}</span>
      </div>
      <input
        type="text"
        placeholder={placeholder}
        className={styles.textInput}
        value={localValue}
        onChange={handleChange}
      />
    </div>
  );
}
export default TextInput;
