import { useRouter } from 'next/navigation';
import { Deployment } from '../types';

interface DeploymentTableProps {
  deployments: Deployment[];
  onDelete: (id: string) => void;
}

const DeploymentTable: React.FC<DeploymentTableProps> = ({ deployments, onDelete }) => {
  const router = useRouter();

  const handleEdit = (id: string): void => {
    router.push(`/edit/${id}`);
  };

  const statusBadge = (status: string) => {
    switch(status) {
      case 'Running':
        return <span className="badge badge-success">Running</span>;
      case 'Failed':
        return <span className="badge badge-error">Failed</span>;
      default:
        return <span className="badge badge-neutral">Stopped</span>;
    }
  };

  return (
    <div className="w-full overflow-auto rounded-lg border border-border">
      <table className="table-base">
        <thead className="table-header">
          <tr>
            <th>ID</th>
            <th>Name</th>
            <th>Status</th>
            <th>Created</th>
            <th className="text-right">Actions</th>
          </tr>
        </thead>
        <tbody className="table-body">
          {deployments.length === 0 ? (
            <tr>
              <td colSpan={5} className="text-center py-8 text-muted-foreground">
                No deployments found
              </td>
            </tr>
          ) : (
            deployments.map((deployment) => (
              <tr key={deployment.id}>
                <td>{deployment.id}</td>
                <td className="font-medium">{deployment.name}</td>
                <td>{statusBadge(deployment.status)}</td>
                <td>{deployment.createdAt}</td>
                <td className="text-right">
                  <div className="flex justify-end gap-2">
                    <button
                      onClick={() => handleEdit(deployment.id)}
                      className="btn btn-sm btn-outline"
                    >
                      Edit
                    </button>
                    <button
                      onClick={() => onDelete(deployment.id)}
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
  );
};

export default DeploymentTable;