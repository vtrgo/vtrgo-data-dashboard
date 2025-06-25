package influx

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"
	"vtarchitect/config"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

type Client struct {
	influxClient influxdb2.Client
	writeAPI     api.WriteAPIBlocking
	queryAPI     api.QueryAPI
	org          string
	bucket       string
}

func NewClient(cfg *config.Config) (*Client, error) {
	url := cfg.Values["INFLUXDB_URL"]
	token := cfg.Values["INFLUXDB_TOKEN"]
	org := cfg.Values["INFLUXDB_ORG"]
	bucket := cfg.Values["INFLUXDB_BUCKET"]
	if url == "" || token == "" || org == "" || bucket == "" {
		return nil, fmt.Errorf("missing required InfluxDB configuration values")
	}
	client := influxdb2.NewClient(url, token)
	return &Client{
		influxClient: client,
		writeAPI:     client.WriteAPIBlocking(org, bucket),
		queryAPI:     client.QueryAPI(org),
		org:          org,
		bucket:       bucket,
	}, nil
}

func (c *Client) WritePoint(measurement string, tags map[string]string, fields map[string]interface{}, t time.Time) error {
	p := influxdb2.NewPoint(measurement, tags, fields, t)
	return c.writeAPI.WritePoint(context.Background(), p)
}

func (c *Client) Query(queryStr string) (*api.QueryTableResult, error) {
	return c.queryAPI.Query(context.Background(), queryStr)
}

func (c *Client) Close() {
	c.influxClient.Close()
}

func StructToInfluxFields(input any, prefix string) map[string]interface{} {
	fields := make(map[string]interface{})
	v := reflect.ValueOf(input)
	t := reflect.TypeOf(input)

	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			ft := t.Field(i)
			name := ft.Name
			if field.Kind() == reflect.Struct {
				for k, val := range StructToInfluxFields(field.Interface(), prefix+name+".") {
					fields[k] = val
				}
			} else if field.Kind() == reflect.Array || field.Kind() == reflect.Slice {
				tag := ft.Tag.Get("influx")
				for j := 0; j < field.Len(); j++ {
					elem := field.Index(j)
					if elem.Kind() == reflect.Struct {
						// Recursively flatten struct fields, prefixing with index
						for k, val := range StructToInfluxFields(elem.Interface(), fmt.Sprintf("%s%s[%d].", prefix, name, j)) {
							fields[k] = val
						}
					} else if tag != "" {
						if tag == "Fault" {
							fields[fmt.Sprintf("%s%s.%s%d", prefix, name, tag, j)] = elem.Interface()
						} else {
							fields[fmt.Sprintf("%s%s.%s%d", prefix, name, tag, j+1)] = elem.Interface()
						}
					} else if name == "VibrationDataFloats" {
						// Map specific indices to meaningful names
						fieldNames := []string{"VibrationX", "VibrationY", "VibrationZ", "Temperature"}
						for _, subName := range fieldNames {
							fields[fmt.Sprintf("%s%s[%d].%s", prefix, name, j/4, subName)] = elem.Interface()
						}
					} else {
						fields[fmt.Sprintf("%s%s[%d]", prefix, name, j)] = elem.Interface()
					}
				}
			} else {
				fields[prefix+name] = field.Interface()
			}
		}
	}
	return fields
}

// AggregateBooleanPercentages calculates the percentage of true values for specified boolean fields
// in a given time range from the specified InfluxDB bucket.
func (c *Client) AggregateBooleanPercentages(measurement, bucket string, fields []string, start, stop string) (map[string]float64, error) {
	var filters []string
	for _, f := range fields {
		filters = append(filters, fmt.Sprintf(`r["_field"] == "%s"`, f))
	}

	query := fmt.Sprintf(`
from(bucket: "%s")
  |> range(start: %s, stop: %s)
  |> filter(fn: (r) => r["_measurement"] == "%s")
  |> filter(fn: (r) => %s)
  |> map(fn: (r) => ({ r with _value: if r._value then 1.0 else 0.0 }))
  |> group(columns: ["_field"])
  |> mean()
  |> map(fn: (r) => ({ r with _value: r._value * 100.0 }))
`, bucket, start, stop, measurement, strings.Join(filters, " or "))

	res, err := c.queryAPI.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}

	percentages := make(map[string]float64)
	for res.Next() {
		field := res.Record().Field()
		if val, ok := res.Record().Value().(float64); ok {
			percentages[field] = val
		}
	}
	return percentages, res.Err()
}

