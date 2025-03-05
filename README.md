# Serverless Platform v3

This project is a simplified serverless platform, allowing users to deploy and manage serverless functions written in Node.js, Go, and Python. It provides a web-based dashboard to create, edit, build, deploy, and monitor your serverless functions.

This is version 3 of the project, building upon previous iterations with a focus on improved user experience and code clarity.

## Features

- **Deployment Creation:** Easily create new serverless deployments with a chosen name and programming language (Node.js, Go, Python) through the dashboard.
- **Code Editing:** Integrated code editor directly in the browser to modify the source code and dependency files of your deployments.
- **Build & Deploy:** Trigger builds for your deployments. _(Note: Build process is simulated in this version, actual deployment to a serverless environment is not implemented in this iteration)_
- **Start & Stop Functions:** Control the lifecycle of your deployed functions with start and stop actions.
- **Deployment Status Monitoring:** View the current status of your deployments (Running, Stopped, Building, Failed).
- **Basic Authentication:** Simple username/password login for access control (demo credentials provided for easy testing).
- **Mock Backend:** The backend logic is implemented in Go and provides basic API endpoints. Data persistence and actual serverless execution are simulated using in-memory data and command executions.

## Tech Stack

- **Frontend:**
  - [Next.js](https://nextjs.org/) (v15) - React framework for building user interfaces.
  - [TypeScript](https://www.typescriptlang.org/) - For type safety and improved code quality.
  - [Tailwind CSS](https://tailwindcss.com/) - Utility-first CSS framework for styling.
  - [shadcn/ui](https://ui.shadcn.com/) - Reusable UI components styled with Tailwind CSS.
  - [Monaco Editor](https://microsoft.github.io/monaco-editor/) - Code editor component for in-browser code editing.
- **Backend:**
  - [Go](https://go.dev/) - Programming language for backend API and function management.
  - Standard Go libraries for HTTP handling and JSON processing.

## Getting Started

To run this project locally, you'll need to have the following installed:

- **Go:** [Installation Guide](https://go.dev/doc/install) (for the backend)
- **Node.js & npm (or pnpm/yarn/bun):** [Installation Guide](https://nodejs.org/) (for the frontend)
- **`func` CLI:** [Installation Guide](https://knative.dev/docs/functions/install-func-cli/) - The Knative `func` CLI is used by the backend to simulate function creation and building. Ensure this is installed and in your PATH.

**Steps to run the project:**

1.  **Clone the repository:**

    ```bash
    git clone <repository_url>
    cd minorv3
    ```

2.  **Start the Backend:**

    ```bash
    cd backend
    go run main.go
    ```

    The backend server will start on `http://localhost:8080`.

3.  **Start the Frontend:**

    ```bash
    cd ../frontend
    npm install  # or yarn install / pnpm install / bun install
    npm run dev    # or yarn dev / pnpm dev / bun dev
    ```

    The frontend development server will start on `http://localhost:3000`.

4.  **Access the Dashboard:**
    Open your browser and navigate to [http://localhost:3000](http://localhost:3000).

5.  **Login:**
    Use the following demo credentials to log in:

    - **Username:** `admin`
    - **Password:** `admin`

    You should now be able to access the dashboard and start creating and managing deployments.

## Important Notes

- **Simulated Deployment:** This project simulates the key aspects of a serverless platform. The backend **does not** deploy functions to a real serverless environment (like AWS Lambda, Google Cloud Functions, or Azure Functions). Function execution is simulated using local `func run` commands.
- **In-Memory Data:** Deployment data and function code are stored in memory and on the local filesystem respectively. Data will be reset when the backend server is restarted.
- **Basic Authentication:** The authentication is for demonstration purposes only and is not secure for production use.
- **Error Handling:** Basic error handling is implemented, but this is not production-ready and may need further refinement.

## Learn More

- **Next.js Documentation:** [https://nextjs.org/docs](https://nextjs.org/docs)
- **Go Programming Language:** [https://go.dev/](https://go.dev/)
- **Knative Functions (func CLI):** [https://knative.dev/docs/functions/](https://knative.dev/docs/functions/)

## Contributing

Contributions are welcome! If you have ideas for improvements, bug fixes, or new features, please feel free to open issues and pull requests.

## License

[MIT License](LICENSE) (You can add a LICENSE file to the root directory and mention it here if you choose to use a license).
