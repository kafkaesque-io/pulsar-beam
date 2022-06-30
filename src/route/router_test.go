package route

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEffectiveRoutes(t *testing.T) {
	receiverRoutesLen := len(ReceiverRoutes)
	restRoutesLen := len(RestRoutes)
	prometheusLen := len(PrometheusRoute)
	// mode := "hybrid"
	// assert.Equal(t, len(GetEffectiveRoutes(&mode)), (receiverRoutesLen + restRoutesLen + prometheusLen))
	mode := "rest"
	assert.Equal(t, len(GetEffectiveRoutes(&mode)), (receiverRoutesLen + restRoutesLen + prometheusLen))
	mode = "receiver"
	assert.Equal(t, len(GetEffectiveRoutes(&mode)), (receiverRoutesLen + restRoutesLen + prometheusLen))
}
