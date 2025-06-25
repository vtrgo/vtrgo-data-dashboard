# VTArchitect

VTArchitect is a Go-based application that implements a Modbus TCP server and interacts with an InfluxDB database. It collects data from PLCs (Programmable Logic Controllers) and writes it to InfluxDB. Additionally, it queries InfluxDB for boolean field percentages over the last minute and logs the results.

---

## **Key Features**
1. **Modbus TCP Server**:
   - The server listens on a configurable port (default: `5020`) and reads data from Modbus registers.
   - It processes the data and writes it to InfluxDB.

2. **Ethernet/IP Support**:
   - The application can connect to a PLC using Ethernet/IP and read/write tags.

3. **InfluxDB Integration**:
   - The application writes data to InfluxDB and queries it for boolean field percentages.
   - It uses environment variables for configuration, such as `INFLUXDB_URL`, `INFLUXDB_TOKEN`, `INFLUXDB_ORG`, and `INFLUXDB_BUCKET`.

4. **API Server**:
   - An API server is implemented to expose endpoints for querying boolean field percentages from InfluxDB.

---

## **Requirements**
### **1. Connected PLC**
- The application requires a connected PLC to function properly.
- Supported PLC communication protocols:
  - **Modbus TCP**: The PLC must support Modbus TCP communication.
  - **Ethernet/IP**: If using Ethernet/IP, ensure the PLC is configured to allow tag-based communication.
- Configure the PLC connection details in the `.env` file:
  ```env
  PLC_DATA_SOURCE=modbus  # or ethernet-ip
  PLC_ETHERNET_IP_ADDRESS=<plc-ip-address>  # for Ethernet/IP only
  MODBUS_TCP_PORT=<modbus-port>  # for Modbus TCP only
  MODBUS_REGISTER_START=<modbus-register-start>  # for Modbus TCP only
  MODBUS_REGISTER_END=<modbus-register-end>  # for Modbus TCP only
  PLC_POLL_MS=<poll-interval-in-ms>
  ```

### **2. InfluxDB**
- An InfluxDB instance is required to store and query the data.
- Configure the InfluxDB connection details in the `.env` file:
  ```env
  INFLUXDB_URL=<your-influxdb-url>
  INFLUXDB_TOKEN=<your-influxdb-token>
  INFLUXDB_ORG=<your-influxdb-organization>
  INFLUXDB_BUCKET=<your-influxdb-bucket>
  INFLUXDB_MEASUREMENT=<your-influxdb-measurement>
  ```

---

## **Installation**
1. Clone the repository:
   ```bash
   git clone https://github.com/vtrgo/vtarchitect.git
   cd vtarchitect
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Create a `.env` file in the root directory and populate it with your configuration.

---

## **Usage**
### **Running the Application**
1. Start the application:
   ```bash
   go run main.go
   ```

2. The application will:
   - Start a Modbus TCP server (if `PLC_DATA_SOURCE=modbus`).
   - Connect to an Ethernet/IP PLC (if `PLC_DATA_SOURCE=ethernet-ip`).
   - Write PLC data to InfluxDB.
   - Start an API server on port `8080`.

### **API Endpoints**
- **GET /api/percentages**
  - Query boolean field percentages from InfluxDB.
  - Query parameters:
    - `bucket` (optional): InfluxDB bucket name (default: `vtrFeederData`)
    - `start` (optional): Start time for the query (default: `-1h`)
    - `stop` (optional): Stop time for the query (default: `now()`)

---

## **Development**
### **Code Structure**

```text
vtarchitect/
├── .env
├── .gitignore
├── LICENSE
├── README.md
├── MERMAID.md
├── TODO.md
├── ...
│
├── service/ # Go data service application
│   ├── main.go
│   ├── go.mod
│   ├── api/
│   ├── config/
│   ├── data/
│   ├── influx/
│   ├── tools/
│   └── ...
│
├── console/ # React + Vite frontend application
│   ├── public/
│   │   ├── textures/
│   ├── src/
│   │   ├── assets/
│   │   ├── components/
│   │   ├── hooks/
│   │   ├── lib/
│   │   ├── pages/
│   │   ├── utils/
│   │   ├── App.tsx
│   │   ├── main.tsx
│   │   └── index.css
│   ├── index.html
│   ├── package.json
│   └── vite.config.js
│
└── shared/
    ├── architect.yaml
    ├── go-import-tag.csv
    └── ...
```

### **Running Tests**
To run tests:
```bash
go test ./...
```

---

## **License**
This project is licensed under the MIT License. See the LICENSE file for details.
