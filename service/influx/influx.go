// file: service/influx/influx.go
// InfluxDB client and utilities for writing and querying data
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
		record := res.Record()
		if field, ok := record.ValueByKey("_field").(string); ok {
			if val, ok := record.Value().(float64); ok {
				percentages[field] = val
			}
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
	log.Printf("INFLUX: Added point to buffer. Current buffer size: %d", len(cbw.buffer))
	if len(cbw.buffer) >= cbw.maxSize {
		log.Println("INFLUX: Buffer size reached max capacity. Triggering flush.")
		select {
		case cbw.flushCh <- struct{}{}:
		default:
		}
	}
}

func (cbw *ChannelBatchWriter) Flush() {
	if len(cbw.buffer) == 0 {
		log.Println("INFLUX: Flush called but buffer is empty. No action taken.")
		return
	}
	log.Printf("INFLUX: Flushing %d points from buffer.", len(cbw.buffer))
	points := cbw.buffer
	cbw.buffer = make([]*write.Point, 0, cbw.maxSize)
	if err := cbw.writeAPI.WritePoint(context.Background(), points...); err != nil {
		log.Printf("INFLUX: Error writing points in batch: %v", err)
	}
	log.Println("INFLUX: Flush completed.")
}

func (cbw *ChannelBatchWriter) Close() {
	close(cbw.closeCh)
}