// AggregateBooleanStats calculates the percentage true and time-in-true (in seconds) for specified boolean fields.
func (c *Client) AggregateBooleanStats(measurement, bucket string, fields []string, start, stop string) (map[string]struct {
	Percentage float64
	Seconds    float64
}, error) {
	var filters []string
	for _, f := range fields {
		filters = append(filters, fmt.Sprintf(`r["_field"] == "%s"`, f))
	}

	query := fmt.Sprintf(`
from(bucket: "%s")
  |> range(start: %s, stop: %s)
  |> filter(fn: (r) => r["_measurement"] == "%s")
  |> filter(fn: (r) => %s)
  |> aggregateWindow(every: 1m, fn: last)
  |> fill(usePrevious: true)
  |> map(fn: (r) => ({
      _field: r._field,
      _value: if r._value == true then 60 else 0
  }))
  |> group(columns: ["_field"])
  |> reduce(
      identity: {field: "", totalSeconds: 0, count: 0},
      fn: (r, accumulator) => ({
          field: r._field,
          totalSeconds: accumulator.totalSeconds + int(v: r._value),
          count: accumulator.count + 60
      })
  )
  |> map(fn: (r) => ({
      _field: r.field,
      percentageTrue: (float(v: r.totalSeconds) / float(v: r.count)) * 100.0,
      timeInTrue: float(v: r.totalSeconds)
  }))
`, bucket, start, stop, measurement, strings.Join(filters, " or "))

	res, err := c.queryAPI.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}

	stats := make(map[string]struct {
		Percentage float64
		Seconds    float64
	})

	for res.Next() {
		// fmt.Printf("DEBUG: Record: Field=%v, Value=%v, Values=%v\n", res.Record().Field(), res.Record().Value(), res.Record().Values())
		field, ok := res.Record().ValueByKey("_field").(string)
		if !ok {
			// fmt.Printf("DEBUG: Skipping record with missing _field: %v\n", res.Record().Values())
			continue
		}
		vals := res.Record().Values()
		percentage, _ := vals["percentageTrue"].(float64)
		timeInTrue, _ := vals["timeInTrue"].(float64)
		stats[field] = struct {
			Percentage float64
			Seconds    float64
		}{
			Percentage: percentage,
			Seconds:    timeInTrue,
		}
	}
	return stats, res.Err()
}

type ChannelBatchWriter struct {
	writeAPI api.WriteAPIBlocking
	buffer   []*write.Point
	maxSize  int
	flushCh  chan struct{}
	closeCh  chan struct{}
}

func NewChannelBatchWriter(writeAPI api.WriteAPIBlocking, maxSize int) *ChannelBatchWriter {
	cbw := &ChannelBatchWriter{
		writeAPI: writeAPI,
		buffer:   make([]*write.Point, 0, maxSize),
		maxSize:  maxSize,
		flushCh:  make(chan struct{}, 1),
		closeCh:  make(chan struct{}),
	}
	go cbw.run(5 * time.Second) // Default flush interval
	return cbw
}

func (cbw *ChannelBatchWriter) AddPoint(measurement string, tags map[string]string, fields map[string]interface{}, t time.Time) {
	p := influxdb2.NewPoint(measurement, tags, fields, t)
	cbw.buffer = append(cbw.buffer, p)
	log.Printf("DEBUG: Added point to buffer. Current buffer size: %d", len(cbw.buffer))
	if len(cbw.buffer) >= cbw.maxSize {
		log.Println("DEBUG: Buffer size reached max capacity. Triggering flush.")
		select {
		case cbw.flushCh <- struct{}{}:
		default:
		}
	}
}

