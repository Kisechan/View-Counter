package handler

import (
	"net/http"

	"view-counter/service"
)

type ViewsHandler struct {
	counterService *service.CounterService
}

func NewViewsHandler(counterService *service.CounterService) *ViewsHandler {
	return &ViewsHandler{
		counterService: counterService,
	}
}

func (h *ViewsHandler) HandleViewsRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.counterService.IncrementView(w, r)
	case http.MethodGet:
		h.counterService.GetView(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}