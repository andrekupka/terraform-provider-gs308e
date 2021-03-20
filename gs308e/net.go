package gs308e

import (
	"fmt"
	"net"
)

func maskToString(mask net.IPMask) string {
	return fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3])
}
