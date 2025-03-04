import { Deployment, DeploymentFileMap } from '../types';

export const mockDeployments: Deployment[] = [
  { id: '1', name: 'Deployment 1', status: 'Running', createdAt: '2023-08-01' },
  { id: '2', name: 'Deployment 2', status: 'Stopped', createdAt: '2023-08-02' },
  { id: '3', name: 'Deployment 3', status: 'Failed', createdAt: '2023-08-03' },
];

export const mockFiles: DeploymentFileMap = {
  '1': {
    packageFile: '{\n  "name": "deployment-1",\n  "version": "1.0.0",\n  "dependencies": {\n    "express": "^4.17.1"\n  }\n}',
    codeFile: 'const express = require("express");\nconst app = express();\n\napp.get("/", (req, res) => {\n  res.send("Hello World!");\n});\n\napp.listen(3000, () => {\n  console.log("Server running on port 3000");\n});'
  },
  '2': {
    packageFile: '{\n  "name": "deployment-2",\n  "version": "1.0.0",\n  "dependencies": {\n    "lodash": "^4.17.21"\n  }\n}',
    codeFile: 'const _ = require("lodash");\n\nconst array = [1, 2, 3, 4, 5];\nconsole.log(_.sum(array));'
  },
  '3': {
    packageFile: '{\n  "name": "deployment-3",\n  "version": "1.0.0",\n  "dependencies": {\n    "axios": "^0.21.1"\n  }\n}',
    codeFile: 'const axios = require("axios");\n\nasync function getData() {\n  try {\n    const response = await axios.get("https://jsonplaceholder.typicode.com/todos/1");\n    console.log(response.data);\n  } catch (error) {\n    console.error(error);\n  }\n}\n\ngetData();'
  }
};