# TODO

## High Priority - Must Do
- [x] Cache Environment Variable Values (load all config at startup and use *config.Config throughout the codebase)
- [x] Implement Batch Writes to InfluxDB (buffer data points and write in batches to reduce write frequency and improve performance)
- [x] Optimize PLC Polling Interval (adjust polling interval for best balance between data freshness and system load)

## Medium Priority - Should Do
- [ ] Profile the Application (use Go profiling tools to identify and address CPU/memory bottlenecks)
- [ ] Review and Optimize Flux Queries (analyze and improve InfluxDB query performance)

## Low Priority - Could Do (If Time Permits)
- [ ] Optimize Data Structures (refactor data structures for memory and performance efficiency)
- [ ] Implement Caching for API Endpoint (cache API responses to reduce database load and improve response times)
- [ ] Implement Gzip Compression for API Responses (enable gzip to reduce response size and bandwidth usage)

## Local vs Remote Content Routing
- [ ] Implement split-horizon DNS: local `vtrdata.com` resolves to LAN IP; public `vtrdata.com` resolves to public IP
- [ ] Use Nginx or Go to detect local vs external client (via IP range)
- [ ] Serve Vite + React app to LAN users, fallback HTML to remote internet users
- [ ] Build and deploy Vite static files to correct directory for LAN delivery
  
## React Native Mobile App
### High Priority
- [ ] Set up React Native project structure (navigation, state management, theming)
- [ ] Implement API integration (connect to backend endpoints for data retrieval and actions)
- [ ] Design and build core mobile UI screens (dashboard, alerts, settings)

### Medium Priority
- [ ] Add offline support and data caching (ensure app works with intermittent connectivity)
- [ ] Implement authentication and user management (secure access to app features)
- [ ] Add push notifications (notify users of important events or data changes)

### Low Priority
- [ ] Polish mobile UI/UX (animations, accessibility)
- [ ] Add advanced mobile data visualization (charts, widgets)
- [ ] Write end-to-end and integration tests for mobile app

## Vite + React Dashboard (Local Web)
### High Priority
- [x] Build LAN-accessible dashboard UI with core status panels and visualization
- [x] Connect to backend API endpoints for data polling and analytics
- [ ] Integrate modular components for system stats, boolean breakdowns, and fault logging

### Medium Priority
- [ ] Optimize component performance and reduce re-renders
- [ ] Add support for responsive layout (desktop/tablet view)
- [ ] Enable browser-based data export (CSV, JSON)

### Low Priority
- [ ] Polish visual layout and design aesthetics
- [ ] Add user-adjustable settings (interval, theme, measurement source)
- [ ] Add context-aware tooltips and glossary for field names

## FUTURE: Multi-PLC Support

- [ ] Centralized PLC Configuration
- [ ] Concurrent PLC Watchers
- [ ] InfluxDB Data Tagging
- [ ] Isolated Error Handling
- [ ] Dynamic YAML Loading (or Caching)