# Hermod BPM

Hermod BPM (formerly GoBPM) is a professional, production-ready BPMN orchestrator built with **Go** and **React**. It provides a powerful engine for executing complex workflows, a visual designer for modeling processes, and robust management tools.

## 🚀 Key Features

- **BPMN 2.0 Engine**: Supports essential BPMN elements including:
  - **Tasks**: User Tasks, Service Tasks (HTTP/Connectors), Script Tasks (JavaScript), and Call Activities (sub-processes).
  - **Gateways**: Exclusive and Parallel Gateways.
  - **Events**: Start, End, and Intermediate Timer/Message Catch Events.
- **RabbitMQ Integration**: Production-ready messaging capabilities:
  - **Outbound Connectors**: Publish messages to RabbitMQ exchanges directly from Service Tasks.
  - **Inbound Message Correlation**: Automatically correlate RabbitMQ messages to BPMN Message Events.
  - **External Task Bridge**: Seamlessly bridge External Tasks to RabbitMQ for distributed worker patterns.
- **Connector Framework**: Plug-and-play architecture for third-party integrations (HTTP, Slack, Email, RabbitMQ).
- **Visual Designer**: Drag-and-drop BPMN modeler powered by React Flow, featuring:
  - **Edit Mode**: Load and modify existing process definitions.
  - **Property Panel**: Context-aware configuration for all BPMN nodes.
  - **Auto-Layout**: Integrated view centering for complex diagrams.
- **Asynchronous Execution**: A robust job worker system for Service Tasks and Timers with:
  - **Reliability**: Persistent job storage and execution.
  - **Error Handling**: Automatic retries with exponential backoff.
  - **Incident Management**: Capture execution failures as "Incidents" for manual resolution and retry.
- **Scripting Engine**: Integrated **Goja** (JavaScript engine) for:
  - **Script Tasks**: Complex data transformations within workflows.
  - **Dynamic Conditions**: JavaScript-based evaluation for Gateway logic.
- **Task Inbox**: A dedicated view for users to manage, claim, and complete their assigned tasks.
- **Enterprise Persistence**:
  - **Audit Logging**: Comprehensive, persistent audit trail for every state change and node transition.
  - **Security**: **AES-GCM encryption** for sensitive process variables at rest.
  - **Dual DB Support**: Supports **SQLite** for development and **PostgreSQL** for high-availability production.

## 🏗️ Architecture & Design Patterns

The project is built following **Clean Code** principles and **SOLID** design, utilizing several advanced design patterns:

- **Structural Patterns**: Facade (Service Layer), Adapter (Transports), Composite (BPMN Sub-processes), Decorator/Middleware (Logging/Auth).
- **Behavioral Patterns**: Strategy (Node Handlers), Command (Execution Steps), Observer (Event Dispatching), State (Process/Task Lifecycle), Visitor (Definition Validation), Chain of Responsibility (Condition Evaluation).
- **Creational Patterns**: Factory (Handler Creation), Builder (Test Data Setup), Singleton (DB Initialization).

## 🛠️ Technology Stack

- **Backend**: Go (1.26+), Go Kit, GORM, Connect RPC (gRPC-compatible).
- **Frontend**: React (19+), Vite, Mantine UI, React Flow, Zustand, TanStack Query, TanStack Router.
- **Integrations**: Goja (JS Runtime), Protobuf, AES-GCM Encryption.

## 📂 Project Structure

```text
├── api/              # Protocol Buffer definitions and generated code
├── cmd/gobpm/        # Main entry point (Server)
├── internal/pkg/     # Shared internal packages (Crypto, Logger)
├── migrations/       # Database migration scripts (SQL)
├── server/           # Backend Implementation
│   ├── domains/      # Core entities and business logic
│   ├── endpoints/    # Go Kit endpoint definitions
│   ├── repositories/ # Persistence layer (GORM Models & Implementations)
│   ├── services/     # Workflow engine, handlers, and business services
│   ├── transports/   # HTTP and gRPC transport layers
│   └── interceptors/ # Centralized interceptors (Logging, Auth)
└── ui/               # Frontend React Application
    ├── src/pages/    # Designer, Task Inbox, and Admin views
    ├── src/components/ # Shared UI components and BPMN Nodes
    └── src/services/  # API client generated from Protobuf
```

## 🚦 Getting Started

### Prerequisites

- **Go**: 1.26 or higher
- **Node.js**: With Bun (recommended) or npm
- **PostgreSQL**: (Optional) For production-grade persistence

### Running the Backend

1. **Clone the repository**
2. **Setup environment variables** (Optional):
   ```bash
   export DATABASE_URL="postgres://user:pass@localhost:5432/gobpm"
   ```
3. **Build the UI** (Required once before running or after UI changes):
   ```bash
   go run ./cmd/gobpm --build-ui
   ```
4. **Run the server**:
   ```bash
   go run ./cmd/gobpm
   ```
   *The server will listen on `:8080` (HTTP) and `:8081` (gRPC).*

### Running the Frontend

1. **Navigate to the UI directory**:
   ```bash
   cd ui
   ```
2. **Install dependencies**:
   ```bash
   bun install
   ```
3. **Start the development server**:
   ```bash
   bun run dev
   ```
   *Access the designer at `http://localhost:5173`.*

## 🧪 Testing

Run the backend test suite:
```bash
go test ./server/...
```

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.