func (cbw *ChannelBatchWriter) Flush() {
	if len(cbw.buffer) == 0 {
		log.Println("DEBUG: Flush called but buffer is empty. No action taken.")
		return
	}
	log.Printf("DEBUG: Flushing %d points from buffer.", len(cbw.buffer))
	points := cbw.buffer
	cbw.buffer = make([]*write.Point, 0, cbw.maxSize)
	for _, p := range points {
		if err := cbw.writeAPI.WritePoint(context.Background(), p); err != nil {
			log.Printf("DEBUG: Error writing point: %v", err)
		}
	}
	log.Println("DEBUG: Flush completed.")
}

func (cbw *ChannelBatchWriter) Close() {
	close(cbw.closeCh)
}

func (cbw *ChannelBatchWriter) run(flushInterval time.Duration) {
	log.Println("DEBUG: ChannelBatchWriter run loop started.")
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-cbw.flushCh:
			log.Println("DEBUG: Received flush signal.")
			cbw.Flush()
		case <-ticker.C:
			log.Println("DEBUG: Flush interval reached. Checking buffer.")
			if len(cbw.buffer) > 0 {
				log.Printf("DEBUG: Buffer has %d points. Triggering flush.", len(cbw.buffer))
				cbw.Flush()
			} else {
				log.Println("DEBUG: Buffer is empty. No flush needed.")
			}
		case <-cbw.closeCh:
			log.Println("DEBUG: Received close signal. Flushing remaining points and exiting.")
			cbw.Flush()
			return
		}
	}
}

func (c *Client) GetWriteAPI() api.WriteAPIBlocking {
	return c.writeAPI
}

// StatsResult holds the combined results from all three stats queries
// You may want to define more precise types for each field
// For now, use interface{} for flexibility

type StatsResult struct {
	Boolean     interface{} `json:"boolean"`
	FaultCounts interface{} `json:"faultcounts"`
	Vibration   interface{} `json:"vibration"`
}

