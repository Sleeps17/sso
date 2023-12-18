package tests

import (
	"fmt"
	ssov1 "github.com/Sleeps17/protos/gen/go/sso"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sso/tests/suite"
	"testing"
	"time"
)

const (
	emptyAppID     = 0
	appID          = 1
	undefinedAppId = -1
	appSecret      = "test-secret"

	passDefaultLen = 10

	deltaTime = 60
)

func TestRegisterLogin_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	password := randomFakePassword()

	registerResp, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{Password: password, Email: email})
	require.NoError(t, err)
	assert.NotEmpty(t, registerResp.GetUserId())

	loginResp, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Email:    email,
		Password: password,
		AppId:    appID,
	})
	require.NoError(t, err)

	loginTime := time.Now()

	token := loginResp.GetToken()
	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	assert.Equal(t, registerResp.GetUserId(), int64(claims["uid"].(float64)))
	assert.Equal(t, email, claims["email"].(string))
	assert.Equal(t, appID, int(claims["app_id"].(float64)))

	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), int64(claims["exp"].(float64)), deltaTime)
}

func TestRegisterLogin_DuplicateRegistration(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	password := randomFakePassword()

	registerResp, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, registerResp.GetUserId())

	registerResp2, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})

	require.Error(t, err)
	assert.Empty(t, registerResp2.GetUserId())
	assert.ErrorContains(t, err, "user already exists")
}

func TestRegisterLogin_UndefinedApp(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	password := randomFakePassword()

	_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)

	loginResp, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Email:    email,
		Password: password,
		AppId:    undefinedAppId,
	})

	require.Error(t, err)
	assert.Empty(t, loginResp.GetToken())
	assert.ErrorContains(t, err, "invalid app id")
}

func TestRegister_FailCases(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name        string
		email       string
		password    string
		expectedErr string
	}{
		{
			name:        "Register with Empty Password",
			email:       gofakeit.Email(),
			password:    "",
			expectedErr: "len password required will be grate then 8",
		},
		{
			name:        "Register with Empty Email",
			email:       "",
			password:    randomFakePassword(),
			expectedErr: "email is not valid",
		},
		{
			name:        "Register with Both Empty",
			email:       "",
			password:    "",
			expectedErr: "email is not valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				Email:    tt.email,
				Password: tt.password,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)

		})
	}
}

func TestLogin_FailCases(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name        string
		email       string
		password    string
		appID       int32
		expectedErr string
	}{
		{
			name:        "Login with Empty Password",
			email:       gofakeit.Email(),
			password:    "",
			appID:       appID,
			expectedErr: "len password required will be grate then 8",
		},
		{
			name:        "Login with Empty Email",
			email:       "",
			password:    randomFakePassword(),
			appID:       appID,
			expectedErr: "email is not valid",
		},
		{
			name:        "Login with Both Empty Email and Password",
			email:       "",
			password:    "",
			appID:       appID,
			expectedErr: "email is not valid",
		},
		{
			name:        "Login with Non-Matching Password",
			email:       gofakeit.Email(),
			password:    randomFakePassword(),
			appID:       appID,
			expectedErr: "invalid email or password",
		},
		{
			name:        "Login without AppID",
			email:       gofakeit.Email(),
			password:    randomFakePassword(),
			appID:       emptyAppID,
			expectedErr: "invalid app id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				Email:    gofakeit.Email(),
				Password: randomFakePassword(),
			})
			require.NoError(t, err)

			_, err = st.AuthClient.Login(ctx, &ssov1.LoginRequest{
				Email:    tt.email,
				Password: tt.password,
				AppId:    tt.appID,
			})
			fmt.Println(tt.expectedErr)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func randomFakePassword() string {
	return gofakeit.Password(
		true,
		true,
		true,
		true,
		false,
		passDefaultLen,
	)
}
