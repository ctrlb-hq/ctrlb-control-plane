# Frontend Application Overview

---

## Table of Contents

1. [Folder Structure](#folder-structure)
2. [Key Features](#key-features)
3. [How to Run](#how-to-run)

---

## Folder Structure

### `src/`

- **`App.tsx`**: Root component that initializes the application.
- **`main.tsx`**: Entry point that renders the React application.
- **`index.css`**: Global styles for the application.
- **`constants.ts`**: Global constants used across the project.

- **`components/`**: Reusable UI components used throughout the application.

  - **`HealthChart.tsx`**: Component to display health metrics (CPU/Memory) charts.
  - **`Pipelines/`**:

    - **`AddPipelineComponents/`**:
      - `AddPipelineCanvas.tsx`: Canvas for designing new pipelines.
      - `AddPipelineDetails.tsx`: Form for entering pipeline details.
      - `AddPipelineSheet.tsx`: Sheet view for adding pipeline steps.
      - `ProgressFlow.tsx`: Visual progress indicator of pipeline creation.
    - **`DropdownOptions/`**:
      - `SourceDropdownOptions.tsx`: Dropdown options for selecting sources.
      - `ProcessorDropdownOptions.tsx`: Dropdown options for selecting processors.
      - `DestinationDropdownOptions.tsx`: Dropdown options for selecting destinations.
    - `ExistingPipelineOverview.tsx`: Displays a list of existing pipelines.
    - **`Nodes/`**:
      - `SourceNode.tsx`: Represents a source node in the pipeline.
      - `ProcessorNode.tsx`: Represents a processor node in the pipeline.
      - `DestinationNode.tsx`: Represents a destination node in the pipeline.
    - `PipelineTable.tsx`: Table view of pipelines.
    - `ViewPipelineDetails.tsx`: Detailed view of a specific pipeline.

  - **`ui/`**: Library of UI primitives.

    - `alert.tsx`, `badge.tsx`, `button.tsx`, `card.tsx`, `chart.tsx`, `checkbox.tsx`, `command.tsx`, `dialog.tsx`, `dropdown-menu.tsx`, `form.tsx`, `input.tsx`, `label.tsx`, `multi-select.tsx`, `popover.tsx`, `select.tsx`, `sheet.tsx`, `switch.tsx`, `table.tsx`, `toast.tsx`, `toaster.tsx`, `toggle.tsx`.

  - **`YAML/`**:
    - `EditConfig.tsx`: Component for editing configuration in YAML.
    - `EditPipelineYAML.tsx`: Component for editing pipeline in YAML format.

- **`context/`**: Custom React hooks for global state and context.

  - `useNodeContext.tsx`, `usePipelineChangesLog.tsx`, `usePipelineDetailContext.tsx`, `usePipelineStatus.tsx`.

- **`hooks/`**: Custom hooks.

  - `use-toast.ts`: Hook for managing toast notifications.

- **`lib/`**: Utility functions and libraries.

  - `utils.ts`: General helper functions.

- **`pages/`**: Top-level page components.

  - `HomePage.tsx`: Main landing page after login.
  - **`auth/`**:
    - `Login.tsx`: Login page for user authentication.
    - `Signup.tsx`: Signup page for user registration.

- **`services/`**: API service functions for interacting with the backend.

  - `agentServices.ts`: Handles agent-related API calls.
  - `authService.ts`: Handles authentication-related API calls.
  - `pipelineServices.ts`: Handles pipeline-related API calls.
  - `queryServices.ts`: Handles generic query-related API calls.
  - `transporterService.ts`: Handles transporter-related API calls.

- **`types/`**: TypeScript type definitions.

  - `agent.types.ts`, `agentValues.type.ts`, `auth.types.ts`, `destination.type.ts`, `node.type.ts`, `pipeline.types.ts`, `source.types.ts`, `sourceConfig.type.ts`.

- **`utils/`**: Additional utilities.
  - `axiosInstance.ts`: Axios instance configuration for API calls.

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
