package pkg

import (
	"bytes"
	"fmt"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

type MetricDesc struct {
	help       string
	metricType dto.MetricType
}

var (
	metricHelp = map[string]*MetricDesc{
		"kubernetes_query_resources_total": {"number of resources matching the query", dto.MetricType_GAUGE},
	}
)

func AddNewMetric(metrics []*dto.MetricFamily, metricType string, value uint64, timestampMS int64) []*dto.MetricFamily {
	metricDesc := metricHelp[metricType]
	if metricDesc == nil {
		return metrics
	}

	family := &dto.MetricFamily{
		Name:   &metricType,
		Help:   &metricDesc.help,
		Type:   &metricDesc.metricType,
		Metric: []*dto.Metric{},
	}

	if metricDesc.metricType == dto.MetricType_COUNTER {
		addNewCounterMetric(family, float64(value), timestampMS)
	} else if metricDesc.metricType == dto.MetricType_GAUGE {
		addNewGaugeMetric(family, float64(value), timestampMS)
	}
	return append(metrics, family)
}

func addNewCounterMetric(family *dto.MetricFamily, value float64, timestampMS int64) {
	counter := &dto.Metric{
		Label: []*dto.LabelPair{},
		Counter: &dto.Counter{
			Value: &value,
		},
		TimestampMs: &timestampMS,
	}
	family.Metric = append(family.Metric, counter)
}

func addNewGaugeMetric(family *dto.MetricFamily, value float64, timestampMS int64) {
	gauge := &dto.Metric{
		Label: []*dto.LabelPair{},
		Gauge: &dto.Gauge{
			Value: &value,
		},
		TimestampMs: &timestampMS,
	}
	family.Metric = append(family.Metric, gauge)
}

func PrintMetrics(metrics []*dto.MetricFamily) error {
	var buf bytes.Buffer
	for _, family := range metrics {
		buf.Reset()
		encoder := expfmt.NewEncoder(&buf, expfmt.FmtText)
		err := encoder.Encode(family)
		if err != nil {
			return err
		}

		fmt.Print(buf.String())
	}

	return nil
}
