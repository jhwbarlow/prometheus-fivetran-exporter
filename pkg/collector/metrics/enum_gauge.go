package metrics

import (
	"fmt"
	"strings"
)

type EnumGaugeValue struct {
	name  string
	value int
}

func (gv *EnumGaugeValue) GaugeValue() float64 {
	return float64(gv.value)
}

func (gv *EnumGaugeValue) DisplayValue() int {
	return gv.value
}

func (gv *EnumGaugeValue) Name() string {
	return gv.name
}

func NewEnumGaugeValue(name string, value int) *EnumGaugeValue {
	return &EnumGaugeValue{
		name:  name,
		value: value,
	}
}

type EnumGauge struct {
	allValues   []*EnumGaugeValue
	description string
}

func NewEnumGauge(allValues []*EnumGaugeValue, description string) *EnumGauge {
	return &EnumGauge{
		allValues:   allValues,
		description: description,
	}
}

func (eg *EnumGauge) Describe() string {
	builder := new(strings.Builder)
	builder.WriteString(fmt.Sprintf("%s. Values: ", eg.description))
	for i, v := range eg.allValues {
		builder.WriteString(fmt.Sprintf("%d=%s", v.DisplayValue(), v.Name()))
		if i < len(eg.allValues)-1 {
			builder.WriteString(", ")
		}
	}

	return builder.String()
}

var (
	EnumGaugeValueFalse       = NewEnumGaugeValue("false", 0)
	EnumGaugeValueTrue        = NewEnumGaugeValue("true", 1)
	BooleanMetricsGaugeValues = []*EnumGaugeValue{
		EnumGaugeValueFalse,
		EnumGaugeValueTrue,
	}

	EnumGaugeValuePresent     = NewEnumGaugeValue("present", 1)
	PresentMetricsGaugeValues = []*EnumGaugeValue{EnumGaugeValuePresent}
)
