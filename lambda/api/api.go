package api

import (
	"encoding/json"
	"fmt"
	"lambda/database"
	"lambda/types"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

type ApiHandler struct {
	dbStore database.UserStore
}

func NewApiHandler(dbStore database.UserStore) ApiHandler {
	return ApiHandler{
		dbStore: dbStore,
	}
}

func (api ApiHandler) RegisterUserHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	var registerUser types.RegisterUser

	err := json.Unmarshal([]byte(request.Body), &registerUser)

	if err != nil {
		return events.APIGatewayProxyResponse{
			Body: "Invalid Request",
			StatusCode: http.StatusBadRequest,
		}, err
	}

	if registerUser.Username == "" || registerUser.Password == "" {
		return events.APIGatewayProxyResponse{
			Body: "Invalid Request - fileds empty",
			StatusCode: http.StatusBadRequest,
		}, err
	}

	// does this user already exist
	userExists , err := api.dbStore.DoesUserExist(registerUser.Username);

	if err != nil {
		return events.APIGatewayProxyResponse{
			Body: "Internal Server Error",
			StatusCode: http.StatusInternalServerError,
		}, err
	}

	if userExists {
		return events.APIGatewayProxyResponse{
			Body: "User already exist",
			StatusCode: http.StatusConflict,
		}, err
	}

	// new user logic
	user, err := types.NewUser(registerUser)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body: "Internal Server Error",
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("could not create new user %w",err)
	}


	// we know that a user does not exist , here we will insert the user
	err = api.dbStore.InsertUser(user)

	if err != nil {
		return events.APIGatewayProxyResponse{
			Body: "Internal Server Error",
			StatusCode: http.StatusInternalServerError,
		}, err
	}

	return events.APIGatewayProxyResponse{
		Body: "Successfully refistered user",
		StatusCode: http.StatusOK,
	}, nil

}

// login user handler here 
func (api ApiHandler) Loginuser(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse,error) {
	type LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var loginRequest LoginRequest

	err := json.Unmarshal([]byte(request.Body), &loginRequest)

	if err != nil {
		return events.APIGatewayProxyResponse{
			Body: "Invalid Request",
			StatusCode: http.StatusBadRequest,
		}, err
	}

	user , err := api.dbStore.GetUser(loginRequest.Username)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body: "Internal Server Error",
			StatusCode: http.StatusInternalServerError,
		}, err
	}

	if !types.ValidatePassword(user.Passwordhash,loginRequest.Password) {
		return events.APIGatewayProxyResponse{
			Body: "Invalid User Credentials",
			StatusCode: http.StatusBadRequest,
		}, err
	}

	accessToken := types.CreateToken(user)

	successMsg := fmt.Sprintf(`{"access_token: "%s"}`,accessToken)

	return events.APIGatewayProxyResponse{
			Body: successMsg,
			StatusCode: http.StatusOK,
		}, nil
}