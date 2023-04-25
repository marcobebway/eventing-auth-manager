package oidc_test

import (
	"bytes"
	"context"
	"errors"
	"github.com/kyma-project/eventing-auth-manager/internal/ias/internal/oidc"
	"github.com/stretchr/testify/require"
	"io"
	"k8s.io/client-go/rest/fake"
	"k8s.io/utils/pointer"
	"net/http"
	"testing"
)

const oidcConfigMock = `{"token_endpoint":"https://domain-url.com/token"}`

func Test_oidcClient_getTokenUrl(t *testing.T) {
	type fields struct {
		httpClient *http.Client
	}
	tests := []struct {
		name    string
		fields  fields
		want    *string
		wantErr error
	}{
		{
			name: "should return token endpoint",
			fields: fields{
				httpClient: mockHttpClientResponseOk([]byte(oidcConfigMock)),
			},
			want: pointer.String("https://domain-url.com/token"),
		},
		{
			name: "should return token URL by using base url from client and well-known OIDC config path for request",
			fields: fields{
				httpClient: fake.CreateHTTPClient(func(request *http.Request) (*http.Response, error) {
					require.Equal(t, http.MethodGet, request.Method)
					require.Equal(t, "https", request.URL.Scheme)
					require.Equal(t, "domain-url.com", request.URL.Host)
					require.Equal(t, "/.well-known/openid-configuration", request.URL.Path)
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader([]byte(oidcConfigMock))),
					}, nil
				}),
			},
			want: pointer.String("https://domain-url.com/token"),
		},
		{
			name: "should return nil when well known contains no token endpoint",
			fields: fields{
				httpClient: mockHttpClientResponseOk([]byte("{}")),
			},
		},
		{
			name: "should return error when response status code is not 200",
			fields: fields{
				httpClient: fake.CreateHTTPClient(func(request *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusInternalServerError,
					}, nil
				}),
			},
			wantErr: errors.New("unexpected status code 500"),
		},
		{
			name: "should return error when response body is nil",
			fields: fields{
				httpClient: mockHttpClientResponseOk(nil),
			},
			wantErr: errors.New("unexpected end of JSON input"),
		},
		{
			name: "should return error when response body is no json",
			fields: fields{
				httpClient: mockHttpClientResponseOk([]byte("invalid json")),
			},
			wantErr: errors.New("invalid character 'i' looking for beginning of value"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			c := oidc.NewOidcClient(tt.fields.httpClient, "https://domain-url.com")

			// when
			got, err := c.GetTokenEndpoint(context.TODO())

			// then
			if tt.wantErr != nil {
				require.Error(t, err)
				require.EqualError(t, tt.wantErr, err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func mockHttpClientResponseOk(body []byte) *http.Client {
	return fake.CreateHTTPClient(func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}, nil
	})
}
