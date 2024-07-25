package util

import (
	"crypto/x509"
	"log"
	"os"
	"strconv"

	"google.golang.org/protobuf/types/known/structpb"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/credentials/oauth"

	"github.com/kiettran199/go-commons/api"
)

type ApiClient struct {
	target      string
	secure      bool
	customToken string
}

func NewApiClient() *ApiClient {
	target := os.Getenv("AIA_ENGINE_API_TARGET")
	customToken := os.Getenv("AIA_ENGINE_API_CUSTOM_TOKEN")
	if target == "" || customToken == "" {
		return nil
	}
	secure := true
	if secureStr, ok := os.LookupEnv("AIA_ENGINE_API_SECURE"); ok {
		boolValue, err := strconv.ParseBool(secureStr)
		if err != nil {
			log.Fatal(err)
		}
		secure = boolValue
	}

	return &ApiClient{target: target, customToken: customToken, secure: secure}
}

func withTransportCredentials(secure bool) grpc.DialOption {
	var creds credentials.TransportCredentials
	if !secure {
		creds = insecure.NewCredentials()
	} else {
		pool, _ := x509.SystemCertPool()
		// error handling omitted
		creds = credentials.NewClientTLSFromCert(pool, "")
	}
	return grpc.WithTransportCredentials(creds)
}

func fetchToken(apiAddress string, customToken string, secure bool) (string, error) {
	conn, err := grpc.Dial(apiAddress, withTransportCredentials(secure))
	if err != nil {
		return "", err
	}
	defer conn.Close()
	userClient := api.NewUserServiceClient(conn)
	request := &api.SignInWithCustomTokenRequest{Token: customToken}
	response, err := userClient.SignInWithCustomToken(context.Background(), request)
	//request := &api.SignInWithEmailRequest{Email:"abc@gmail.com", Password:"Abc@1234"}
	//response, err := userClient.SignInWithEmail(context.Background(), request)
	if err != nil {
		return "", err
	}

	return response.Token, nil
}

// Create connection
func (c *ApiClient) CreateConnection() (*grpc.ClientConn, error) {
	oauthToken, err := fetchToken(c.target, c.customToken, c.secure)
	if err != nil {
		return nil, err
	}
	// Set up the credentials for the connection.
	var perRPC credentials.PerRPCCredentials
	if !c.secure {
		perRPC = NewInsecurePerRPCCredentials(oauthToken)
	} else {
		perRPC = oauth.NewOauthAccess(&oauth2.Token{
			AccessToken: oauthToken,
		})
	}
	return grpc.Dial(c.target, withTransportCredentials(c.secure), grpc.WithPerRPCCredentials(perRPC))
}

func ToStructpb(value interface{}) *structpb.Struct {
	if value == nil {
		return nil
	}
	if returnValue, err := structpb.NewStruct(value.(map[string]interface{})); err != nil {
		println("Unable to init Struct: %v", err)
		return nil
	} else {
		return returnValue
	}
}

func ToStructpbArr(value interface{}) []*structpb.Struct {
	if value == nil {
		return nil
	}
	if arr, ok := value.([]interface{}); ok {
		structpbArr := make([]*structpb.Struct, 0, len(arr))
		for _, el := range arr {
			structpbArr = append(structpbArr, ToStructpb(el))
		}
		return structpbArr
	}
	if arr, ok := value.([]map[string]interface{}); ok {
		structpbArr := make([]*structpb.Struct, 0, len(arr))
		for _, el := range arr {
			structpbArr = append(structpbArr, ToStructpb(el))
		}
		return structpbArr
	}
	println("Cannot convert the data to structpb array")
	return nil
}