// GetStats runs the three stats queries and aggregates their results
func (c *Client) GetStats(bucket, start, stop string) (*StatsResult, error) {
	// Query templates (use raw string literals for proper formatting)
	booleanQuery := `from(bucket: "%s")
  |> range(start: %s, stop: %s)
  |> filter(fn: (r) =>
    r._measurement == "status_data" and
    contains(value: r._field, set: [
      "FeederStatusBits.BulkHopperEnabled",
      "FeederStatusBits.BulkHopperLevelNotOK",
      "FeederStatusBits.CrossConveyorEnabled",
      "FeederStatusBits.CrossConveyorLevelNotOK",
      "FeederStatusBits.ElevatorEnabled",
      "FeederStatusBits.ElevatorLevelNotOK",
      "FeederStatusBits.EscapementAdvEnabled",
      "FeederStatusBits.EscapementRetEnabled",
      "FeederStatusBits.OrientationEnabled",
      "FeederStatusBits.OrientationLevelNotOK",
      "FeederStatusBits.TransferEnabled",
      "FeederStatusBits.TransferLevelNotOK",
      "JamStatusBits.JamInOrientation.Lane1",
      "JamStatusBits.JamInOrientation.Lane2",
      "JamStatusBits.JamInOrientation.Lane3",
      "JamStatusBits.JamInOrientation.Lane4",
      "JamStatusBits.JamInOrientation.Lane5",
      "JamStatusBits.JamInOrientation.Lane6",
      "JamStatusBits.JamInOrientation.Lane7",
      "JamStatusBits.JamInOrientation.Lane8",
      "LevelStatusBits.HighLevelLane.Lane1",
      "LevelStatusBits.HighLevelLane.Lane2",
      "LevelStatusBits.HighLevelLane.Lane3",
      "LevelStatusBits.HighLevelLane.Lane4",
      "LevelStatusBits.HighLevelLane.Lane5",
      "LevelStatusBits.HighLevelLane.Lane6",
      "LevelStatusBits.HighLevelLane.Lane7",
      "LevelStatusBits.HighLevelLane.Lane8",
      "LevelStatusBits.NotLowLevelLane.Lane1",
      "LevelStatusBits.NotLowLevelLane.Lane2",
      "LevelStatusBits.NotLowLevelLane.Lane3",
      "LevelStatusBits.NotLowLevelLane.Lane4",
      "LevelStatusBits.NotLowLevelLane.Lane5",
      "LevelStatusBits.NotLowLevelLane.Lane6",
      "LevelStatusBits.NotLowLevelLane.Lane7",
      "LevelStatusBits.NotLowLevelLane.Lane8"
    ])
  )
  |> map(fn: (r) => ({ r with _value: if r._value == true then 1.0 else 0.0 }))
  |> group(columns: ["_field"])
  |> mean(column: "_value")
  |> yield(name: "boolean")`

	faultCountsQuery := `from(bucket: "%s")
  |> range(start: %s, stop: %s)
  |> filter(fn: (r) =>
      r._measurement == "status_data" and
      r._field =~ /FaultStatusBits.FaultArray[01].Fault([0-9]|1[0-5])/)
  |> map(fn: (r) => ({ r with _value: if r._value then 1.0 else 0.0 }))
  |> group(columns: ["_field"])
  |> sum(column: "_value")
  |> yield(name: "fault_true_counts")`

	vibrationQuery := `from(bucket: "%s")
  |> range(start: %s, stop: %s)
  |> filter(fn: (r) =>
      r._measurement == "status_data" and
      contains(value: r._field, set: [
        "VibrationDataFloats[0].VibrationX",
        "VibrationDataFloats[0].VibrationY",
        "VibrationDataFloats[0].VibrationZ",
        "VibrationDataFloats[0].Temperature",
        "VibrationDataFloats[1].VibrationX",
        "VibrationDataFloats[1].VibrationY",
        "VibrationDataFloats[1].VibrationZ",
        "VibrationDataFloats[1].Temperature",
        "VibrationDataFloats[2].VibrationX",
        "VibrationDataFloats[2].VibrationY",
        "VibrationDataFloats[2].VibrationZ",
        "VibrationDataFloats[2].Temperature",
        "VibrationDataFloats[3].VibrationX",
        "VibrationDataFloats[3].VibrationY",
        "VibrationDataFloats[3].VibrationZ",
        "VibrationDataFloats[3].Temperature",
        "VibrationDataFloats[4].VibrationX",
        "VibrationDataFloats[4].VibrationY",
        "VibrationDataFloats[4].VibrationZ",
        "VibrationDataFloats[4].Temperature"
      ])
  )
  |> group(columns: ["_field"])
  |> mean(column: "_value")
  |> map(fn: (r) => ({ _time: %s, _measurement: r._field, _value: r._value }))
  |> yield(name: "mean_over_range")`

	// Format queries with parameters
	bq := fmt.Sprintf(booleanQuery, bucket, start, stop)
	fq := fmt.Sprintf(faultCountsQuery, bucket, start, stop)
	vq := fmt.Sprintf(vibrationQuery, bucket, start, stop, stop)

	// Helper to parse results into a generic slice of maps
	parseResults := func(res *api.QueryTableResult) ([]map[string]interface{}, error) {
		var out []map[string]interface{}
		for res.Next() {
			row := make(map[string]interface{})
			for k, v := range res.Record().Values() {
				row[k] = v
			}
			out = append(out, row)
		}
		return out, res.Err()
	}

	// Run all three queries with error logging
	bRes, err := c.queryAPI.Query(context.Background(), bq)
	if err != nil {
		log.Printf("Error running boolean query: %v", err)
		return nil, fmt.Errorf("boolean query error: %w", err)
	}
	booleanData, err := parseResults(bRes)
	if err != nil {
		log.Printf("Error parsing boolean query results: %v", err)
		return nil, fmt.Errorf("boolean parse error: %w", err)
	}

	fRes, err := c.queryAPI.Query(context.Background(), fq)
	if err != nil {
		log.Printf("Error running faultcounts query: %v", err)
		return nil, fmt.Errorf("faultcounts query error: %w", err)
	}
	faultData, err := parseResults(fRes)
	if err != nil {
		log.Printf("Error parsing faultcounts query results: %v", err)
		return nil, fmt.Errorf("faultcounts parse error: %w", err)
	}
	log.Printf("DEBUG: faultData = %+v", faultData)
	if faultData == nil {
		faultData = make([]map[string]interface{}, 0)
	}

	vRes, err := c.queryAPI.Query(context.Background(), vq)
	if err != nil {
		log.Printf("Error running vibration query: %v", err)
		return nil, fmt.Errorf("vibration query error: %w", err)
	}
	vibrationData, err := parseResults(vRes)
	if err != nil {
		log.Printf("Error parsing vibration query results: %v", err)
		return nil, fmt.Errorf("vibration parse error: %w", err)
	}

	return &StatsResult{
		Boolean:     booleanData,
		FaultCounts: faultData,
		Vibration:   vibrationData,
	}, nil
}

