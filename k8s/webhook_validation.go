package k8s

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/go-hclog"
)

type admissionKind struct {
	Group    string
	Version  string
	Kind     string
	Resource string
}

type userInfo struct {
	Username string
	UID      string
	Extra    map[string][]string
}

type admissionReview struct {
	APIVersion string
	Kind       string
	Request    *admissionReviewRequest
	Response   *admissionReviewResponse
}

type admissionReviewRequest struct {
	UUID        string
	Kind        *admissionKind
	Resource    *admissionKind
	SubResource string

	Name      string
	Namespace string
	operation string

	UserInfo *userInfo

	Object    *Item
	OldObject *Item

	DryRun bool
}

type status struct {
	Code    uint64
	Message string
}

type admissionReviewResponse struct {
	UID      string
	Allowed  bool
	Status   *status
	Warnings []string
}

type Validator interface {
	Validate(req *admissionReviewRequest) (*admissionReviewResponse, error)
}

type webhookValidation struct {
	logger    hclog.Logger
	validator Validator
}

func (h *webhookValidation) startHTTPServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", h.Handle)

	go http.ListenAndServe(":4578", mux)
	h.logger.Info("webhook server started", "port", "4578")
}

func (h *webhookValidation) Handle(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("request received", "method", r.Method)

	handleErr := func(code int, msg string) {
		h.logger.Warn("error on admission request", "msg", msg)
		http.Error(w, msg, code)
	}

	if ct := r.Header.Get("Content-Type"); ct != "application/json" {
		handleErr(http.StatusBadRequest, fmt.Sprintf("Invalid content-type: %q", ct))
		return
	}

	var body []byte
	if r.Body != nil {
		var err error
		if body, err = ioutil.ReadAll(r.Body); err != nil {
			handleErr(http.StatusBadRequest, fmt.Sprintf("error reading body: %v", err))
			return
		}
	}
	fmt.Println(string(body))
	if len(body) == 0 {
		handleErr(http.StatusBadRequest, fmt.Sprintf("Body is empty"))
		return
	}

	var obj *admissionReview
	if err := json.Unmarshal(body, &obj); err != nil {
		handleErr(http.StatusBadRequest, fmt.Sprintf("failed to decode json: %v", err))
		return
	}

	admResp, err := h.validator.Validate(obj.Request)
	if err != nil {
		handleErr(http.StatusInternalServerError, fmt.Sprintf("failed to validate: %v", err))
		return
	}

	obj.Request = nil
	obj.Response = admResp

	resp, err := json.Marshal(obj)
	if err != nil {
		handleErr(http.StatusInternalServerError, fmt.Sprintf("failed to encode admission response: %v", err))
		return
	}

	if _, err := w.Write(resp); err != nil {
		h.logger.Error("failed to write", "err", err)
	}
}
