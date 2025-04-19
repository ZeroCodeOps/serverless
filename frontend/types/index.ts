export interface Deployment {
  id: string;
  name: string;
  status: 'Running' | 'Stopped' | 'Failed' | 'Building' | 'Built';
  createdAt: string;
  language: 'node' | 'go' | 'python';
  port?: string;
}

export interface DeploymentFiles {
  packageFile: string;
  codeFile: string;
}

export interface DeploymentFileMap {
  [id: string]: DeploymentFiles;
}