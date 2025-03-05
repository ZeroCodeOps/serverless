import { useEffect, useState } from 'react';
import Editor from '@monaco-editor/react';

interface CodeEditorProps {
  language: string;
  value: string;
  onChange: (value: string) => void;
}

const CodeEditor: React.FC<CodeEditorProps> = ({ language, value, onChange }) => {
  const [mounted, setMounted] = useState<boolean>(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  if (!mounted) {
    return (
      <div className="h-96 border rounded-md bg-muted flex items-center justify-center text-muted-foreground">
        Loading editor...
      </div>
    );
  }

  return (
    <div className="code-editor-container">
      <Editor
        height="700px"
        language={language}
        value={value}
        onChange={(value: any) => onChange(value || '')}
        theme="vs-dark"
        options={{
          minimap: { enabled: false },
          scrollBeyondLastLine: false,
          fontSize: 14,
          fontFamily: "var(--font-mono)",
          lineNumbers: "on",
          folding: true,
          automaticLayout: true,
          tabSize: 2,
          wordWrap: "on",
          padding: { top: 16 }
        }}
      />
    </div>
  );
};

export default CodeEditor;
