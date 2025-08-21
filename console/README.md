# console application for vtrgo-data-collection
 
The web-based dashboard for visualizing the industrial data collected by [service](../service/README.md). It provides real-time and historical insights into machine status, faults, and performance metrics through an intuitive, newspaper-themed interface.

This application is built with React, TypeScript, and Vite, and it communicates with the backend service to fetch and display data.

## Features

*   **Comprehensive Dashboard**: A single-page application that presents key performance indicators in a clean, grid-based layout.
*   **Rich Data Visualization**:
    *   **Status Cards**: Display uptime and operational state percentages for different system components.
    *   **Fault Counters**: Show aggregated fault counts over selected time ranges.
    *   **Float Averages**: Present the mean values of analog signals like motor speed and temperature.
    *   **Interactive Time-Series Charts**: Utilizes **Recharts** to plot historical trends for any available float (analog) value.
*   **Dynamic Theming**: Includes several built-in UI themes that can be changed on the fly from the configuration drawer.
*   **User-Friendly Interactivity**:
    *   **Time Range Selector**: Easily switch between different historical views (e.g., Last Hour, Last 24 Hours, Last 7 Days).
    *   **Chart Field Selector**: Dynamically choose which analog signal to plot in the main chart.
*   **Backend Configuration**:
    *   **Configuration Drawer**: A slide-out panel for managing UI themes and backend settings.
    *   **Dynamic Mapping Upload**: Provides a file upload interface to send a new Tag/Register Mapping CSV to the backend service.

## Tech Stack

*   **Framework**: React 18
*   **Language**: TypeScript
*   **Build Tool**: Vite
*   **Styling**: Tailwind CSS
*   **UI Components**: shadcn/ui
*   **Charting**: Recharts
*   **Icons**: Lucide React
*   **Linting**: ESLint + TypeScript ESLint

## Project Structure

The `console` directory contains all the frontend source code.

```
console/
├── dist/           # The compiled, production-ready output
├── src/            # Source code
│   ├── assets/     # Static assets like fonts and images
│   ├── components/ # React components
│   │   ├── ui/     # Auto-generated shadcn/ui components
│   │   └── *.tsx   # Application-specific components (Cards, Charts, Drawer)
│   ├── hooks/      # Custom React hooks for data fetching
│   ├── lib/        # Utility functions (theming, time formatting)
│   └── main.tsx    # Main application entry point
├── index.html      # HTML entry point for Vite
├── package.json    # Project dependencies and scripts
└── vite.config.ts  # Vite build configuration
```

## How It Works

The console is a client-side application that is served statically by the backend `service`.

1.  **Data Fetching**: The dashboard uses custom React hooks (`useStats`, `useFloatRange`) to make periodic calls to the backend's REST API endpoints:
    *   `GET /api/stats`: Fetches aggregated data (percentages, counts, averages) for the statistics cards.
    *   `GET /api/float-range`: Fetches detailed time-series data for the line chart.
2.  **State Management**: Component state (like the selected time range and chart field) is managed locally within the React components. Fetched data is stored in state and passed down to child components as props.
3.  **Rendering**:
    *   The main `Dashboard` component arranges responsive `Card` components in a grid.
    *   `StatsCard`, `FaultsCard`, and `FloatsCard` components iterate over the data from `/api/stats` to render the metrics.
    *   The `FloatChart` component uses the data from `/api/float-range` to render an interactive line chart with `recharts`.
4.  **Configuration**: The `ConfigDrawer` component provides a UI to interact with application settings. When a user uploads a CSV file, it is sent via a `POST` request to the `/api/upload-csv` endpoint on the service.

## Prerequisites

*   Node.js (LTS version recommended)
*   `pnpm` (or `npm`/`yarn`) package manager
*   A running instance of the **VTArchitect Service**.

## Available Scripts

In the `console` directory, you can run the following commands:

#### Install dependencies

```bash
pnpm install
```

#### Start the development server

This will run the frontend on `http://localhost:5173` with Hot-Module Replacement (HMR). This is useful for UI development, but note that API calls will fail unless the backend service is running and configured to handle CORS for this origin.

```bash
pnpm run dev
```

#### Lint the code

Runs ESLint to check for code quality and style issues.

```bash
pnpm run lint
```

#### Build for production

This command transpiles the TypeScript/React code and bundles it into a static `dist/` directory. The VTArchitect Service is configured to serve these files.

```bash
pnpm run build
```

#### Preview the production build

This serves the contents of the `dist/` directory locally. It's a good way to test the production build before deployment.

```bash
pnpm run preview
```

## Key Components

*   **`Dashboard.tsx`**: The main component that orchestrates the layout and data fetching for the entire page.
*   **`StatsCard.tsx`**: A reusable card to display a collection of named boolean percentage values.
*   **`FaultsCard.tsx`**: A card specifically for displaying fault counts, with custom formatting.
*   **`FloatsCard.tsx`**: A card for displaying named float average values.
*   **`FloatChart.tsx`**: Renders a `recharts` line chart based on a selected float field and time range.
*   **`ConfigDrawer.tsx`**: A slide-out drawer using `vaul` that contains UI for changing themes and uploading a new backend configuration CSV.
*   **`VTRTitle.tsx`**: The main stylized title component, which changes its appearance based on the selected theme.

## Adding shadcn/ui Components

This project uses `shadcn/ui` for its base components. To add new ones, you first need to initialize the project:

```bash
pnpm dlx shadcn@latest init
```

Then, you can add new components:

```bash
pnpm dlx shadcn@latest add [component-name]
```

Refer to the shadcn/ui documentation for a list of available components.