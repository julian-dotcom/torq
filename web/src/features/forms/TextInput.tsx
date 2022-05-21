import styles from "./textInput.module.scss";

interface textInputProps {
  label?: string;
}
function TextInput({ label }: textInputProps) {
  return (
    <div style={{ marginBottom: "var(--form-margin-bottom)" }}>
      <div style={{ marginBottom: "var(--form-label-margin-bottom)" }}>
        <span>{label}</span>
      </div>
      <input type="text" className={styles.textInput} />
    </div>
  );
}
export default TextInput;
