package health

import (
	"net/http"
)

func (h *Health) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusOK)
		if h.isReady.Load() {
			_, _ = w.Write([]byte(`{"ok": true}`)) // nolint:errcheck

			return
		}

		_, _ = w.Write([]byte(`{"ok": false}`)) // nolint:errcheck
	}
}
