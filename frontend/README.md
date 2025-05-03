# Frontend Application Overview

---

## Table of Contents

1. [Folder Structure](#folder-structure)
2. [Key Features](#key-features)
3. [How to Run](#how-to-run)

---

## Folder Structure

### `src/`

Contains the main source code for the application.

- **`components/`**: Reusable UI components used throughout the application.

  - **`Agents/`**: Components for managing and displaying agents, including:
    - `AgentsTable`: Displays a list of agents.
    - `Charts`: Visualizes CPU/Memory usage and metrics for agents.
    - `AgentLandingPage`: The main landing page for agents.
  - **`CanvasForPipelines/`**: Components for visualizing pipelines, including:
    - Source, Processor, and Destination nodes.
    - Drag-and-drop functionality for pipeline design.
  - **`Pipelines/`**: Components for managing pipelines, including:
    - `PipelineOverviewTable`: Displays a list of pipelines and their details.
    - `PipelineDetails`: Shows detailed information about a specific pipeline.
  - **`Pipelines/AddNewPipelineComponent/`**: Components for creating pipelines, including:
    - Adding pipeline information.
    - Configuring source and destination nodes.
    - Assigning agents to pipelines.
  - **`Pipelines/DropdownOptions/`**: Contains dropdown options for selecting types (e.g., source, destination, processor).

- **`services/`**: Contains API service functions for interacting with the backend.

  - `agentServices.ts`: Handles API calls related to agents (e.g., fetching, creating, deleting agents).
  - `pipelineServices.ts`: Handles API calls related to pipelines (e.g., fetching, creating, updating pipelines).

- **`constants/`**: Holds constant values used across the application to maintain consistency.

  - Example: API endpoints, route paths, and reusable strings.

- **`types/`**: TypeScript type definitions for ensuring type safety in the application.

  - Example: `Agent.types.ts`, `Pipeline.types.ts`.

- **`context/`**: Custom React hooks encapsulating reusable logic.

  - Example: `useAgentsValues.tsx`, `usePipelineOverview.tsx`.

- **`lib/`**: Utility functions and libraries that support the application.

  - Example: Axios instance configuration, helper functions.

- **`pages/`**: Contains standalone pages for the application.
  - `Login.tsx`: Login page for user authentication.
  - `Signup.tsx`: Signup page for user registration.

---

### `public/`

Contains static assets such as images, icons, and other files that are served directly.

---

## Key Features

1. **Agent Management**:

   - View, create, update, and delete agents.
   - Visualize agent metrics (CPU/Memory usage, health metrics).

2. **Pipeline Management**:

   - Create and configure pipelines with source, processor, and destination nodes.
   - Visualize pipelines using a drag-and-drop interface.
   - Assign agents to pipelines and manage their configurations.

3. **Authentication**:

   - Login and signup functionality for users.
   - Token-based authentication with automatic token refresh.

4. **Error Handling**:
   - Graceful error handling for API calls and user interactions.

---

## How to Run

To set up and run the frontend application, follow these steps:

### Prerequisites

- Ensure you have Node.js (v16 or later) and npm installed on your system.

### Steps

1. Clone the repository:

   ```bash
   git clone https://github.com/your-repo/frontend.git
   cd frontend

   ```

2. Install the dependencies

   ```bash
   npm install --legacy-peer-deps

   ```

3. Start the development server:

   ```bash
   npm run dev

   ```

4. Open your browser and navigate to:
   ```bash
   http://localhost:3030
   ```