// AggregateFaultCounts sums the number of true values for each fault field in the given time range.
func (c *Client) AggregateFaultCounts(measurement, bucket string, fields []string, start, stop string) (map[string]float64, error) {
	if len(fields) == 0 {
		return map[string]float64{}, nil
	}
	var filters []string
	for _, f := range fields {
		filters = append(filters, fmt.Sprintf(`r["_field"] == "%s"`, f))
	}
	query := fmt.Sprintf(`
from(bucket: "%s")
  |> range(start: %s, stop: %s)
  |> filter(fn: (r) => r["_measurement"] == "%s")
  |> filter(fn: (r) => %s)
  |> map(fn: (r) => ({ r with _value: if r._value then 1.0 else 0.0 }))
  |> group(columns: ["_field"])
  |> sum()
`, bucket, start, stop, measurement, strings.Join(filters, " or "))

	res, err := c.queryAPI.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	counts := make(map[string]float64)
	for res.Next() {
		field := res.Record().Field()
		if val, ok := res.Record().Value().(float64); ok {
			counts[field] = val
		}
	}
	return counts, res.Err()
}

// AggregateFloatMeans computes the mean value for each float field in the given time range.
func (c *Client) AggregateFloatMeans(measurement, bucket string, fields []string, start, stop string) (map[string]float64, error) {
	if len(fields) == 0 {
		return map[string]float64{}, nil
	}
	var filters []string
	for _, f := range fields {
		filters = append(filters, fmt.Sprintf(`r["_field"] == "%s"`, f))
	}
	query := fmt.Sprintf(`
from(bucket: "%s")
  |> range(start: %s, stop: %s)
  |> filter(fn: (r) => r["_measurement"] == "%s")
  |> filter(fn: (r) => %s)
  |> group(columns: ["_field"])
  |> mean()
`, bucket, start, stop, measurement, strings.Join(filters, " or "))

	res, err := c.queryAPI.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	means := make(map[string]float64)
	for res.Next() {
		field := res.Record().Field()
		if val, ok := res.Record().Value().(float64); ok {
			means[field] = val
		}
	}
	return means, res.Err()
}

// GetFloatRange queries a specific float field over a given time range and returns time-value pairs.
func (c *Client) GetFloatRange(bucket, field, start, stop string) ([]map[string]interface{}, error) {
	query := fmt.Sprintf(`
from(bucket: "%s")
  |> range(start: %s, stop: %s)
  |> filter(fn: (r) => r["_measurement"] == "status_data" and r["_field"] == "%s")
  |> keep(columns: ["_time", "_value"])
`, bucket, start, stop, field)

	res, err := c.queryAPI.Query(context.Background(), query)
	if err != nil {
		log.Printf("Error running float range query for field '%s': %v", field, err)
		return nil, fmt.Errorf("float range query error: %w", err)
	}

	var data []map[string]interface{}
	for res.Next() {
		record := res.Record()
		row := make(map[string]interface{})
		// InfluxDB _time is a time.Time object, _value is float64
		row["time"] = record.Time()
		row["value"] = record.Value()
		data = append(data, row)
	}

	if res.Err() != nil {
		log.Printf("Error parsing float range query results for field '%s': %v", field, res.Err())
		return nil, fmt.Errorf("float range parse error: %w", res.Err())
	}

	log.Printf("Successfully fetched %d points for float field '%s'", len(data), field)
	return data, nil
}
