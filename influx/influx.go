package influx

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"
	"vtarchitect/config"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
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
					if tag != "" {
						fields[fmt.Sprintf("%s%s.%s%d", prefix, name, tag, j+1)] = field.Index(j).Interface()
					} else {
						fields[fmt.Sprintf("%s%s[%d]", prefix, name, j)] = field.Index(j).Interface()
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
		fmt.Printf("DEBUG: Record: Field=%v, Value=%v, Values=%v\n", res.Record().Field(), res.Record().Value(), res.Record().Values())
		field, ok := res.Record().ValueByKey("_field").(string)
		if !ok {
			fmt.Printf("DEBUG: Skipping record with missing _field: %v\n", res.Record().Values())
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
