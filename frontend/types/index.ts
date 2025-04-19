export interface Deployment {
  id: string;
  name: string;
  status: 'Running' | 'Stopped' | 'Failed' | 'Building' | 'Creating' | 'Starting';
  createdAt: string;
  language: 'node' | 'go' | 'python';
  port?: string;
  built: boolean;
}

export interface DeploymentFiles {
  packageFile: string;
  codeFile: string;
}

export interface DeploymentFileMap {
  [id: string]: DeploymentFiles;
}