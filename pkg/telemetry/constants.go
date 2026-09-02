package telemetry

var (
	DefaultObjectives = map[float64]float64{
		0.5:   0.01,
		0.95:  0.001,
		0.99:  0.001,
		0.999: 0.0001,
		1.0:   0,
	}

	DefaultHistogramBuckets = []float64{
		0.001,
		0.01,
		0.1,
		0.3,
		0.6,
		1,
		3,
		6,
		9,
		20,
		30,
		60,
		90,
		120,
	}
)

// ErrLabel is error static label.
const ErrLabel = "error"

// ErrLabelValue returns string representation of error label value.
func ErrLabelValue(err error) string {
	if err != nil {
		return "true"
	}

	return "false"
}
