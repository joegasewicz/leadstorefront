package routes

import (
	"encoding/json"
	"fmt"
	"leadstorefront/pkgs"
	"leadstorefront/pkgs/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	identity_client "github.com/joegasewicz/identity-client"
	multipart_requests "github.com/joegasewicz/multipart-requests"
)

type APIClient struct {
	BaseURL string
}

func NewAPIClient() *APIClient {
	return &APIClient{
		BaseURL: fmt.Sprintf("http://%s%s", pkgs.Config.API.Domain, pkgs.Config.API.Addr),
	}
}

func (client *APIClient) URL(path string) string {
	return client.BaseURL + utils.GetVersion(path)
}

func apiURL(path string) string {
	return fmt.Sprintf("http://%s%s%s", pkgs.Config.API.Domain, pkgs.Config.API.Addr, utils.GetVersion(path))
}

func (client *APIClient) Get(c *gin.Context, path string, out any) error {
	identity := identity_client.Identity{URL: client.URL(path)}
	data, err := identity.Get(c.Request)
	if err != nil {
		return err
	}
	return decodeAPIData(data, out)
}

func (client *APIClient) Post(c *gin.Context, path string, payload map[string]interface{}, out any) error {
	identity := identity_client.Identity{URL: client.URL(path)}
	data, err := identity.Post(c.Request, payload)
	if err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	return decodeAPIData(data, out)
}

func (client *APIClient) Put(c *gin.Context, path string, payload map[string]interface{}, out any) error {
	identity := identity_client.Identity{URL: client.URL(path)}
	data, err := identity.Put(c.Request, payload)
	if err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	return decodeAPIData(data, out)
}

func (client *APIClient) Delete(c *gin.Context, path string, out any) error {
	identity := identity_client.Identity{URL: client.URL(path)}
	data, err := identity.Delete(c.Request)
	if err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	return decodeAPIData(data, out)
}

func decodeAPIData(data interface{}, out any) error {
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, out)
}

func (client *APIClient) UploadArticleImage(c *gin.Context, articleID uint) error {
	multipartRequest := multipart_requests.MultipartRequest{
		TempPath: "temp",
		Url:      client.URL(fmt.Sprintf("/admin/articles/%d/main-image", articleID)),
	}
	fileName, file, err := multipartRequest.GetFile(c.Request, "main_image")
	if err != nil {
		if err == http.ErrMissingFile {
			return nil
		}
		return err
	}
	defer file.Close()
	resp, err := multipartRequest.Upload(file, *fileName, "main_image")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("api returned %s", resp.Status)
	}
	return nil
}

func uintPayload(value uint) interface{} {
	if value == 0 {
		return nil
	}
	return value
}

func uintPtrPayload(value *uint) interface{} {
	if value == nil {
		return nil
	}
	return *value
}

func int64PtrPayload(value *int64) interface{} {
	if value == nil {
		return nil
	}
	return *value
}

func timePtrPayload(value interface{}) interface{} {
	return value
}

func apiPathID(id string) (string, bool) {
	parsed, err := strconv.Atoi(id)
	if err != nil || parsed <= 0 {
		return "", false
	}
	return id, true
}
