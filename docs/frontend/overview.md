## 🎨 Frontend – Technical Docs

### 📅 Overview

A modern UI built with **React + Vite** for managing and visualizing telemetry collectors. It supports configuration of pipelines, agent lifecycle operations, and real-time metrics display.

---

### 🔧 Setup for Development

#### 1. Prerequisites

- Node.js (>=18)
- npm

#### 2. Fork & Install

```bash
git clone https://github.com/your-username/ctrlb-control-plane.git
cd ctrlb-control-plane/frontend
npm install
```

#### 3. Required Environment Variable

Create a `.env` file:

```env
VITE_API_URL=http://localhost:8096
```

> ✅ This should point to your backend server URL.

#### 4. Run Locally

```bash
npm run dev
```

App will be available at `http://localhost:5173` by default.

---

### 🔹 Features

- Graph-based pipeline configuration
- Start/stop agents
- View agent metrics (CPU, memory, telemetry rates)
- Real-time status of agents and pipelines

---

### 🚧 Project Structure (Simplified)

```
frontend/
├── src/
│   ├── App.tsx             # Main entry point
│   ├── components/         # UI components
│   ├── constants.ts        # App-wide constants
│   ├── context/            # React context providers
│   ├── hooks/              # Custom hooks
│   ├── index.css           # Global styles
│   ├── main.tsx            # Entry file for ReactDOM
│   ├── services/           # API interaction services
│   ├── types/              # Shared type definitions
│   └── utils/              # Reusable utilities
```

---

### 🔍 Debugging Tips

- Ensure `VITE_API_URL` is correct and backend is reachable
- Network tab can help debug REST API issues
- Use React DevTools for inspecting context state

---

### ✨ Production Build

```bash
npm run build
```

Build output will be in the `dist/` folder.

---

### ❌ Troubleshooting

| Issue                        | Possible Fix                       |
| ---------------------------- | ---------------------------------- |
| API calls failing            | Check `VITE_API_URL`, backend port |
| Blank page / UI doesn't load | Ensure Vite dev server is running  |
|                              |                                    |

---

> For deployment, bundle `dist/` behind a web server like Nginx or serve statically via cloud hosting.

Let us know via GitHub Issues if you run into any bugs or UI quirks.
