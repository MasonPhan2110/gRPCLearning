package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"

	"example.com/pcbook/pb"
	"example.com/pcbook/sample"
	"example.com/pcbook/serializer"
	"github.com/test-go/testify/require"
	"google.golang.org/grpc"
)

func TestClientCreatelaptop(t *testing.T) {
	t.Parallel()

	laptopStore := NewInMemoryLaptopStore()

	serverAddress := startTestLaptopServer(t, laptopStore, nil, nil)

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
	other, err := laptopStore.Find(res.Id)
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

	laptopStore := NewInMemoryLaptopStore()
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
		err := laptopStore.Save(laptop)
		require.NoError(t, err)
	}

	serverAddress := startTestLaptopServer(t, laptopStore, nil, nil)

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

func TestClientUploadImage(t *testing.T) {
	t.Parallel()

	testImgFolder := "../tmp"

	laptopStore := NewInMemoryLaptopStore()
	imageStore := NewDiskImageStore(testImgFolder)

	laptop := sample.NewLaptop()
	err := laptopStore.Save(laptop)
	require.NoError(t, err)

	serverAddress := startTestLaptopServer(t, laptopStore, imageStore, nil)
	laptopClient := newTestLaptopClient(t, serverAddress)

	imagePath := fmt.Sprintf("%s/laptop.jpg", testImgFolder)

	file, err := os.Open(imagePath)
	require.NoError(t, err)
	defer file.Close()

	stream, err := laptopClient.UploadImage(context.Background())
	require.NoError(t, err)
	imageType := filepath.Ext(imagePath)

	req := &pb.UploadImageRequest{
		Data: &pb.UploadImageRequest_Info{
			Info: &pb.ImageInfo{
				LaptopId:  laptop.GetId(),
				ImageType: imageType,
			},
		},
	}

	err = stream.Send(req)
	require.NoError(t, err)

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)
	size := 0

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		size += n
		req := &pb.UploadImageRequest{
			Data: &pb.UploadImageRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}
		err = stream.Send(req)
		require.NoError(t, err)
	}
	res, err := stream.CloseAndRecv()
	require.NoError(t, err)
	require.NotZero(t, res.GetId())
	require.EqualValues(t, size, res.GetSize())

	saveImagePath := fmt.Sprintf("%s/%s%s", testImgFolder, res.GetId(), imageType)
	_, err = os.Stat(saveImagePath)

	require.NoError(t, err)
	require.NoError(t, os.Remove(saveImagePath))

}

func TestClientRateLaptop(t *testing.T) {
	t.Parallel()

	laptopStore := NewInMemoryLaptopStore()
	ratingStore := NewInMemoryRatingStore()

	laptop := sample.NewLaptop()
	err := laptopStore.Save(laptop)
	require.NoError(t, err)

	serverAddress := startTestLaptopServer(t, laptopStore, nil, ratingStore)
	laptopClient := newTestLaptopClient(t, serverAddress)

	stream, err := laptopClient.Ratelaptop(context.Background())
	require.NoError(t, err)

	scores := []float64{8, 7.5, 10}
	averages := []float64{8, 7.75, 8.5}

	n := len(scores)

	for i := 0; i < n; i++ {
		req := &pb.RateLaptopRequest{
			LapotopId: laptop.GetId(),
			Score:     scores[i],
		}

		err := stream.Send(req)
		require.NoError(t, err)
	}

	err = stream.CloseSend()
	require.NoError(t, err)

	for idx := 0; ; idx++ {
		res, err := stream.Recv()
		if err == io.EOF {
			require.Equal(t, n, idx)
			return
		}
		require.NoError(t, err)
		require.Equal(t, laptop.GetId(), res.GetLapotopId())
		require.Equal(t, uint32(idx+1), res.GetRatedCount())
		require.Equal(t, averages[idx], res.GetAverageScore())
	}
}

func startTestLaptopServer(t *testing.T, laptopStore LaptopStore, imageStore ImageStore, ratingStore RatingStore) string {
	laptopServer := NewLaptopServer(laptopStore, imageStore, ratingStore)
	grpcServer := grpc.NewServer()
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)

	listener, err := net.Listen("tcp", ":0")

	require.NoError(t, err)

	go grpcServer.Serve(listener)

	return listener.Addr().String()
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
