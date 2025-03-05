import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Deployment } from '../types';

interface DeploymentTableProps {
  deployments: Deployment[];
  onDelete: (id: string) => void;
  onToggle: (id: string) => void;
}

const DeploymentTable: React.FC<DeploymentTableProps> = ({ 
  deployments, 
  onDelete,
  onToggle
}) => {
  const router = useRouter();
  const [confirmDelete, setConfirmDelete] = useState<string | null>(null);

  const handleEdit = (name: string): void => {
    router.push(`/edit/${name}`);
  };
  
  const handleToggle = (id: string, e: React.MouseEvent): void => {
    e.stopPropagation();
    onToggle(id);
  };
  
  const handleDeleteClick = (id: string, e: React.MouseEvent): void => {
    e.stopPropagation();
    setConfirmDelete(id);
  };
  
  const handleConfirmDelete = (): void => {
    if (confirmDelete) {
      onDelete(confirmDelete);
      setConfirmDelete(null);
    }
  };

  const statusBadge = (status: string) => {
      switch(status) {
        case 'Running':
          return <span className="badge badge-success">Running</span>;
        case 'Failed':
          return <span className="badge badge-error">Failed</span>;
        case 'Building':
          return (
            <span className="badge bg-amber-100 text-amber-700 border-amber-200 dark:bg-amber-900/30 dark:text-amber-400 dark:border-amber-800 flex items-center gap-1 w-min">
              <svg className="animate-spin h-3 w-3 mr-1" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              Building
            </span>
          );
        default:
          return <span className="badge badge-neutral">Stopped</span>;
      }
    };

  const languageIcon = (language: string) => {
    switch(language) {
      case 'node':
        return (
          <span className="flex items-center gap-1 text-green-600 dark:text-green-400">
            <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 21.985c-.275 0-.532-.074-.772-.202l-2.439-1.448c-.365-.203-.182-.277-.072-.314.496-.165.588-.201 1.101-.493.056-.037.129-.01.185.019l1.87 1.12c.074.036.166.036.221 0l7.319-4.237c.074-.036.11-.11.11-.202V7.768c0-.091-.036-.165-.11-.201l-7.319-4.219c-.073-.037-.166-.037-.221 0L4.552 7.566c-.073.036-.11.129-.11.201v8.457c0 .073.037.165.11.201l2 1.157c1.082.548 1.762-.095 1.762-.735V8.502c0-.11.091-.221.203-.221h.936c.11 0 .22.092.22.221v8.348c0 1.449-.788 2.294-2.164 2.294-.422 0-.752 0-1.688-.46l-1.925-1.099a1.55 1.55 0 0 1-.771-1.34V7.786c0-.55.293-1.064.771-1.339l7.316-4.237a1.637 1.637 0 0 1 1.544 0l7.317 4.237c.479.274.771.789.771 1.339v8.458c0 .549-.293 1.063-.771 1.34l-7.317 4.236c-.241.11-.53.184-.806.184zm2.294-5.771c-3.21 0-3.87-1.468-3.87-2.714 0-.11.092-.221.22-.221h.954c.11 0 .202.074.202.184.147.971.568 1.449 2.514 1.449 1.54 0 2.202-.35 2.202-1.175 0-.477-.184-.825-2.587-1.063-1.999-.2-3.246-.643-3.246-2.238 0-1.485 1.247-2.366 3.339-2.366 2.348 0 3.503.809 3.649 2.568a.213.213 0 0 1-.056.166c-.037.036-.092.073-.147.073h-.953a.211.211 0 0 1-.202-.164c-.221-1.012-.789-1.34-2.292-1.34-1.689 0-1.891.587-1.891 1.027 0 .531.237.696 2.514.99 2.256.293 3.32.715 3.32 2.294-.02 1.615-1.339 2.531-3.67 2.531z" />
            </svg>
            Node.js
          </span>
        );
      case 'go':
        return (
          <span className="flex items-center gap-1 text-blue-600 dark:text-blue-400">
            <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" viewBox="0 0 24 24" fill="currentColor">
              <path d="M1.811 10.231c-.047 0-.058-.023-.035-.059l.246-.315c.023-.035.081-.058.128-.058h4.172c.046 0 .058.035.035.07l-.199.303c-.023.036-.082.07-.117.07zM.047 11.306c-.047 0-.059-.023-.035-.058l.245-.316c.023-.035.082-.058.129-.058h5.328c.047 0 .07.035.058.07l-.093.28c-.012.047-.058.07-.105.07zm2.828 1.075c-.047 0-.059-.035-.035-.07l.163-.292c.023-.035.07-.07.117-.07h2.337c.047 0 .07.035.07.082l-.023.28c0 .047-.047.082-.082.082zm12.129-2.36c-.736.187-1.239.327-1.963.514-.176.046-.187.058-.34-.117-.176-.199-.304-.327-.548-.444-.737-.362-1.45-.257-2.115.175-.795.514-1.204 1.274-1.192 2.22.011.935.654 1.706 1.577 1.835.795.105 1.46-.175 1.987-.77.105-.13.198-.27.315-.434H10.47c-.245 0-.304-.152-.222-.35.152-.362.432-.97.596-1.274a.315.315 0 0 1 .292-.187h4.253c-.023.316-.023.631-.07.947a4.983 4.983 0 0 1-.958 2.29c-.841 1.11-1.94 1.8-3.33 1.986-1.145.152-2.209-.07-3.143-.77-.902-.655-1.577-1.54-1.704-2.659-.152-1.274.222-2.431.993-3.444.83-1.087 1.928-1.776 3.272-2.02 1.098-.2 2.15-.07 3.096.571.62.41 1.063.97 1.356 1.648.07.105.023.164-.117.199M22.8 10.222c-.012-.82-.563-1.518-1.31-1.716-.868-.233-1.98.152-2.337 1.4-.304 1.074-.047 2.359.83 2.951.795.548 1.705.315 2.337-.444.517-.62.766-1.4.822-2.191h-1.658c-.035 0-.07-.023-.07-.07v-.187c0-.058.035-.07.07-.07h2.243c.035 0 .07.023.07.058 0 .778-.07 1.557-.34 2.29a3.178 3.178 0 0 1-1.436 1.799c-.634.315-1.308.41-2.008.27-.398-.082-.788-.233-1.133-.468-.82-.538-1.251-1.274-1.367-2.265a3.6 3.6 0 0 1 .526-2.177c.73-1.144 1.893-1.695 3.226-1.436.63.117 1.156.445 1.566.97.070.082.082.152-.023.187-.128.094-.257.176-.386.258-.035.023-.07.011-.117-.023a1.957 1.957 0 0 0-1.492-.538zm0 0" />
            </svg>
            Go
          </span>
        );
      case 'python':
        return (
          <span className="flex items-center gap-1 text-yellow-600 dark:text-yellow-400">
            <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" viewBox="0 0 24 24" fill="currentColor">
              <path d="M14.31.18l.9.2.73.26.59.3.45.32.34.34.25.34.16.33.1.3.04.26.02.2-.01.13V8.5l-.05.63-.13.55-.21.46-.26.38-.3.31-.33.25-.35.19-.35.14-.33.1-.3.07-.26.04-.21.02H8.83l-.69.05-.59.14-.5.22-.41.27-.33.32-.27.35-.2.36-.15.37-.1.35-.07.32-.04.27-.02.21v3.06H3.23l-.21-.03-.28-.07-.32-.12-.35-.18-.36-.26-.36-.36-.35-.46-.32-.59-.28-.73-.21-.88-.14-1.05L0 11.97l.06-1.22.16-1.04.24-.87.32-.71.36-.57.4-.44.42-.33.42-.24.4-.16.36-.1.32-.05.24-.01h.16l.06.01h8.16v-.83H6.24l-.01-2.75-.02-.37.05-.34.11-.31.17-.28.25-.26.31-.23.38-.2.44-.18.51-.15.58-.12.64-.1.71-.06.77-.04.84-.02 1.27.05 1.07.13zm-6.3 1.98l-.23.33-.08.41.08.41.23.34.33.22.41.09.41-.09.33-.22.23-.34.08-.41-.08-.41-.23-.33-.33-.22-.41-.09-.41.09zm13.09 3.95l.28.06.32.12.35.18.36.27.36.35.35.47.32.59.28.73.21.88.14 1.04.05 1.23-.06 1.23-.16 1.04-.24.86-.32.71-.36.57-.4.45-.42.33-.42.24-.4.16-.36.09-.32.05-.24.02-.16-.01h-8.22v.82h5.84l.01 2.76.02.36-.05.34-.11.31-.17.29-.25.25-.31.24-.38.2-.44.17-.51.15-.58.13-.64.09-.71.07-.77.04-.84.01-1.27-.04-1.07-.14-.9-.2-.73-.25-.59-.3-.45-.33-.34-.34-.25-.34-.16-.33-.1-.3-.04-.25-.02-.2.01-.13v-5.34l.05-.64.13-.54.21-.46.26-.38.3-.32.33-.24.35-.2.35-.14.33-.1.3-.06.26-.04.21-.02.13-.01h5.84l.69-.05.59-.14.5-.21.41-.28.33-.32.27-.35.2-.36.15-.36.1-.35.07-.32.04-.28.02-.21V6.07h2.09l.14.01zm-6.47 14.25l-.23.33-.08.41.08.41.23.33.33.23.41.08.41-.08.33-.23.23-.33.08-.41-.08-.41-.23-.33-.33-.23-.41-.08-.41.08z" />
            </svg>
            Python
          </span>
        );
      default:
        return <span>Unknown</span>;
    }
  };

  return (
    <>
      <div className="w-full overflow-auto rounded-lg border border-border">
        <table className="table-base">
          <thead className="table-header">
            <tr>
              <th>ID</th>
              <th>Name</th>
              <th>Language</th>
              <th>Status</th>
              <th>Created</th>
              <th className="text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="table-body">
            {!deployments|| deployments.length === 0 ? (
              <tr>
                <td colSpan={6} className="text-center py-8 text-muted-foreground">
                  No deployments found
                </td>
              </tr>
            ) : (
              deployments.map((deployment) => (
                <tr 
                  key={deployment.name}
                  onClick={() => handleEdit(deployment.name)}
                  className="cursor-pointer"
                >
                  <td>{deployment.id}</td>
                  <td className="font-medium">{deployment.name}</td>
                  <td>{languageIcon(deployment.language)}</td>
                  <td>{statusBadge(deployment.status)}</td>
                  <td>{deployment.createdAt}</td>
                  <td className="text-right">
                    <div className="flex justify-end gap-2">
                      <button
                        onClick={(e) => handleToggle(deployment.id, e)}
                        className={`btn btn-sm ${
                          deployment.status === 'Running' ? 'btn-secondary' : 'btn-primary'
                        }`}
                      >
                        {deployment.status === 'Running' ? 'Stop' : 'Start'}
                      </button>
                      <button
                        onClick={() => handleEdit(deployment.name)}
                        className="btn btn-sm btn-outline"
                      >
                        Edit
                      </button>
                      <button
                        onClick={(e) => handleDeleteClick(deployment.name, e)}
                        className="btn btn-sm btn-destructive"
                      >
                        Delete
                      </button>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
      
      {confirmDelete && (
        <div className="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4">
          <div className="bg-card border border-border rounded-lg shadow-lg w-full max-w-md p-6">
            <h3 className="text-lg font-semibold mb-2">Confirm Deletion</h3>
            <p className="text-muted-foreground mb-4">
              Are you sure you want to delete this deployment? This action cannot be undone.
            </p>
            <div className="flex justify-end gap-2">
              <button
                onClick={() => setConfirmDelete(null)}
                className="btn btn-outline"
              >
                Cancel
              </button>
              <button
                onClick={handleConfirmDelete}
                className="btn btn-destructive"
              >
                Delete
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
};

export default DeploymentTable;