import { Deployment, DeploymentFileMap } from '../types';

export const mockDeployments: Deployment[] = [
  { id: '1', name: 'Deployment 1', status: 'Running', createdAt: '2023-08-01', language: 'nodejs' },
  { id: '2', name: 'Deployment 2', status: 'Stopped', createdAt: '2023-08-02', language: 'go' },
  { id: '3', name: 'Deployment 3', status: 'Building', createdAt: '2023-08-03', language: 'python' },
];

export const mockFiles: DeploymentFileMap = {
  '1': {
    packageFile: '{\n  "name": "deployment-1",\n  "version": "1.0.0",\n  "dependencies": {\n    "express": "^4.17.1"\n  }\n}',
    codeFile: 'const express = require("express");\nconst app = express();\n\napp.get("/", (req, res) => {\n  res.send("Hello World!");\n});\n\napp.listen(3000, () => {\n  console.log("Server running on port 3000");\n});'
  },
  '2': {
    packageFile: '{\n  "name": "deployment-2",\n  "version": "1.0.0"\n}',
    codeFile: 'package main\n\nimport (\n\t"fmt"\n\t"net/http"\n)\n\nfunc handler(w http.ResponseWriter, r *http.Request) {\n\tfmt.Fprintf(w, "Hello, World!")\n}\n\nfunc main() {\n\thttp.HandleFunc("/", handler)\n\thttp.ListenAndServe(":8080", nil)\n}'
  },
  '3': {
    packageFile: '{\n  "name": "deployment-3",\n  "version": "1.0.0",\n  "dependencies": {}\n}',
    codeFile: 'from flask import Flask\n\napp = Flask(__name__)\n\n@app.route("/")\ndef hello_world():\n    return "Hello, World!"\n\nif __name__ == "__main__":\n    app.run(host="0.0.0.0", port=5000)'
  }
};