package service

import (
	"context"
	"net"
	"testing"

	"example.com/pcbook/pb"
	"example.com/pcbook/sample"
	"example.com/pcbook/serializer"
	"github.com/test-go/testify/require"
	"google.golang.org/grpc"
)

func TestClientCreatelaptop(t *testing.T) {
	t.Parallel()

	LaptopServer, serverAddress := startTestLaptopServer(t)

	laptopClient := newTestLaptopClient(t, serverAddress)

	laptop := sample.NewLaptop()

	expectedId := laptop.Id
	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}
	res, err := laptopClient.CreateLaptop(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, res.Id, expectedId)

	// check laptop is saved to server
	other, err := LaptopServer.Store.Find(res.Id)
	require.NoError(t, err)
	require.NotNil(t, other)

	//check that the saved latpop is the same as the one send
	requireSameLaptop(t, laptop, other)

}

func startTestLaptopServer(t *testing.T) (*LaptopServer, string) {
	laptopServer := NewLaptopServer(NewInMemoryLaptopStore())
	grpcServer := grpc.NewServer()
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)

	listener, err := net.Listen("tcp", ":0")

	require.NoError(t, err)

	go grpcServer.Serve(listener)

	return laptopServer, listener.Addr().String()
}

func newTestLaptopClient(t *testing.T, serverAddress string) pb.LaptopServiceClient {
	conn, err := grpc.Dial(serverAddress, grpc.WithInsecure())

	require.NoError(t, err)
	return pb.NewLaptopServiceClient(conn)
}

func requireSameLaptop(t *testing.T, laptop1, laptop2 *pb.Laptop) {
	json1, err := serializer.ProtobuffToJSON(laptop1)
	require.NoError(t, err)

	json2, err := serializer.ProtobuffToJSON(laptop2)
	require.NoError(t, err)

	require.Equal(t, json1, json2)
}
