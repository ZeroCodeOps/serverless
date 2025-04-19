# Serverless Platform v3

This project is a simplified serverless platform, allowing users to deploy and manage serverless functions written in Node.js, Go, and Python. It provides a web-based dashboard to create, edit, build, deploy, and monitor your serverless functions.

This is version 3 of the project, building upon previous iterations with a focus on improved user experience and code clarity.

## Features

- **Deployment Creation:** Easily create new serverless deployments with a chosen name and programming language (Node.js, Go, Python) through the dashboard.
- **Code Editing:** Integrated code editor directly in the browser to modify the source code and dependency files of your deployments.
- **Build & Deploy:** Trigger builds for your deployments using the Knative `func` CLI.
- **Start & Stop Functions:** Control the lifecycle of your deployed functions with start and stop actions.
- **Real-time Status Updates:** WebSocket-based real-time monitoring of deployment status changes.
- **Deployment Management:** View and manage all your deployments through a centralized dashboard.
- **File Upload:** Upload code and package files for your deployments.
- **Port Management:** Automatic port detection and management for running functions.

## Tech Stack

- **Frontend:**
  - [Next.js](https://nextjs.org/) - React framework for building user interfaces.
  - [TypeScript](https://www.typescriptlang.org/) - For type safety and improved code quality.
  - [Tailwind CSS](https://tailwindcss.com/) - Utility-first CSS framework for styling.
  - [shadcn/ui](https://ui.shadcn.com/) - Reusable UI components styled with Tailwind CSS.
  - [Monaco Editor](https://microsoft.github.io/monaco-editor/) - Code editor component for in-browser code editing.
  - [WebSocket](https://developer.mozilla.org/en-US/docs/Web/API/WebSockets_API) - For real-time status updates.

- **Backend:**
  - [Go](https://go.dev/) - Programming language for backend API and function management.
  - [Gorilla WebSocket](https://github.com/gorilla/websocket) - WebSocket implementation for real-time communication.
  - [Knative func CLI](https://knative.dev/docs/functions/) - For function creation, building, and running.
  - SQLite - For persistent storage of deployment metadata.
  - Standard Go libraries for HTTP handling, JSON processing, and file system operations.

## Getting Started

To run this project locally, you'll need to have the following installed:

- **Go:** [Installation Guide](https://go.dev/doc/install) (for the backend)
- **Node.js & npm (or pnpm/yarn/bun):** [Installation Guide](https://nodejs.org/) (for the frontend)
- **`func` CLI:** [Installation Guide](https://knative.dev/docs/functions/install-func-cli/) - Required for function management.

**Steps to run the project:**

1. **Clone the repository:**
    ```bash
    git clone <repository_url>
    cd serverless
    ```

2. **Start the Backend:**
    ```bash
    cd backend
    go mod download  # Install dependencies
    go run main.go
    ```
    The backend server will start on `http://localhost:8080`.

3. **Start the Frontend:**
    ```bash
    cd ../frontend
    npm install  # or yarn install / pnpm install / bun install
    npm run dev    # or yarn dev / pnpm dev / bun dev
    ```
    The frontend development server will start on `http://localhost:3000`.

4. **Access the Dashboard:**
    Open your browser and navigate to [http://localhost:3000](http://localhost:3000).

## Project Structure

```
serverless/
├── backend/
│   ├── handlers/     # HTTP and WebSocket handlers
│   ├── db/          # Database operations
│   ├── types/       # Type definitions
│   └── main.go      # Entry point
├── frontend/
│   ├── src/         # Source code
│   ├── public/      # Static assets
│   └── package.json # Dependencies
└── README.md
```

## Important Notes

- **Function Execution:** Functions are executed locally using the `func` CLI. Each function runs in its own process and is assigned a unique port.
- **Data Persistence:** Deployment metadata is stored in SQLite, while function code and dependencies are stored in the local filesystem.
- **Port Management:** The platform automatically detects and manages ports for running functions. Ports are released when functions are stopped.
- **Real-time Updates:** The dashboard uses WebSocket connections to receive real-time status updates for deployments.
- **File Management:** Code and package files are managed through the dashboard's file upload functionality.
- **Error Handling:** Comprehensive error handling is implemented for both frontend and backend operations.

## API Endpoints

The backend provides the following REST endpoints:

- `POST /create/{language}` - Create a new deployment
- `POST /upload/{name}` - Upload code and package files
- `POST /build/{name}` - Build a deployment
- `POST /start/{name}` - Start a deployment
- `POST /stop/{name}` - Stop a deployment
- `GET /deployments` - List all deployments
- `GET /deployments/{name}` - Get deployment details
- `GET /ws` - WebSocket connection for real-time updates

## Learn More

- **Next.js Documentation:** [https://nextjs.org/docs](https://nextjs.org/docs)
- **Go Programming Language:** [https://go.dev/](https://go.dev/)
- **Knative Functions (func CLI):** [https://knative.dev/docs/functions/](https://knative.dev/docs/functions/)
- **Gorilla WebSocket:** [https://github.com/gorilla/websocket](https://github.com/gorilla/websocket)

## Contributing

Contributions are welcome! If you have ideas for improvements, bug fixes, or new features, please feel free to open issues and pull requests.

## License

[MIT License](LICENSE)
