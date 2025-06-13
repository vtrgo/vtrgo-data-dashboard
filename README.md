# VTArchitect

VTArchitect is a Go-based application designed to interface with PLCs (Programmable Logic Controllers) and InfluxDB for data collection, processing, and visualization. It supports both Ethernet/IP and Modbus TCP protocols for PLC communication and provides an API for querying aggregated data.

## Features

- **PLC Communication**: Supports Ethernet/IP and Modbus TCP protocols.
- **InfluxDB Integration**: Writes PLC data to InfluxDB and queries aggregated data.
- **API Server**: Provides an HTTP API for querying boolean field percentages.
- **Environment Configuration**: Uses `.env` file for configuration.

## Prerequisites

- Go 1.24 or later
- InfluxDB instance
- `.env` file with the following variables:
  ```env
  INFLUXDB_URL=<your-influxdb-url>
  INFLUXDB_TOKEN=<your-influxdb-token>
  INFLUXDB_ORG=<your-influxdb-organization>
  INFLUXDB_BUCKET=<your-influxdb-bucket>
  INFLUXDB_MEASUREMENT=<your-influxdb-measurement>
  PLC_DATA_SOURCE=<ethernet-ip|modbus>
  PLC_ETHERNET_IP_ADDRESS=<plc-ip-address>
  MODBUS_TCP_PORT=<modbus-port>
  MODBUS_REGISTER_START=<modbus-register-start>
  MODBUS_REGISTER_END=<modbus-register-end>
  PLC_POLL_MS=<poll-interval-in-ms>
  ```

## Installation

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd vtarchitect
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Create a `.env` file in the root directory and populate it with your configuration.

## Usage

### Running the Application

1. Start the application:
   ```bash
   go run main.go
   ```

2. The application will:
   - Start a Modbus TCP server (if `PLC_DATA_SOURCE=modbus`).
   - Connect to an Ethernet/IP PLC (if `PLC_DATA_SOURCE=ethernet-ip`).
   - Write PLC data to InfluxDB.
   - Start an API server on port `8080`.

### API Endpoints

- **GET /api/percentages**
  - Query boolean field percentages from InfluxDB.
  - Query parameters:
    - `bucket` (optional): InfluxDB bucket name (default: `vtrFeederData`)
    - `start` (optional): Start time for the query (default: `-1h`)
    - `stop` (optional): Stop time for the query (default: `now()`)

## Development

### Code Structure

- `main.go`: Entry point of the application.
- `api/`: Contains the API server implementation.
- `data/`: Handles PLC data mapping and communication.
- `influx/`: Manages InfluxDB client and operations.

### Running Tests

To run tests:
```bash
go test ./...
```

## License

This project is licensed under the MIT License. See the LICENSE file for details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any changes.
