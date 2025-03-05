import { useState } from 'react';

interface NewDeploymentDialogProps {
  isOpen: boolean;
  onConfirm: (name: string, language: 'node' | 'go' | 'python') => void;
  onCancel: () => void;
}

export const NewDeploymentDialog: React.FC<NewDeploymentDialogProps> = ({
  isOpen,
  onConfirm,
  onCancel
}) => {
  const [name, setName] = useState<string>('');
  const [language, setLanguage] = useState<'node' | 'go' | 'python'>('node');
  const [error, setError] = useState<string>('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    // Validate deployment name
    if (!name.trim()) {
      setError('Deployment name is required');
      return;
    }
    
    if (name.length < 3) {
      setError('Deployment name must be at least 3 characters');
      return;
    }
    
    if (!/^[a-zA-Z0-9-_]+$/.test(name)) {
      setError('Deployment name can only contain letters, numbers, hyphens, and underscores');
      return;
    }
    
    onConfirm(name, language);
    setName(''); // Reset form after submission
    setLanguage('node'); // Reset to default
    setError(''); // Clear any errors
  };
  
  if (!isOpen) return null;
  
  return (
    <div className="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4">
      <div className="bg-card border border-border rounded-lg shadow-lg w-full max-w-md animate-in fade-in duration-300">
        <div className="p-6">
          <h2 className="text-xl font-semibold mb-4">New Deployment</h2>
          
          <form onSubmit={handleSubmit}>
            <div className="mb-4">
              <label htmlFor="deployment-name" className="block text-sm font-medium mb-1">
                Deployment Name
              </label>
              <input
                id="deployment-name"
                type="text"
                className={`input w-full ${error ? 'border-destructive' : ''}`}
                value={name}
                onChange={(e) => {
                  setName(e.target.value);
                  if (error) setError('');
                }}
                placeholder="my-serverless-function"
                autoFocus
              />
              {error && (
                <p className="mt-1 text-sm text-destructive">{error}</p>
              )}
              <p className="mt-1 text-xs text-muted-foreground">
                Choose a unique name for your deployment. This will help you identify it later.
              </p>
            </div>
            
            <div className="mb-4">
              <label htmlFor="language-select" className="block text-sm font-medium mb-1">
                Language
              </label>
              <select
                id="language-select"
                className="input w-full"
                value={language}
                onChange={(e) => setLanguage(e.target.value as 'node' | 'go' | 'python')}
              >
                <option value="nodejs">Node.js</option>
                <option value="go">Go</option>
                <option value="python">Python</option>
              </select>
              <p className="mt-1 text-xs text-muted-foreground">
                Select the programming language for your serverless function.
              </p>
            </div>
            
            <div className="flex justify-end gap-2 mt-6">
              <button 
                type="button"
                onClick={onCancel}
                className="btn btn-outline"
              >
                Cancel
              </button>
              <button 
                type="submit"
                className="btn btn-primary"
              >
                Create
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
};