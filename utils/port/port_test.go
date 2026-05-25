package port_test

import (
	"net"
	"runtime"
	"testing"

	"github.com/IceWhaleTech/CasaOS-Common/utils/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPortAvailable(t *testing.T) {
	udpPort, err := port.GetAvailablePort("udp")
	require.NoError(t, err)
	assert.True(t, port.IsPortAvailable(udpPort, "udp"))

	tcpPort, err := port.GetAvailablePort("tcp")
	require.NoError(t, err)
	assert.True(t, port.IsPortAvailable(tcpPort, "tcp"))
}

func TestPorts(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("ListPortsInUse reads Linux /proc/net files")
	}

	tcpListener, err := net.Listen("tcp4", "127.0.0.1:0")
	require.NoError(t, err)
	defer tcpListener.Close()
	tcpPort := tcpListener.Addr().(*net.TCPAddr).Port

	udpListener, err := net.ListenPacket("udp4", "127.0.0.1:0")
	require.NoError(t, err)
	defer udpListener.Close()
	udpPort := udpListener.LocalAddr().(*net.UDPAddr).Port

	tcpPorts, udpPorts, err := port.ListPortsInUse()
	require.NoError(t, err)

	assert.Contains(t, tcpPorts, tcpPort)
	assert.Contains(t, udpPorts, udpPort)
}
