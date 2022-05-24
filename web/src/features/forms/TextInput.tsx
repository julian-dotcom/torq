import styles from "./textInput.module.scss";
import React from "react";

interface textInputProps {
  label?: string;
  value?: string;
}
function TextInput({ label, value }: textInputProps) {
  const [localValue, setLocalValue] = React.useState("");
  React.useEffect(() => {
    value && setLocalValue(value);
  }, [value]);
  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement | HTMLInputElement>) => {
    if (e.target.value) {
      setLocalValue(e.target.value);
    }
  };
  return (
    <div style={{ marginBottom: "var(--form-margin-bottom)" }}>
      <div style={{ marginBottom: "var(--form-label-margin-bottom)" }}>
        <span>{label}</span>
      </div>
      <input type="text" className={styles.textInput} value={localValue} onChange={handleChange} />
    </div>
  );
}
export default TextInput;
