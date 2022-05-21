import styles from "./file.module.scss";

interface fileProps {
  label?: string;
}
function File({ label }: fileProps) {
  return (
    <>
      <div style={{ marginBottom: "var(--form-label-margin-bottom)" }}>
        <span>{label}</span>
      </div>
      <div className={styles.fileDropArea}>
        <span>Drop files here or click to select</span>
      </div>
    </>
  );
}

export default File;
