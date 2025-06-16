# TODO

## High Priority - Must Do
- [x] Cache Environment Variable Values (load all config at startup and use *config.Config throughout the codebase)
- [ ] Implement Batch Writes to InfluxDB (buffer data points and write in batches to reduce write frequency and improve performance)
- [ ] Optimize PLC Polling Interval (adjust polling interval for best balance between data freshness and system load)

## Medium Priority - Should Do
- [ ] Implement Caching for API Endpoint (cache API responses to reduce database load and improve response times)
- [ ] Profile the Application (use Go profiling tools to identify and address CPU/memory bottlenecks)
- [ ] Review and Optimize Flux Queries (analyze and improve InfluxDB query performance)

## Low Priority - Could Do (If Time Permits)
- [ ] Optimize Data Structures (refactor data structures for memory and performance efficiency)
- [ ] Implement Gzip Compression for API Responses (enable gzip to reduce response size and bandwidth usage)

## React Native & React Frontend
### High Priority
- [ ] Set up React Native project structure (initialize project and configure navigation, state management, and theming)
- [ ] Implement API integration (connect to backend endpoints for data retrieval and actions)
- [ ] Design and build core UI screens (dashboard, data visualization, settings, etc.)

### Medium Priority
- [ ] Add offline support and data caching (ensure app works with intermittent connectivity)
- [ ] Implement authentication and user management (secure access to app features)
- [ ] Add push notifications (notify users of important events or data changes)

### Low Priority
- [ ] Polish UI/UX (animations, transitions, accessibility improvements)
- [ ] Add advanced data visualization (charts, graphs, and custom widgets)
- [ ] Write end-to-end and integration tests for mobile and web
