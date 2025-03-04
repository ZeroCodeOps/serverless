export interface Deployment {
  id: string;
  name: string;
  status: 'Running' | 'Stopped' | 'Failed';
  createdAt: string;
}

export interface DeploymentFiles {
  packageFile: string;
  codeFile: string;
}

export interface DeploymentFileMap {
  [id: string]: DeploymentFiles;
}