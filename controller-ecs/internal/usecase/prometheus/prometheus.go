package prometheus

import (
	"io"
	"runner-controller-ecs/internal/domain/model"
	"runner-controller-ecs/internal/usecase"
	"strings"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

type prometheusUC struct{}

func NewPrometheusUC() usecase.IPrometheusUC {
	return &prometheusUC{}
}

func (c *prometheusUC) Unmarshal(readers map[string]io.Reader) (map[string]map[string]*dto.MetricFamily, error) {
	res := make(map[string]map[string]*dto.MetricFamily, 0)
	for k, reader := range readers {
		var parser expfmt.TextParser

		mf, err := parser.TextToMetricFamilies(reader)
		if err != nil {
			return nil, err
		}
		res[k] = mf
	}
	return res, nil
}

func (c *prometheusUC) Marshal(mf map[string]*dto.MetricFamily) (string, error) {
	b := new(strings.Builder)
	for _, v := range mf {
		_, err := expfmt.MetricFamilyToText(b, v)
		if err != nil {
			return "", err
		}
	}
	return b.String(), nil
}

func (c *prometheusUC) ConvertToMap(readers map[string]io.Reader) (map[string]model.Metrics, error) {
	mfs, err := c.Unmarshal(readers)
	if err != nil {
		return nil, err
	}

	resMap := make(map[string]model.Metrics)
	for name, mf := range mfs {
		res := make(model.Metrics)
		res["timestamp"] = float64(time.Now().Unix())
		for k, v := range mf {
			if k == "ecs_cpu_seconds_total" {
				sum := 0.0
				counter := 0
				skip := true
				for _, m := range v.Metric {
					for _, l := range m.Label {
						if *l.Name == "container" && *l.Value == "github-runner" {
							skip = false
							break
						}
					}
					if skip {
						continue
					}
					sum += m.GetCounter().GetValue()
					counter++
				}
				if counter > 0 {
					res[k] = sum / float64(counter)
				} else {
					res[k] = 0
				}
			}
			for _, m := range v.Metric {
				if k == "ecs_cpu_seconds_total" {
					continue
				}
				skip := true
				for _, l := range m.Label {
					if *l.Name == "container" && *l.Value == "github-runner" {
						skip = false
						break
					}
				}
				if skip {
					continue
				}

				switch v.GetType() {
				case dto.MetricType_GAUGE:
					res[k] = m.GetGauge().GetValue()
				case dto.MetricType_COUNTER:
					res[k] = m.GetCounter().GetValue()
				default:
				}
			}
		}
		resMap[name] = res
	}
	return resMap, nil
}

func (c *prometheusUC) Combine(readers map[string]io.Reader) (string, error) {
	mfs, err := c.Unmarshal(readers)
	if err != nil {
		return "", err
	}

	resMf := make(map[string]*dto.MetricFamily)
	for name, mf := range mfs {
		for k, v := range mf {
			metrics := make([]*dto.Metric, 0)
			for _, m := range v.Metric {
				skip := true
				for _, l := range m.Label {
					if *l.Name == "container" && *l.Value == "github-runner" {
						skip = false
						n := strings.Clone(name)
						l.Value = &n
						metrics = append(metrics, m)
						break
					}
				}
				if skip {
					continue
				}
			}
			if _, ok := resMf[k]; !ok {
				v.Metric = metrics
				resMf[k] = v
			} else {
				resMf[k].Metric = append(resMf[k].Metric, metrics...)
			}

		}
	}
	marshal, err := c.Marshal(resMf)
	if err != nil {
		return "", err
	}
	return marshal, nil
}
