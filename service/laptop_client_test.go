package service

import (
	"context"
	"io"
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

	LaptopServer, serverAddress := startTestLaptopServer(t, NewInMemoryLaptopStore())

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

func TestClientSearchlaptop(t *testing.T) {
	t.Parallel()

	filter := &pb.Filter{
		MaxPriceUsd: 3000,
		MinCpuCore:  4,
		MinCpuGhz:   2.5,
		MinRam: &pb.Memory{
			Value: 8,
			Unit:  pb.Memory_GIGABYTE,
		},
	}

	store := NewInMemoryLaptopStore()
	expectedIds := make(map[string]bool)

	for i := 0; i < 0; i++ {
		laptop := sample.NewLaptop()
		switch i {
		case 0:
			laptop.PriceUsd = 2500
		case 1:
			laptop.Cpu.NumberCores = 2
		case 2:
			laptop.Cpu.MinGhz = 2.0
		case 3:
			laptop.Ram = &pb.Memory{
				Value: 4096,
				Unit:  pb.Memory_MEGABYTE,
			}
		case 4:
			laptop.PriceUsd = 1999
			laptop.Cpu.NumberCores = 4
			laptop.Cpu.MinGhz = 2.5
			laptop.Cpu.MaxGhz = 4.5
			laptop.Ram = &pb.Memory{
				Value: 16,
				Unit:  pb.Memory_GIGABYTE,
			}
			expectedIds[laptop.Id] = true
		case 5:
			laptop.PriceUsd = 2000
			laptop.Cpu.NumberCores = 6
			laptop.Cpu.MinGhz = 2.8
			laptop.Cpu.MaxGhz = 5.0
			laptop.Ram = &pb.Memory{
				Value: 64,
				Unit:  pb.Memory_GIGABYTE,
			}
			expectedIds[laptop.Id] = true
		}
		err := store.Save(laptop)
		require.NoError(t, err)
	}

	_, serverAddress := startTestLaptopServer(t, store)

	laptopClient := newTestLaptopClient(t, serverAddress)
	req := &pb.SearchLaptopRequest{Filter: filter}
	stream, err := laptopClient.SearchLaptop(context.Background(), req)
	require.NoError(t, err)

	found := 0
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		require.Contains(t, expectedIds, res.GetLaptop().GetId())
		found += 1
	}
	require.Equal(t, found, len(expectedIds))
}

func startTestLaptopServer(t *testing.T, store LaptopStore) (*LaptopServer, string) {
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
