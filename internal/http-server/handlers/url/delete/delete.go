package delete

import (
	"Url-shortener-go/internal/lib/api/response"
	"Url-shortener-go/internal/storage"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

type URLDeleter interface {
	DeleteURL(alias string) (int64, error)
}

func New(log *slog.Logger, urlDeleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const f = "handlers.url.delete.New"

		log = log.With(
			slog.String("function", f),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")

		if alias == "" {
			log.Info("missing url alias")
			render.JSON(w, r, response.Error("missing url alias"))
			return
		}

		id, err := urlDeleter.DeleteURL(alias)
		if errors.Is(err, storage.ErrUrlNotFound) {
			log.Info("url not found", "alias", alias)
			render.JSON(w, r, response.Error("url not found"))
			return
		}
		if err != nil {
			log.Error("error deleting url", "alias", alias, "error", err)
			render.JSON(w, r, response.Error("internal server error"))
			return
		}

		log.Info("url deleted", "alias", alias, "id", id)
		render.JSON(w, r, response.OK())
	}
}
