package rest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVolumesExists(t *testing.T) {
	existingVolume := "vol1"
	nonExistingVolume := "vol2"

	endPoint := "/api/1.0/volumes"

	testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		switch request.URL.RequestURI() {
		case endPoint:
			response := response{Ok: true, Err: ""}
			volumesList := []volume{volume{Name: existingVolume}}
			volumeResponse := volumeResponse{Data: volumesList, response: response}
			jsonData, _ := json.Marshal(volumeResponse)
			fmt.Println(jsonData)
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write(jsonData)

		}
	}))
	defer testServer.Close()

	glusterClient := NewClient(testServer.URL, endPoint)

	//Check for valid volume
	volumeExists, _ := glusterClient.VolumeExist(existingVolume)
	if !volumeExists {
		t.Fatalf("%s volume does not exist", existingVolume)
	}

	//check for invalid volume
	volumeExists, _ = glusterClient.VolumeExist(nonExistingVolume)
	if volumeExists {
		t.Fatalf("%s volume should exist", nonExistingVolume)
	}
}

func TestCreateVolume(t *testing.T) {

	volumeTobeCreated := "vol1"
	endPoint := "/api/1.0/volume/"
	endPointWithVolume := fmt.Sprintf("%s%s", endPoint, volumeTobeCreated)

	testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		switch request.URL.RequestURI() {
		case endPointWithVolume:
			var createResponse response
			if request.Method == http.MethodPost {
				body, err := ioutil.ReadAll(request.Body)
				if err != nil {
					createResponse = response{Ok: false, Err: "Not able to ready body"}
					break
				}
				if len(body) < 0 {
					createResponse = response{Ok: false, Err: "body of the request is empty"}
					break
				}
				createResponse = response{Ok: true, Err: ""}
			} else {
				createResponse = response{Ok: false, Err: "unknown HTTP method"}
			}
			jsonData, _ := json.Marshal(createResponse)
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write(jsonData)
		}
	}))
	defer testServer.Close()

	glusterClient := NewClient(testServer.URL, endPointWithVolume)
	peers := []string{testServer.URL}
	if createVolErr := glusterClient.CreateVolume(volumeTobeCreated, peers); createVolErr != nil {
		t.Fatalf("Volume creation failed with following error %s", createVolErr)
	}
}

func TestStopVolume(t *testing.T) {
	VolumeToBeStopped := "vol1"
	endPointWithVolume := fmt.Sprintf("/api/1.0/volume/%s/stop", VolumeToBeStopped)

	testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		switch request.URL.RequestURI() {
		case endPointWithVolume:
			var createResponse response
			if request.Method == http.MethodPut {
				createResponse = response{Ok: true, Err: ""}
			} else {
				createResponse = response{Ok: true, Err: "unknow HTTP method"}
			}
			jsonData, _ := json.Marshal(createResponse)
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write(jsonData)
		}
	}))
	defer testServer.Close()

	glusterClient := NewClient(testServer.URL, endPointWithVolume)
	if createVolErr := glusterClient.StopVolume(VolumeToBeStopped); createVolErr != nil {
		t.Fatalf("Volume creation failed with following error %s", createVolErr)
	}
}