func (cbw *ChannelBatchWriter) run(flushInterval time.Duration) {
	log.Println("INFLUX: ChannelBatchWriter run loop started.")
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-cbw.flushCh:
			log.Println("INFLUX: Received flush signal.")
			cbw.Flush()
		case <-ticker.C:
			log.Println("INFLUX: Flush interval reached. Checking buffer.")
			if len(cbw.buffer) > 0 {
				log.Printf("INFLUX: Buffer has %d points. Triggering flush.", len(cbw.buffer))
				cbw.Flush()
			} else {
				log.Println("INFLUX: Buffer is empty. No flush needed.")
			}
		case <-cbw.closeCh:
			log.Println("INFLUX: Received close signal. Flushing remaining points and exiting.")
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

// AggregateFaultCounts counts the number of false-to-true transitions for each fault field.
// This accurately reflects the number of times a fault occurred, rather than how many
// polling cycles it was active for.
func (c *Client) AggregateFaultCounts(measurement, bucket string, fields []string, start, stop string) (map[string]float64, error) {
	if len(fields) == 0 {
		return map[string]float64{}, nil
	}
	var filters []string
	for _, f := range fields {
		filters = append(filters, fmt.Sprintf(`r["_field"] == "%s"`, f))
	}
	// This query correctly counts fault occurrences, including faults that are
	// persistently true throughout the time range. It works by combining two sets of data:
	// 1. `transitions`: Counts the number of times a fault changes from `false` to `true`.
	// 2. `initial_trues`: Identifies faults that were already in a `true` state at the
	//    very beginning of the time range.
	// By summing these two counts, we get a total number of fault occurrences.
	query := fmt.Sprintf(`
transitions = from(bucket: "%[1]s")
  |> range(start: %[2]s, stop: %[3]s)
  |> filter(fn: (r) => r["_measurement"] == "%[4]s" and (%[5]s))
  |> sort(columns: ["_time"])
  |> map(fn: (r) => ({ r with _value: if r._value then 1 else 0 }))
  |> difference(nonNegative: false, columns: ["_value"])
  |> filter(fn: (r) => r._value == 1)
  |> group(columns: ["_field"])
  |> count()
  |> rename(columns: {_value: "count"})

initial_trues = from(bucket: "%[1]s")
  |> range(start: %[2]s, stop: %[3]s)
  |> filter(fn: (r) => r["_measurement"] == "%[4]s" and (%[5]s))
  |> group(columns: ["_field"])
  |> first()
  |> filter(fn: (r) => r._value == true)
  |> map(fn: (r) => ({_field: r._field, count: 1}))
  |> keep(columns: ["_field", "count"])

union(tables: [transitions, initial_trues])
  |> group(columns: ["_field"])
  |> sum(column: "count")
  |> rename(columns: {count: "_value"})
  |> group()
`, bucket, start, stop, measurement, strings.Join(filters, " or "))

	res, err := c.queryAPI.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	counts := make(map[string]float64)
	for res.Next() {
		record := res.Record()
		// The result of the query has a `_field` column with the fault name.
		field, ok := record.ValueByKey("_field").(string)
		if !ok {
			continue // Should not happen with this query structure
		}

		// The count() function returns an int64 value in the `_value` column.
		if val, ok := record.Value().(int64); ok {
			counts[field] = float64(val)
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
		record := res.Record()
		if field, ok := record.ValueByKey("_field").(string); ok {
			if val, ok := record.Value().(float64); ok {
				means[field] = val
			}
		}
	}
	return means, res.Err()
}

// GetFloatRange queries a specific float field over a given time range and returns time-value pairs.
func (c *Client) GetFloatRange(bucket, field, start, stop string) ([]map[string]interface{}, error) {
	window := inferWindowSize(start) // 👈 new logic here

	query := fmt.Sprintf(`
from(bucket: "%s")
  |> range(start: %s, stop: %s)
  |> filter(fn: (r) => r["_measurement"] == "status_data" and r["_field"] == "%s")
  |> aggregateWindow(every: %s, fn: mean, createEmpty: false)
  |> keep(columns: ["_time", "_value"])
`, bucket, start, stop, field, window)

	res, err := c.queryAPI.Query(context.Background(), query)
	if err != nil {
		log.Printf("ERROR: Error running float range query for field '%s': %v", field, err)
		return nil, fmt.Errorf("ERROR: float range query error: %w", err)
	}

	var data []map[string]interface{}
	for res.Next() {
		record := res.Record()
		row := map[string]interface{}{
			"time":  record.Time(),
			"value": record.Value(),
		}
		data = append(data, row)
	}
	if res.Err() != nil {
		log.Printf("ERROR: Error parsing float range query results for field '%s': %v", field, res.Err())
		return nil, fmt.Errorf("ERROR: Float range parse error: %w", res.Err())
	}
	log.Printf("INFLUX: Successfully fetched %d points for float field '%s'", len(data), field)
	return data, nil
}

func inferWindowSize(start string) string {
	switch {
	case strings.HasPrefix(start, "-1h"):
		return "1s"
	case strings.HasPrefix(start, "-3h"):
		return "30s"
	case strings.HasPrefix(start, "-6h"):
		return "10s"
	case strings.HasPrefix(start, "-12h"):
		return "1m"
	case strings.HasPrefix(start, "-1d"):
		return "1m"
	case strings.HasPrefix(start, "-2d"):
		return "2m"
	case strings.HasPrefix(start, "-3d"):
		return "3m"
	case strings.HasPrefix(start, "-1w"):
		return "10m"
	case strings.HasPrefix(start, "-2w"):
		return "10m"
	case strings.HasPrefix(start, "-3w"):
		return "10m"
	case strings.HasPrefix(start, "-1mo"):
		return "10m"
	default:
		return "5m" // fallback for short or malformed inputs
	}
}

// GetSystemStatus retrieves the most recent boolean value for each of the system status fields.
func (c *Client) GetSystemStatus(measurement, bucket string, fields []string) (map[string]bool, error) {
	if len(fields) == 0 {
		return map[string]bool{}, nil
	}
	var filters []string
	for _, f := range fields {
		filters = append(filters, fmt.Sprintf(`r["_field"] == "%s"`, f))
	}
	query := fmt.Sprintf(`
from(bucket: "%s")
  |> range(start: -1d) // Limit to the last day to make the query faster
  |> filter(fn: (r) => r["_measurement"] == "%s")
  |> filter(fn: (r) => %s)
  |> group(columns: ["_field"])
  |> last()
`, bucket, measurement, strings.Join(filters, " or "))

	res, err := c.queryAPI.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}

	statuses := make(map[string]bool)
	for res.Next() {
		record := res.Record()
		field, ok := record.ValueByKey("_field").(string)
		if !ok {
			continue
		}
		value, ok := record.Value().(bool)
		if !ok {
			continue
		}
		statuses[field] = value
	}
	return statuses, res.Err()
}
